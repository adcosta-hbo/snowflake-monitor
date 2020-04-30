# secrets

Golang version of Hurley-Secret: a central location for storing secrets (keys, certificates, etc.). Secrets should be retrieved at runtime, and not be included as part of any kind of a build process.

### Overview

Hurley services keep secrets in either a protected S3 bucket or in a Hashicorp
Vault cluster. This library will fetch the secret by key and cache the secret
in memory.

### Usage example
```
    // S3
    // Create a store instance
    //  - ttl indicates how long secret store should cache the secret in memory
    //  - bucket is name of s3 bucket where secrets are held
    //  - region is the AWS region where secrets are held
    store, err := NewStore(time.Duration(10)*time.Second, "hurley-keys", "us-east-1")

    // Check error to see whether store creation is successful
    if err != nil {
        t.Errorf("Failed to create secret store. Error %s", err)
        t.Fail()
    }

    // Get secret bytes from the store, by passing in the key. The key will be the
    // path of secret file inside the secret bucket.
    byts, err := store.Get("production/*token.json")

    // Vault
    // Create a Vault store instance
    // Use configuration functions to override default values for
    // cache entry TTL, vault timeout and vault max retries.
    ttl := secrets.CacheTTL(30 * time.Second)
    vaultAddress := secrets.VaultAddress("https://vault-dev-default-us-east-1.development.hurley.hbo.com")
    vaultTimeout := secrets.VaultTimeout(2 * time.Second)
    vaultMaxRetries := secrets.VaultMaxRetries(4)
    appRole := secrets.AppRole("beta-demo")
    // Try out a custom http client
    var customHttp= &http.Client{
        Timeout: time.Second * 10,
    }
    httpClient := secrets.HTTPClient(customHttp)

    vaultStore, err := NewVaultStore(ttl, vaultAddress, vaultTimeout, vaultMaxRetries, appRole, httpClient)

    // Check error to see whether store creation is successful
    if err != nil {
        t.Errorf("Failed to create secret store. Error %s", err)
        t.Fail()
    }

    // Get secret bytes from the store, by passing in the key. The key will be the
    // path of secret file inside the secret bucket.  The returned bytes are marshaled
    // from key-value pairs under the secret path.
    byts, err := vaultStore.Get("production/token")

    // Unmarshal the byts 
    var secretMap map[string]interface{}
    err := json.Unmarshal(byts, &secretMap)

    // Get the secret value with attribute names
    secret1 := secretMap["s1"].(string)
    secret2 := secretMap["s2"].(string)

    // Unmarshal secrets map to a struct
    byts, err := vaultStore.Get("production/dbsecret")
    type Credentials struct{
        ROUsername string `json:"ro_username"`
        ROPassword string `json:"ro_password"`
        RWUsername string `json:"rw_username"`
        RWPassword string `json:"rw_password"`
    }
    var creds Credentials
    err := json.Unmarshal(byts, &creds)

```

### Details

#### AWS Session
Because session creation is expensive, this library will create and use a singleton of the session object.
[AWS sdk doc](https://docs.aws.amazon.com/sdk-for-go/api/aws/session/)

#### Vault configuration
There are two ways you can pass Vault service endpoint url's to the secrets vault store.  The first is to
simply pass along the url through the `NewVauleStore` function.  The other way is to set the Vault service
endpoint in the `VAULT_ADDR` environment variable.

Aside from setting a Vault `AppRole` config setting you might also need to set the `KubernetesAuthClusterID`
to let the hurley-secrets Vault client know how to contact your particular Vault cluster for auth.

#### Vault authentication setup
For more information regarding how to configure a Hurley service running inside Kubernetes to
authenticate with Vault please see the
[Vault Kubernetes Authentication Guide](https://github.com/HBOCodeLabs/SRE-Vault/blob/master/docs/guides/vault-kubernetes-authentication.md)

For local testing and CI you can set the `VAULT_TOKEN` environment variable and the library will use that token to authenticate with Vault.
Note: Vault tokens used for this purpose should never have access to production secrets.

#### Dependencies

##### [llog](http://github.com/HBOCodeLabs/hurley-kit/llog)
This is the leveled-logging library Hurley projects use.

##### [aws-sdk-go](http://github.com/aws/aws-sdk-go)
This is the aws-sdk in golang.

##### [vault api](github.com/hashicorp/vault/api)
##### [vault logical api](github.com/hashicorp/vault/logical)

##### [goconvey](http://github.com/smartystreets/goconvey)
This is a BDD-style testing library for golang.
