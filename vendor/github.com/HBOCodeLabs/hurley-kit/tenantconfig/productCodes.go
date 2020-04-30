package tenantconfig

import "fmt"

// ProductCode is an enum specifying the streaming application branding
type ProductCode string

const (
	// ProductCodeHBOGO is for HBO GO
	ProductCodeHBOGO ProductCode = "hboGo"
	// ProductCodeHBOMAX is for HBOMAX
	ProductCodeHBOMAX ProductCode = "hboMax"
	// ProductCodeHBONOW is for HBO NOW
	ProductCodeHBONOW ProductCode = "hboNow"
	// ProductCodeMAXGO is for MAX Go
	ProductCodeMAXGO ProductCode = "maxGo"
)

// String converts ProductCode value to a string
func (productCode ProductCode) String() string {
	return string(productCode)
}

var (
	// ProductCodes lists all the supported productCodes
	ProductCodes = []ProductCode{
		ProductCodeHBOGO,
		ProductCodeHBOMAX,
		ProductCodeHBONOW,
		ProductCodeMAXGO,
	}

	strToProductCodeFunc = func() map[string]ProductCode {
		mapper := map[string]ProductCode{}
		for _, productCode := range ProductCodes {
			mapper[productCode.String()] = productCode
		}
		return mapper
	}
	strToProductCode = strToProductCodeFunc()
)

// GetProductCode returns ProductCode when given a string.
func GetProductCode(str string) (ProductCode, error) {
	productCode, isPresentInMap := strToProductCode[str]
	if isPresentInMap {
		return productCode, nil
	}

	return "", fmt.Errorf("no matching ProductCode found for string: %q", str)
}

// GetPlatformTenantCodeFor gets parent platform tenant code for a given product code (e.g. hboGo -> hboTve)
func GetPlatformTenantCodeFor(productCode ProductCode) (PlatformTenantCode, error) {
	platformTenantCode, isPresentInMap := productToPlatformTenant[productCode]
	if isPresentInMap {
		return platformTenantCode, nil
	}

	return "", fmt.Errorf("no matching platformTenantCode found for productCode: %q", productCode)
}
