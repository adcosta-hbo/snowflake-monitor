package sql_exporter

import (
	"time"

//	"github.com/HBOCodeLabs/hurley-kit/secrets"
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
func FetchSecrets() (vaultKey string, err error) {

	var opts Options

	opts.Secrets.Endpoint = "https://vault.api.hbo.com"
	// var K8sAuthCluster string = `jenkins`
	// var AppRole string = `staging-tests`
	// var CacheTimeoutInSeconds int = 10
	// var TimeoutInSeconds int = 10
	// var MaxRetries int = 10

	// var AuthTokenSecretPath string = `secret/hurley/snowflake/staging/hurley-staging`
	// var TokensDBPasswordSecretPath string = `secret/hurley/snowflake/staging/hurley-staging`
	// RMDBUsernameSecretPath     string `json:"rmDBUsernameSecretPath"`
	// RMDBPasswordSecretPath     string `json:"rmDBPasswordSecretPath"`

	//cfg := opts.Secrets
<<<<<<< HEAD
//	vaultAddress := secrets.VaultAddress(cfg.Endpoint)
=======
	//vaultAddress := secrets.VaultAddress(cfg.Endpoint)
>>>>>>> 2f68b11ba8fa186d79e8a31c5b4cd32245ad5f99
	// kubeCluster := secrets.KubernetesAuthClusterID(K8sAuthCluster)
	// appRole := secrets.AppRole(AppRole)
	// //note cacheTTL is irrelevant at this time since we only fetch each secret once on startup
	// cacheTTL := secrets.CacheTTL(time.Duration(CacheTimeoutInSeconds) * time.Second)
	// vaultTimeout := secrets.VaultTimeout(time.Duration(TimeoutInSeconds) * time.Second)
	// vaultMaxRetries := secrets.VaultMaxRetries(MaxRetries)

	// store, err := secrets.NewVaultStore(vaultAddress, kubeCluster, appRole, cacheTTL, vaultTimeout, vaultMaxRetries)
	// if err != nil {
	// 	return
	// }

	// Snowflake password
	// password, err = fetchSecret(store, vaultAddress.TokensDBPasswordSecretPath)
	// if err != nil {
	// 	return
	// }

	//store, err := secrets.NewVaultStore()

	return opts.Secrets.Endpoint, nil
}

// fetchSecret
// func fetchSecret(store *secrets.VaultStore, secretPath string) (secretValue string, err error) {
// 	secretBytes, err := store.Get(secretPath)
// 	if err != nil {
// 		return
// 	}
// 	secretValue = string(secretBytes)
// 	log.Infof(fmt.Sprintf("Secret key: %s value: %s", secretPath, secretValue))
// 	return
// }
