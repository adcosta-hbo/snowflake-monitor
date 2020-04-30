package btc

// RegionCode for Country mapping
type RegionCode string

const (

	// list of Regions that will eventually map to the US
	RegionCodeGU RegionCode = "GU"
	RegionCodePR RegionCode = "PR"
	RegionCodeVI RegionCode = "VI"
	RegionCodeAS RegionCode = "AS"
	RegionCodeMP RegionCode = "MP"
	RegionCodeUM RegionCode = "UM"
)

// maps US Region  to US CountryCode
var (
	regionToCountry = map[RegionCode]CountryCode{
		RegionCodeGU: CountryCodeUS,
		RegionCodePR: CountryCodeUS,
		RegionCodeVI: CountryCodeUS,
		RegionCodeAS: CountryCodeUS,
		RegionCodeMP: CountryCodeUS,
		RegionCodeUM: CountryCodeUS,
	}
)
