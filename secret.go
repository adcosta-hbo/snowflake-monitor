package sql_exporter

import (
	"encoding/json"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/secrets"
)

// VaultConfig comment
type VaultConfig struct {
	Endpoint              string `json:"endpoint"`
	K8sAuthCluster        string `json:"k8SAuthCluster"`
	AppRole               string `json:"appRole"`
	CacheTimeoutInSeconds int    `json:"cacheTimeoutInSeconds"`
	TimeoutInSeconds      int    `json:"timeoutInSeconds"`
	MaxRetries            int    `json:"maxRetries"`

	AuthTokenSecretPath        string `json:"authTokenSecretPath"`
	TokensDBPasswordSecretPath string `json:"tokensDBPasswordSecretPath"`
	RMDBUsernameSecretPath     string `json:"rmDBUsernameSecretPath"`
	RMDBPasswordSecretPath     string `json:"rmDBPasswordSecretPath"`
}

// Options comment
type Options struct {
	Port        int
	Environment string
	Region      string

	// statsd metrics collection
	Metrics struct {
		Host     string
		Port     int
		Buffered bool
		Interval int
	} `json:"metrics"`

	// options for using the secret store
	Secrets VaultConfig `json:"secrets"`

	// Time until tokens are expired. Configured in minutes and parsed to a time.Duration.
	Expiry time.Duration `json:"expiryMinutes"`

	// ProjectRoot is an absolute path to the directory containing the project.
	// Any relative directory paths are created based on this dir. If not set,
	// this value will use the working directory returned by the os.Getwd
	// function. It can be overriden at runtime using the PROJECT_ROOT env var.
	ProjectRoot string `json:"projectRoot"`

	// JSONSchemasDir is the relative path to the directory containing the JSON
	// schema definitions to use for request/response validation
	JSONSchemasDir string `json:"jsonSchemasDir"`

	// DownloadRulesFile is the file containing download eligibility rules
	DownloadRulesFile string `json:"downloadRulesFile"`

	// ExceptionAccountsFile is the file containing the accounts that use special rules
	ExceptionAccountsFile string `json:"exceptionAccountsFile"`

	// Hurley Auth Token secret
	AuthTokenSecret string
}

// FetchSecrets Use Vault configuration to acquire secrets according to their path, then populate the secret values in Options.
func FetchSecrets() (vaultKey []byte, err error) {

	var opts Options

	opts.Secrets.Endpoint = "https://vault.api.hbo.com"
	opts.Secrets.K8sAuthCluster = `jenkins`
	opts.Secrets.AppRole = `staging-tests`
	opts.Secrets.CacheTimeoutInSeconds = 10
	opts.Secrets.TimeoutInSeconds = 10
	opts.Secrets.MaxRetries = 10
	opts.Secrets.AuthTokenSecretPath = "secret/hurley/auth/*token"
	opts.Secrets.TokensDBPasswordSecretPath = "secret/hurley/snowflake/staging/hurley-staging"
	// var TokensDBPasswordSecretPath string = `secret/hurley/snowflake/staging/hurley-staging`
	// RMDBUsernameSecretPath     string `json:"rmDBUsernameSecretPath"`
	// RMDBPasswordSecretPath     string `json:"rmDBPasswordSecretPath"`

	cfg := opts.Secrets
	vaultAddress := secrets.VaultAddress(cfg.Endpoint)
	kubeCluster := secrets.KubernetesAuthClusterID(cfg.K8sAuthCluster)
	appRole := secrets.AppRole(cfg.AppRole)
	//note cacheTTL is irrelevant at this time since we only fetch each secret once on startup
	cacheTTL := secrets.CacheTTL(time.Duration(cfg.CacheTimeoutInSeconds) * time.Second)
	vaultTimeout := secrets.VaultTimeout(time.Duration(cfg.TimeoutInSeconds) * time.Second)
	vaultMaxRetries := secrets.VaultMaxRetries(cfg.MaxRetries)

	store, err := secrets.NewVaultStore(vaultAddress, kubeCluster, appRole, cacheTTL, vaultTimeout, vaultMaxRetries)

	if err != nil {
		return
	}

	// Snowflake password
	password, err := fetchSecret(store, opts.Secrets.TokensDBPasswordSecretPath)
	if err != nil {
		return
	}
	return password, nil
}

//fetchSecret
func fetchSecret(store *secrets.VaultStore, secretPath string) (secretValue []byte, err error) {
	secretBytes, err := store.Get(secretPath)
	if err != nil {
		return
	}
	secretValue = secretBytes
	return
}

// fetchAuthTokenSecret Auth token is stored as a single-valued json array requiring special parsing
func fetchAuthTokenSecret(store *secrets.VaultStore, secretPath string) (authTokenValue string, err error) {
	secretBytes, err := store.Get(secretPath)
	if err != nil {
		return
	}

	var keys []string
	err = json.Unmarshal(secretBytes, &keys)
	if err != nil {
		return
	}
	authTokenValue = keys[0]

	return
}
