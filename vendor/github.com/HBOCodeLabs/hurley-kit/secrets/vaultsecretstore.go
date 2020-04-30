package secrets

import (
	"errors"
	"net/http"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/llog"
)

const kubernetesJWTLocation = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// vaultStoreOpts encapsulates all the various configuration options that can be
// modified by a user of a Vault secret store. Instances are only modified by
// the various ConfigurationFuncs, and used by the NewVaultStore "constructor" to
// create the actual VaultStore.
type vaultStoreOpts struct {
	ttl                     time.Duration
	vaultAddress            string
	vaultTimeout            time.Duration
	vaultMaxRetries         int
	appRole                 string
	kubernetesAuthClusterID string
	httpClient              *http.Client
}

// A ConfigurationFunc is a typedef for the return values of the various
// functions that configure some aspect of a new Vault secret store.
type ConfigurationFunc func(*vaultStoreOpts) error

// CacheTTL returns a configuration function that sets up the
// secret store cache time to live.
func CacheTTL(ttl time.Duration) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		if ttl <= zeroDuration {
			return errors.New("ttl must be greater than 0")
		}
		opts.ttl = ttl
		return nil
	}
}

// VaultAddress returns a configuration function that sets up the
// Vault service endpoint to use to retrieve secrets.
func VaultAddress(vaultAddress string) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		opts.vaultAddress = vaultAddress
		return nil
	}
}

// VaultTimeout returns a configuration function that sets up the
// request timeout to use for each request made by the vault client.
func VaultTimeout(vaultTimeout time.Duration) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		opts.vaultTimeout = vaultTimeout
		return nil
	}
}

// VaultMaxRetries returns a configuration function that sets up the
// vault max retry value for the vault client.
func VaultMaxRetries(vaultMaxRetries int) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		opts.vaultMaxRetries = vaultMaxRetries
		return nil
	}
}

// AppRole returns a configuration function that sets up the
// application role to be used for Vault secret authorization.
func AppRole(appRole string) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		if appRole == "" {
			return errors.New("must specify appRole")
		}
		opts.appRole = appRole
		return nil
	}
}

// KubernetesAuthClusterID returns a configuration function that
// sets up the Kubernetes cluster ID to use within the auth API
// call which verifies the Kubernetes Service Account JWT.
func KubernetesAuthClusterID(kubernetesAuthClusterID string) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		if kubernetesAuthClusterID == "" {
			return errors.New("must specify kubernetesAuthClusterID")
		}
		opts.kubernetesAuthClusterID = kubernetesAuthClusterID
		return nil
	}
}

// HTTPClient returns a configuration function that sets up the
// http client that will be used for all vault related requests.
func HTTPClient(httpClient *http.Client) ConfigurationFunc {
	return func(opts *vaultStoreOpts) error {
		opts.httpClient = httpClient
		return nil
	}
}

// NewVaultStore is a constructor function that creates a new VaultStore.
// It accepts 0 or more configuration functions, to customize the behavior
// of the returned store.
//
// If no configuration functions are used, the store defaults to using
// a 10 second ttl for the secrets cache, a 3 second client timeout
// and a max retry value of 5.
func NewVaultStore(configFuncs ...ConfigurationFunc) (*VaultStore, error) {
	// start with the defaults
	opts := &vaultStoreOpts{
		ttl:                     10 * time.Second,
		vaultAddress:            "",
		vaultTimeout:            3 * time.Second,
		vaultMaxRetries:         5,
		appRole:                 "",
		kubernetesAuthClusterID: "kubernetes",
		httpClient:              http.DefaultClient,
	}

	// apply each configuration function that was supplied
	for _, configFunc := range configFuncs {
		if err := configFunc(opts); err != nil {
			return nil, err
		}
	}

	return &VaultStore{
		cacheEntries:      make(map[string]cacheEntry),
		opts:              opts,
		jwtLocation:       kubernetesJWTLocation,
		vaultObjectGetter: &vaultClient{},
		utcNowGetter:      &clock{},
	}, nil
}

// VaultStore is the struct used to access hurley secrets
type VaultStore struct {
	cacheEntries      map[string]cacheEntry
	opts              *vaultStoreOpts
	jwtLocation       string
	vaultObjectGetter vaultObjectGetter
	utcNowGetter      utcNowGetter
}

// Get returns the secret by key
func (v *VaultStore) Get(key string) ([]byte, error) {
	entry, ok := v.cacheEntries[key]
	llog.Debug("event", "resultFromMap", "key", key, "ok", ok)

	utcNow := v.utcNowGetter.getUTCNow()
	if ok && entry.expires.After(*utcNow) {
		llog.Debug("event", "returnFromCache", "key", key)
		return entry.byts, nil
	}

	byts, err := v.refresh(key)
	if err != nil {
		llog.Error("event", "failedToGetSecret", "err", err)
		return nil, err
	}

	return byts, nil
}

// refresh will reload the secret from Vault via vaultClient and cache
// a successful result in memory
func (v *VaultStore) refresh(key string) ([]byte, error) {
	byts, err := v.vaultObjectGetter.getObject(*v.opts, v.jwtLocation, key)
	if err != nil {
		return nil, err
	}

	utcNow := v.utcNowGetter.getUTCNow()
	expires := (*utcNow).Add(v.opts.ttl)
	llog.Debug("event", "storingCacheEntry", "key", key, "expires", expires)
	v.cacheEntries[key] = cacheEntry{
		expires: expires,
		byts:    byts,
	}

	return byts, nil
}
