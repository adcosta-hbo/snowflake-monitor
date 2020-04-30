# tenantconfig
> a Go package for defining platform tenant configuration APIs

### Overview

The purpose of this library is to define the API of platform tenant configuration.  As such, the
library contains:

* Enumerations
  * `PlatformTenantCode` - Globally unique identifier for a platform tenant for programmatic business rules.
  * `ProductCode` - Globally unique identifier for one of possibly many platform tenant branded experience with 
  specific catalog access.
* Configuration Schemas (TBD)

See also [Core Terminology](https://wiki.hbo.com/pages/viewpage.action?spaceKey=PDLC&title=Core+Terminology).

## Versioning

### 7-Feb-2020 SHA 1bd68cb5e26ec1636f49dc1873e38b9bd75f2dbe
* Removes `warneMediaDirect` as a `platformTenantCode` for NOW -> MAX migration

### 19-Aug-2019 SHA f1cf1d3d936b8f58d0190f1397e1ec58c3cfeb51
* Ensure Full object (de)serializers for enumerations
* Rename platform tenant HBO D2C -> HBO Direct

### 29-Jul-2019 SHA 02c7aba6bf0581167e6410971e45fa69f00abbc3
*  Initial release

### Examples

The Go code itself contains testable examples.  For the most up-to-date
examples, see the source files named `*_test.go`.

### Local development

Please see the [contribution instructions](../../.github/CONTRIBUTING.md) for
details on local development and PR guidelines.
