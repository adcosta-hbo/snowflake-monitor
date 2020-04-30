package btc

// maps CountryCode  to territory
var (
	hboCountryToTerritory = map[CountryCode]Territory{
		CountryCodeUS: TerritoryHBODomestic,
		CountryCodeBR: TerritoryHBOBrazil,
		CountryCodeAR: TerritoryHBOLag,
		CountryCodeBB: TerritoryHBOLag,
	}

	hboMaxCountryToTerritory = map[CountryCode]Territory{
		CountryCodeUS: TerritoryHBOMaxDomestic,
	}

	maxCountryToTerritory = map[CountryCode]Territory{
		CountryCodeUS: TerritoryMAXDomestic,
	}
)
