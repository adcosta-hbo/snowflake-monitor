// Package btc contains types and constants used for representing BTC (Brand, Territory,
// Channels). Read more about BTC on the wiki: https://wiki.hbo.com/display/HUR/BTC
//
// BTC Token utilities are contained in hurley-kit/auth/tokens/btc
package btc

import (
	"fmt"
	"strings"

	"github.com/HBOCodeLabs/hurley-kit/tenantconfig"
)

// BTC stands for Brand, Territory, and Channels. The 3 fields are used to
// restrict offerings, videos, playlists, etc. They can be thought of as
// an entitlement restriction.
type BTC struct {
	Brand     Brand
	Territory Territory
	Channels  Channels
}

// Brand represents the "corporate banner" or business entity that is providing a piece of
// content. In general usage, a brand is a name, term, sign, or design (or some unifying
// combination thereof) that is intended to identify and distinguish a product or service
// from its competitors.
type Brand string

// String converts Brand value to a string
func (brand Brand) String() string {
	return string(brand)
}

// Territory is a business concept to describe a geographical region in which we provide
// entitlements to a set of content. A territory can be a country, a set of countries, or
// an area within a country; however, we do not support defining a hierarchy of
// Territories. Territories cannot overlap (a user can't be in one location that's
// considered part of two distinct Territories.)
type Territory string

// String converts Territory value to a string
func (territory Territory) String() string {
	return string(territory)
}

// Channel represents the distribution, marketing or subscription that is used to deliver
// the content. Channel describes the method a customer uses to get HBO or MAX
// (or other Brand) content. Channel distinction within a single Territory may be required
// due to differences in promotional materials or content entitlements based on the
// distribution method or subscription type.
type Channel string

// String converts Channel value to a string
func (channel Channel) String() string {
	return string(channel)
}

// Channels represents an array of type Channel.
type Channels []Channel

// JoinToString will return a string of channels deliminated by a deliminator
func (cs Channels) JoinToString(del string) string {
	str := make([]string, len(cs))
	for i, c := range cs {
		str[i] = c.String()
	}
	return strings.Join(str, del)
}

// GetTerritoryForCountry will return the territory based on the country and product
func GetTerritoryForCountry(country CountryCode, productCode tenantconfig.ProductCode) (territory Territory, err error) {
	switch productCode {
	case tenantconfig.ProductCodeHBOGO:
		territory = hboCountryToTerritory[country]
	case tenantconfig.ProductCodeHBOMAX:
		territory = hboMaxCountryToTerritory[country]
	case tenantconfig.ProductCodeHBONOW:
		territory = hboCountryToTerritory[country]
	case tenantconfig.ProductCodeMAXGO:
		territory = maxCountryToTerritory[country]
	default:
		err = fmt.Errorf("no matching territory found for productCode: %q", productCode)
	}
	return
}

// ValidRegion returns whether the regionCode is found in the mappings
func ValidRegion(region RegionCode) (CountryCode, bool) {
	country, ok := regionToCountry[region]
	return country, ok
}

// GetCountryForRegion maps US region to the country
func GetCountryForRegion(region RegionCode) CountryCode {
	if country, ok := ValidRegion(region); ok {
		return country
	}
	return CountryCode(region)
}

// AllBrands returns an array of all the Brands
func AllBrands() (arrayOfBrands []Brand) {
	arrayOfBrands = []Brand{BrandHBO, BrandHBOMax, BrandMAX}
	return
}

const (

	// BrandHBO is the brand string used by HBO tokens
	BrandHBO Brand = "HBO"

	// BrandHBOMax is the brand string used by HBOMAX tokens
	BrandHBOMax Brand = "HBO MAX"

	// BrandMAX is the brand string used by MAX tokens
	BrandMAX Brand = "MAX"

	// TerritoryHBODomestic is the territory string used by domestic (aka US) HBO tokens
	TerritoryHBODomestic Territory = "HBO DOMESTIC"

	// TerritoryHBOMaxDomestic is the territory string used by domestic (aka US) HBOMAX tokens
	TerritoryHBOMaxDomestic Territory = "HBO MAX DOMESTIC"

	// TerritoryMAXDomestic is the territory string used by domestic (aka US) MAX tokens
	TerritoryMAXDomestic Territory = "MAX DOMESTIC"

	// TerritoryHBOLag is is the territory used by international  Latin American HBO tokens
	TerritoryHBOLag Territory = "HBO LAG"

	// TerritoryHBOBrazil is is the territory used by Brazil HBO tokens
	TerritoryHBOBrazil Territory = "HBO BRAZIL"

	// ChannelHBOGoFree is the channel string used by non-logged-in HBO Go tokens
	ChannelHBOGoFree Channel = "HBO GO FREE"

	// ChannelHBOGoSubscription is the channel string used by logged-in HBO Go tokens
	ChannelHBOGoSubscription Channel = "HBO GO SUBSCRIPTION"

	// ChannelHBOMaxFree is the channel string used by non-logged-in HBOMAX tokens
	ChannelHBOMaxFree Channel = "HBO MAX FREE"

	// ChannelHBOMaxSubscription is the channel string used by logged-in HBOMAX tokens
	ChannelHBOMaxSubscription Channel = "HBO MAX SUBSCRIPTION"

	// ChannelHBONowFree is the channel string used by non-logged-in HBO Now tokens
	ChannelHBONowFree Channel = "HBO NOW FREE"

	// ChannelHBONowSubscription is the channel string used by logged-in HBO Now tokens
	ChannelHBONowSubscription Channel = "HBO NOW SUBSCRIPTION"

	// ChannelMAXGoFree is the channel string used by non-logged-in MAX Go tokens
	ChannelMAXGoFree Channel = "MAX GO FREE"

	// ChannelMAXGoSubscription is the channel string used by logged-in MAX Go tokens
	ChannelMAXGoSubscription Channel = "MAX GO SUBSCRIPTION"
)
