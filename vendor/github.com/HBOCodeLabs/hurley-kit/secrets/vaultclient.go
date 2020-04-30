package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/HBOCodeLabs/hurley-kit/llog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/logical"
)

// vaultSvcValue atomically stores and retrieves the value of
// Vault service client
var vaultSvcValue atomic.Value

// vaultSvcCreationMutex ensures that only 1 instance of
// vaultSvc is created
var vaultSvcCreationMutex sync.Mutex

// vaultObjectGetter provides an API to return Vault object when
// called by key
type vaultObjectGetter interface {
	getObject(opts vaultStoreOpts, jwtLocation string, key string) ([]byte, error)
}

type jwtPayload struct {
	JWT     string `json:"jwt"`
	AppRole string `json:"role"`
}

// getVaultService will return a pointer to a singleton of Vault service.
// If no vaultAddress is provided then this library will attempt to find
// that vaultAddress in the environment variable VAULT_ADDR which is a Vault
// convention.
func getVaultService(opts vaultStoreOpts, jwtLocation string) (*api.Client, error) {
	if vaultSvcStored := vaultSvcValue.Load(); vaultSvcStored != nil {
		return vaultSvcStored.(*api.Client), nil
	}

	vaultSvcCreationMutex.Lock()
	defer vaultSvcCreationMutex.Unlock()

	// Diagnostics info for httpClient
	llog.Debug("Timeout", opts.httpClient.Timeout)

	var vaultSvc *api.Client
	var kubernetesJWT, clientToken string
	var err error

	if vaultSvcStored := vaultSvcValue.Load(); vaultSvcStored == nil {
		vaultSvc, err = api.NewClient(&api.Config{
			HttpClient: opts.httpClient,
		})
		if err != nil {
			return nil, err
		}

		llog.Debug("message", "Using Kubernetes JWT Auth flow to retrieve Vault token")
		kubernetesJWT, err = getJWTFromFile(jwtLocation)
		if err != nil {
			llog.Warn("message", "A Kubernetes service account JWT file was not located at "+jwtLocation)
		} else {
			clientToken, err = getTokenFromJWT(opts, kubernetesJWT)
			if err != nil {
				return nil, err
			}
		}

		// Override clientToken with VAULT__TOKEN if set
		if len(os.Getenv("VAULT_TOKEN")) > 0 {
			llog.Debug("message", "Setting Vault client token from VAULT_TOKEN environment variable")
			clientToken = os.Getenv("VAULT_TOKEN")
		}

		// At this point error out if we have not established a client token
		if len(clientToken) == 0 {
			llog.Error("message", "A client token could not be found.  If in development mode try setting the VAULT_TOKEN environment variable")
			return nil, err
		}

		// If no Vault address is provided then try to get it from
		// environment
		if opts.vaultAddress == "" {
			opts.vaultAddress = os.Getenv("VAULT_ADDR")
		}

		// If Vault address is still empty then error out
		if opts.vaultAddress == "" {
			return nil, errors.New("You must provide a Vault cluster address")
		}

		err = vaultSvc.SetAddress(opts.vaultAddress)
		if err != nil {
			return nil, err
		}

		if clientToken == "" {
			return nil, errors.New("Client token returned empty")
		}

		vaultSvc.SetToken(clientToken)
		vaultSvc.SetClientTimeout(opts.vaultTimeout)
		vaultSvc.SetMaxRetries(opts.vaultMaxRetries)

		vaultSvcValue.Store(vaultSvc)
	} else {
		vaultSvc = vaultSvcStored.(*api.Client)
	}

	return vaultSvc, nil
}

// vaultClient provides an API to return Vault object when
// called by key.
type vaultClient struct{}

func (v *vaultClient) getObject(opts vaultStoreOpts, jwtLocation string, key string) ([]byte, error) {
	vaultSvc, err := getVaultService(opts, jwtLocation)
	if err != nil {
		return nil, err
	}

	if vaultSvc == nil {
		return nil, errors.New("Unable to initialize a vaultService")
	}

	r, err := vaultSvc.Logical().Read(key)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, errors.New("Secret path does not exist")
	}

	if r.Data == nil {
		return nil, errors.New("Vault response does not contain data")
	}

	return json.Marshal(r.Data)
}

func getJWTFromFile(jwtLocation string) (jwt string, err error) {
	jwtTokenBytes, err := ioutil.ReadFile(jwtLocation)
	if err != nil {
		return "", err
	}

	// Trim newline
	jwtToken := strings.TrimSuffix(string(jwtTokenBytes), "\n")

	return jwtToken, nil
}

func getTokenFromJWT(opts vaultStoreOpts, jwt string) (token string, err error) {
	// Get token from JWT auth method
	jwtRequestToken := jwtPayload{JWT: jwt, AppRole: opts.appRole}

	payload, err := json.Marshal(jwtRequestToken)
	if err != nil {
		return "", err
	}

	jwtBody := strings.NewReader(string(payload))
	// Parse Vault address to properly join
	u, err := url.Parse(opts.vaultAddress)
	if err != nil {
		return "", err
	}

	kubernetesAuthPath := fmt.Sprintf("/v1/auth/%s/login", opts.kubernetesAuthClusterID)
	llog.Debug("Kubernetes Auth Path", kubernetesAuthPath)

	u.Path = path.Join(u.Path, kubernetesAuthPath)
	req, err := http.NewRequest("POST", u.String(), jwtBody)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "applicaiton/x-www-form-urlencoded")

	resp, err := opts.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Unmarshall response into Vault HTTPResponse struct
	var vaultResponse logical.HTTPResponse

	err = json.Unmarshal(responseBody, &vaultResponse)
	if err != nil {
		return "", err
	}

	if vaultResponse.Auth == nil {
		return "", errors.New("Error trying to authenticate with provided app role and JWT")
	}

	return vaultResponse.Auth.ClientToken, nil
}
