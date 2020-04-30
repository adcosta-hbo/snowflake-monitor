package tenantconfig

import "fmt"

// PlatformTenantCode is an enum specifying the distinct and disjoint entity running their applications on the platform
type PlatformTenantCode string

const (
	// PlatformTenantCodeHBODirect is for the tenant managing legacy HBO direct-to-consumer apps (i.e. HBO Now & HBO MAX)
	PlatformTenantCodeHBODirect PlatformTenantCode = "hboDirect"
	// PlatformTenantCodeHBOTVE is for the tenant managing legacy HBO TV everywhere apps (i.e. HBO GO & MAX GO)
	PlatformTenantCodeHBOTVE PlatformTenantCode = "hboTve"
)

// String converts PlatformTenantCode enum value to a string
func (platformTenantCode PlatformTenantCode) String() string {
	return string(platformTenantCode)
}

var (
	// PlatformTenantCodes lists all the supported PlatformTenantCodes
	PlatformTenantCodes = []PlatformTenantCode{
		PlatformTenantCodeHBODirect,
		PlatformTenantCodeHBOTVE,
	}

	strToPlatformTenantCodeFunc = func() map[string]PlatformTenantCode {
		mapper := map[string]PlatformTenantCode{}
		for _, platformTenantCode := range PlatformTenantCodes {
			mapper[platformTenantCode.String()] = platformTenantCode
		}
		return mapper
	}
	strToPlatformTenantCode = strToPlatformTenantCodeFunc()
)

// GetPlatformTenantCode returns PlatformTenantCode when given a string.
func GetPlatformTenantCode(str string) (PlatformTenantCode, error) {
	platformTenantCode, isPresentInMap := strToPlatformTenantCode[str]
	if isPresentInMap {
		return platformTenantCode, nil
	}

	return "", fmt.Errorf("no matching PlatformTenantCode found for string: %q", str)
}

// GetProductCodesFor get child product code(s) for a given platform tenant code (e.g. hboTve -> [ hboGo, maxGo ])
func GetProductCodesFor(platformTenantCode PlatformTenantCode) ([]ProductCode, error) {
	productCodes, isPresentInMap := platformTenantToProducts[platformTenantCode]
	if isPresentInMap {
		return productCodes, nil
	}

	return nil, fmt.Errorf("no matching productCodes found for platformTenantCode: %q", platformTenantCode)
}
