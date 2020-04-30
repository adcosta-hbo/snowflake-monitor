# BTC
A Go package containing BTC (brand, territory, and
channel) type definitions and constants. For details on deriving BTC from auth tokens, see sibling [auth/tokens/btc package](https://github.com/HBOCodeLabs/hurley-kit/auth/tokens/btc).

The package has been extended to include region-to-country mappings and country-to-territory mappings based on the business guidelines. These mappings are kept in this centralized location to be used by Geovalidation library and services (Mercator and Magellan respectively). Magellan will use the mapping lookup functions in it's function `determineAuthorizationForTerritory` that calls GetTerritoryForCountry based on the country obtained from the ipAddress and which then calls GetTerritoryForCountry to get the actual HBO defined territory. https://github.com/HBOCodeLabs/Hurley-Magellan/blob/a4f12f4d65e78f1ccb0d98dbb765f620a5fa20bc/lib/models/Geovalidation.js#L212
