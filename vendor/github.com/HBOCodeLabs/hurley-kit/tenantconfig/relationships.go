package tenantconfig

var platformTenantToProducts = map[PlatformTenantCode][]ProductCode{
	PlatformTenantCodeHBODirect: {ProductCodeHBOMAX, ProductCodeHBONOW},
	PlatformTenantCodeHBOTVE:    {ProductCodeHBOGO, ProductCodeMAXGO},
}

var (
	productToPlatformTenantFunc = func() map[ProductCode]PlatformTenantCode {
		mapper := map[ProductCode]PlatformTenantCode{}
		for platformTenantCode, productCodes := range platformTenantToProducts {
			for _, productCode := range productCodes {
				mapper[productCode] = platformTenantCode
			}
		}
		return mapper
	}
	productToPlatformTenant = productToPlatformTenantFunc()
)
