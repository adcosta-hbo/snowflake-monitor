# tokens
> a Go package for decoding and verifying HBO GO auth tokens

### Overview

The `tokens` package contains functionality for decoding GO/NOW auth tokens,
verifying their signatures, and accessing the data contained within. It has
been in production use across several Go services.

## Versioning

### 11-Mar-2020 SHA 54a2b60ab983cd7f871a76e9179818db79b690e7
* Permissions() method added which returns the current permissions in the token.
* UserInfo() added for backward compatibility (to be deprecated later)
* PlatformTenantCode added to contextdefs

### 7-Feb-2020 SHA  1bd68cb5e26ec1636f49dc1873e38b9bd75f2dbe
*  Remove `warnerMediaDirect` as a `platformTenantCode` for NOW --> MAX migration. It will no longer be assigned as a `platoformTenantCode` in a `token`. 

### 10-Oct-2019 SHA d2aa9b0f467415ab989d959daf6da8e11b4b5b34
*  Added support for extracting `HurleyAccountID` from the `hurleyAccountId` field

### 29-Jul-2019 SHA 02c7aba6bf0581167e6410971e45fa69f00abbc3
*  Moved `ProductCode` enumeration to `tenantconfig` package
*  Removed non-scalabile is(Go|Now|Max)Token; use existing isProductCode() convenience function instead

### 01-Oct-2018 SHA 8f62a70597f787bebf9f6a64cf5f7621a4b5a753
*  Added the `tokens.IsProviderGuest` API to tokens to check if the accountProviderCode is a guest service

### 21-May-2018 SHA fb8d9b3f42e513fe87ad0df1d50f62ab1bc76e0d
*  Removed the `tokens.IsProviderMLBAM` API to tokens to check if the accountProviderCode is MLBAM

### Examples

```
package main

import (
  "fmt"
  "time"

  "github.com/HBOCodeLabs/hurley-kit/auth/tokens"
)

func main() {

  secret := "$om3thing$3cret!!!"

  decoder := tokens.NewDecoder(secret)

  token, err := decoder.Decode("someBigLongEncodedTokenValueFromAnHTTPHeader")
  if err != nil {
    log.Fatalf("Oh no! The token could not be decoded! %s", err.Error())
  }

  log.Println("Cool, your token was decoded!")
}

```

The package's godocs, and the testable examples in the various `*_test.go`
source files are always the place to find the most up-to-date examples.
However, for a quick illustration of the message queueing functionality:

### Dependencies

This package uses the [jwt-go](https://github.com/dgrijalva/jwt-go) package for
performing the actual JWT token decoding.  To install the package, run the
following Go command:

```
go get -u github.com/dgrijalva/jwt-go
```

### Local development

Please see the [contribution instructions](../../.github/CONTRIBUTING.md) for
details on local development and PR guidelines.

### Examples

The Go code itself contains testable examples.  For the most up-to-date
examples, see the source files named `*_test.go`.

