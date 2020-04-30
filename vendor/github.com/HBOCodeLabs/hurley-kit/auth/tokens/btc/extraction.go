package btc

import (
	"fmt"

	"github.com/HBOCodeLabs/hurley-kit/auth/tokens"
	"github.com/HBOCodeLabs/hurley-kit/btc"
	"github.com/HBOCodeLabs/hurley-kit/tenantconfig"
)

var (
	brandByProductCode = map[tenantconfig.ProductCode]btc.Brand{
		tenantconfig.ProductCodeHBOGO:  btc.BrandHBO,
		tenantconfig.ProductCodeHBOMAX: btc.BrandHBOMax,
		tenantconfig.ProductCodeHBONOW: btc.BrandHBO,
		tenantconfig.ProductCodeMAXGO:  btc.BrandMAX,
	}

	territoryByProductAndCountryCode = map[tenantconfig.ProductCode]map[string]btc.Territory{
		tenantconfig.ProductCodeHBOGO: {
			"US": btc.TerritoryHBODomestic,
		},
		tenantconfig.ProductCodeHBOMAX: {
			"US": btc.TerritoryHBOMaxDomestic,
		},
		tenantconfig.ProductCodeHBONOW: {
			"US": btc.TerritoryHBODomestic,
		},
		tenantconfig.ProductCodeMAXGO: {
			"US": btc.TerritoryMAXDomestic,
		},
	}

	channelByProductCode = map[tenantconfig.ProductCode][]btc.Channel{
		tenantconfig.ProductCodeHBOGO:  {btc.ChannelHBOGoFree, btc.ChannelHBOGoSubscription},
		tenantconfig.ProductCodeHBOMAX: {btc.ChannelHBOMaxFree, btc.ChannelHBOMaxSubscription},
		tenantconfig.ProductCodeHBONOW: {btc.ChannelHBONowFree, btc.ChannelHBONowSubscription},
		tenantconfig.ProductCodeMAXGO:  {btc.ChannelMAXGoFree, btc.ChannelMAXGoSubscription},
	}
)

// ExtractBTC returns the brand, channels, and territory from a Hurley OAuth token
func ExtractBTC(t tokens.Tokener) (btc btc.BTC, err error) {
	btc.Brand, err = extractBrand(t)
	if err != nil {
		return
	}

	btc.Territory, err = extractTerritory(t)
	if err != nil {
		return
	}

	btc.Channels, err = extractChannels(t)
	if err != nil {
		return
	}

	return
}

// extractBrand returns the brand string encoded in the token
func extractBrand(t tokens.Tokener) (brand btc.Brand, err error) {
	pc := t.ProductCode()

	brand, ok := brandByProductCode[pc]
	if !ok {
		err = fmt.Errorf("no matching brand found for product code: %q", pc)
		return
	}

	return
}

// extractTerritory returns the territory string encoded in the token
func extractTerritory(t tokens.Tokener) (territory btc.Territory, err error) {
	pc := t.ProductCode()
	cc := t.CountryCode()

	territoryByCountryCode, ok := territoryByProductAndCountryCode[pc]
	if !ok {
		err = fmt.Errorf("no territories found for product code: %q", pc)
		return
	}

	territory, ok = territoryByCountryCode[cc]
	if !ok {
		err = fmt.Errorf("no territory found for country code: %q", cc)
		return
	}

	return
}

// extractChannels returns the channel(s) encoded in the token.
func extractChannels(t tokens.Tokener) (channels []btc.Channel, err error) {
	pc := t.ProductCode()

	channels, ok := channelByProductCode[pc]
	if !ok {
		err = fmt.Errorf("no channels found for product code: %q", pc)
		return
	}

	return
}
