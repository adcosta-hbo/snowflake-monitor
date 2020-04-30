# BTC
A Go package containing functionality to derive BTC (brand, territory, and
channel) data from a hurley OAuth token, like those handled in the [tokens
package](https://github.com/HBOCodeLabs/hurley-kit/auth/tokens).

This library is a counterpart of the [NodeJS Hurley-BTC](https://github.com/HBOCodeLabs/Hurley-btc) library.
**!! Therefore !!** the changes made here should also be made in the NodeJS library, and vice versa.

BTC stands for Brand, Territory, and Channel.  They're different dimensions on
the Catalog offerings that restricts (filter out) contents from users.

Brand
-----
Brand should be used to determine the Brand through which the content is being
distributed (HBO, Cinemax).

Territory
---------
Territory should be used to determine the logical geographic territory in which
the content is being distributed.

Channels
-------
Channels should *not* be confused with TV channels - it refers the *channel of
sale*. In the near future, there are four: GO Subscription, GO Free, NOW
Subscription, NOW Free. Later, there might be others that cover emerging
distribution channels.

### Version History

## Versioning

### 10-Sep-2019 SHA 46dd3e4d6921b17198bbbc2c06d176800fd8235b

[PR #59](https://github.com/HBOCodeLabs/hurley-kit/pull/59) discontinues use of service code for
resolution; uses product code instead.

### 09-Oct-2018 SHA 9bd6c954b1b7dbf8c9c6efde5cb9b4a926052e05

[PR #27](https://github.com/HBOCodeLabs/hurley-kit/pull/27) uses the tokens.Tokener interface
instead of the actual tokens object.

### 12-Apr-2018

[PR #9](https://github.com/HBOCodeLabs/hurley-kit/pull/9) moves the standalone BTC library
to be included in hurley-kit. Introduces breaking changes:

- Divided BTC into 2 packages: `hurley-kit/btc` for constants and `hurley-kit/auth/tokens/btc` for token extraction.

- Added new typedefs for `Brand`, `Territory`, and `Channel` for improved type safety.

- Changed type of `Channels` in `BTC` to `[]Channel`.

- Added errors to return signitures of Extraction functions if the token contains BTC data that is unsupported.

### 29-Sep-2017, SHA `b38b1d5a44ba97e5d53f06c2f97171cfcc68a1d2`

[PR #3](https://github.com/HBOCodeLabs/BTC/pull/3) introduced several breaking
changes to the package:

- Removes `DeriveBTCFromHurleyAuthToken` function.  Use `ExtractBTC` instead.

- Removes exported constants `ChannelGO` and `ChannelNOW`.  Use
`ChannelHBOGoFree`, `ChannelHBOGoSubscription`, `ChannelHBONowFree`, and
`ChannelHBONowSubscription` instead.

- Removes the exported constant `TerritoryDomestic`.  Use `TerritoryHBODomestic` instead.

- Renames the `Channel` field of the `BTC` type to be `Channels`, since the
  value is a comma-delimited string with (potentially) more than 1 value.

### Examples
```
package main

import (
  "fmt"
  "time"

  "github.com/HBOCodeLabs/hurley-kit/auth/tokens"
  "github.com/HBOCodeLabs/hurley-kit/auth/tokens/btc"
  extract "github.com/HBOCodeLabs/hurley-kit/auth/tokens/btc"
)

func main() {

  secret := "$om3thing$3cret!!!"
  decoder := tokens.NewDecoder(secret)

  token, err := decoder.Decode("someBigLongEncodedTokenValueFromAnHTTPHeader")
  if err != nil {
    log.Fatalf("Oh no! The token could not be decoded! %s", err.Error())
  }

  log.Println("Cool, your token was decoded!")
  inputBTC, err := extract.ExtractBTC(token)
  if err != nil {
    log.Fatalf("Oh no! The token must not have a valid BTC! %s", err.Error())
  }

  log.Printf("Cool, you have a BTC: %#v\n", inputBTC)

  if inputBTC.Brand == btc.BrandHBO {
    log.Printf("This is an HBO BTC\n")
  }
}
```
