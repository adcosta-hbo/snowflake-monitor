package tokens

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/tenantconfig"
)

const (
	// RatingNC17 is a value used in the parental controls map to indicate NC-17
	// movies
	RatingNC17 = "NC-17"

	// RatingTVMA is a value used in the parental controls map to indicate
	// MA-rated TV shows
	RatingTVMA = "TV-MA"

	// Movies represents the key used to store parental controls for movies
	Movies = "movies"

	// TV represents the keys used to store parental controls for TV shows
	TV = "tv"

	// AccountProviderGuest is the value encoded in the token when a user's
	// account is stored in a guest system (ie.. Hilton)
	AccountProviderGuest = "guest"

	// AccountProviderHurley is the value encoded in the token when a user's
	// account is stored in the Hurley systems
	AccountProviderHurley = "hurley"
)

// A Token represents an auth token for HBO GO.
type Token struct {
	// Raw is the original string value of the encoded token
	Raw string `json:"Raw"`

	// claims is used for decoding the token from the encoded value in an
	// HTTP auth header. It is unexported so that users of the tokens package
	// don't come to depend on a particular token structure
	// (http://www.hyrumslaw.com/), which is subject to change at the discretion of
	// the PIT/U+A team.
	claims `json:"claims"`
}

// Tokener encapsulates the methods on the tokens.Token type that we use when
// exercising the middleware. In other words, these are the methods that that
// the middleware calls on the tokens.Token type, so by using an interface, we
// can swap "real" Token instances w/ mock implementations for testing puposes.
type Tokener interface {
	HurleyAccountID() string
	ClientID() string
	CountryCode() string
	DeviceCode() string
	DeviceSerialNumber() string
	Environment() string
	HasAllPermissions([]int) bool
	Permissions() []int
	IsAccountProviderCode(string) bool
	IsExpired() bool
	IsPlatformTenantCode(platformTenantCode tenantconfig.PlatformTenantCode) bool
	IsProductCode(productCode tenantconfig.ProductCode) bool
	IsProviderGuest() bool
	IsProviderHurley() bool
	PlatformTenantCode() tenantconfig.PlatformTenantCode
	ProfileID() string
	ProductCode() tenantconfig.ProductCode
	ServiceCode() string
	UserID() string
	UserInfo() (string, error)
}

// HurleyAccountID returns the HurleyAccountID encoded in the token.  It is the unique
// identifier for a user's account.
func (t *Token) HurleyAccountID() string {
	return t.claims.Payload.TokenPropertyData.HurleyAccountID
}

// IsExpired returns true if the authorization expiration encoded in the token
// is less than the current system time.
func (t *Token) IsExpired() bool {
	// the authz expiration is stored in epoch ms, so we convert to seconds
	// for the comparison with the epoch seconds returned by .Unix()
	return t.claims.Payload.ExpirationMetadata.AuthzExpirationUTC/1000 < time.Now().UTC().Unix()
}

// ServiceCode returns the name of the service that the token is being used for (HBO, MAX, etc.)
func (t *Token) ServiceCode() string {
	return t.claims.Payload.TokenPropertyData.ServiceCode
}

// CountryCode returns the name of the country that the token is authorized for
func (t *Token) CountryCode() string {
	return t.claims.Payload.TokenPropertyData.CountryCode
}

// PlatformTenantCode returns the platform tenant code that the token is authorized for
func (t *Token) PlatformTenantCode() tenantconfig.PlatformTenantCode {
	return t.claims.Payload.TokenPropertyData.PlatformTenantCode
}

// ProductCode returns the product code that the token is authorized for
func (t *Token) ProductCode() tenantconfig.ProductCode {
	return t.claims.Payload.TokenPropertyData.ProductCode
}

// ClientID returns the ID (GUID) assigned to the client application. This ID
// will be the same across instances of the same client application (e.g.
// iphone, firetv, etc.)
func (t *Token) ClientID() string {
	return t.claims.Payload.TokenPropertyData.ClientID
}

// DeviceCode returns the name of the device code encoded in the token. These
// devicecodes identify the "class" of device, but not an individual device
// instance. An example device code might be "android" or "desktop".
func (t *Token) DeviceCode() string {
	return t.claims.Payload.TokenPropertyData.DeviceCode
}

// DeviceSerialNumber returns the serial number for an individual device that
// is encoded in the token.
func (t *Token) DeviceSerialNumber() string {
	return t.claims.Payload.TokenPropertyData.DeviceSerialNumber
}

// ProfileID return the ProfileID in the token if it's availble.  There is a
// case where the ProfileID can be missing.  Then the fallback is the userID.
func (t *Token) ProfileID() string {
	if t.claims.Payload.TokenPropertyData.ProfileID != "" {
		return t.claims.Payload.TokenPropertyData.ProfileID
	}
	return t.claims.Payload.TokenPropertyData.UserID
}

// UserID returns the UserID encoded in the token. This userID is the unique
// identifier used to locate a user's account info, and is not related to a
// screen name or "friendly name".
func (t *Token) UserID() string {
	return t.claims.Payload.TokenPropertyData.UserID
}

// IsProviderGuest returns true if the token's "account provider" (e.g. the
// system-of-record for the user's account) is a guest system (e.g. Hilton
// hotel), false otherwise
func (t *Token) IsProviderGuest() bool {
	return t.IsAccountProviderCode(AccountProviderGuest)
}

// IsProviderHurley returns true if the token's "account provider" (e.g. the
// system-of-record for the user's account) is v2/Hurley, false otherwise
func (t *Token) IsProviderHurley() bool {
	return t.IsAccountProviderCode(AccountProviderHurley)
}

// IsAccountProviderCode returns true if the token contains a provider code
// equal to `accountProviderCode`. This is used to indicate which system is the
// system-of-record for a user's account.
//
// The name of the function was chosen to achieve parity with the Javascript
// equivalent (in Hurley-AuthCheck) and to make searching for usages more
// convenient, so care should be exercised if renaming.
func (t *Token) IsAccountProviderCode(accountProviderCode string) bool {
	return t.claims.Payload.TokenPropertyData.AccountProviderCode == accountProviderCode
}

// IsPlatformTenantCode is a convenience method to easily pivot on platformTenantCode ('hboGo' or 'hboNow' or 'maxGo')
func (t *Token) IsPlatformTenantCode(desiredPlatformTenantCode tenantconfig.PlatformTenantCode) bool {
	return t.claims.Payload.TokenPropertyData.PlatformTenantCode == desiredPlatformTenantCode
}

// IsProductCode is a convenience method to easily pivot on productCode ('hboGo' or 'hboNow' or 'maxGo')
func (t *Token) IsProductCode(desiredProductCode tenantconfig.ProductCode) bool {
	return t.claims.Payload.TokenPropertyData.ProductCode == desiredProductCode
}

// Environment returns the environment associated with the token
func (t *Token) Environment() string {
	return t.claims.Payload.Environment
}

// HasAllPermissions returns true if the token contains all of the supplied
// permissions, false otherwise.  Note that the permissions supplied must be
// valid permission as defined by the constants in this package.
func (t *Token) HasAllPermissions(permissions []int) bool {
	tokenPermissions := t.claims.Payload.TokenPropertyData.Permissions

	for _, permission := range permissions {
		if !contains(tokenPermissions, permission) {
			return false
		}
	}
	return true
}

// Permissions returns the []int of all permissions from the Token
func (t *Token) Permissions() []int {
	return t.claims.Payload.TokenPropertyData.Permissions
}

// UserInfo provides backward compatibility to github.com/HBOCodelabs/tokens/decode.go for setting UserInfo in ctx
// UserInfo is used by legacy methods of extracting the token and inserting it into `ctx` which is later used to set the X-UserInfo header by services
// TODO: to be deprecated
func (t *Token) UserInfo() (string, error) {
	payload, err := json.Marshal(t.claims.Payload)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

// contains returns true if slice `s` contains element `e`, false otherwise
func contains(s []int, e int) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// setRaw implements the setRawer interface, used by the token parsing code
func (t *Token) setRaw(rawTokenValue string) {
	t.Raw = rawTokenValue
}

// claims is the structure that contains the information transmitted by the
// encoded token, along with the timestamp and expiration metadata.
type claims struct {
	Payload    payload `json:"payload"`
	Expiration int64   `json:"expiration"`
	Timestamp  int64   `json:"timestamp"`
}

// Valid implements the jwt.claims interface, which lets use this when decoding
// a token.
func (c claims) Valid() error {
	// TODO: implement some more rigorous validation
	if c.Payload.Environment == "" || c.Payload.TokenPropertyData.DeviceCode == "" {
		return errors.New("invalid token: missing environment and/or device code")
	}
	return nil
}

// payload is the client/user-specific information conveyed by the encoded
// token.
type payload struct {
	HistoricalMetadata historicalMetadata `json:"historicalMetadata"`
	TokenPropertyData  tokenPropertyData  `json:"tokenPropertyData"`
	ExpirationMetadata expirationMetadata `json:"expirationMetadata"`
	CurrentMetadata    currentMetadata    `json:"currentMetadata"`
	Permissions        []int              `json:"permissions"` // e.g. PLAY_VIDEO. See https://github.com/HBOCodeLabs/Hurley-AuthCheck/blob/master/lib/Permissions.js
	TokenType          string             `json:"token_type"`  // e.g. "access"
	Environment        string             `json:"environment"` // TODO: duplicative
	Version            int                `json:"version"`     // TODO: duplicative
}

// expirationMetadata contains the timestamps that govern the lifetime of the
// token for authorization and authentication purposes.
type expirationMetadata struct {
	AuthzTimeout       int64 `json:"authzTimeoutMs"`
	AuthnTimeout       int64 `json:"authnTimeoutMs"`
	AuthzExpirationUTC int64 `json:"authzExpirationUtc"`
	AuthnExpirationUTC int64 `json:"authnExpirationUtc"`
}

// currentMetadata encapsulates miscellaneous metadata about the issuing
// environment for the token.
type currentMetadata struct {
	Environment     string `json:"environment"`
	Version         int    `json:"version"`
	Nonce           string `json:"nonce"` // GUID
	IssuedTimestamp int64  `json:"issuedTimestamp"`
}

// historicalMetadata encapsulates historical data about the token instance
// that may or may not still be relevant.
type historicalMetadata struct {
	OriginalIssuedTimestamp int64  `json:"originalIssuedTimestamp"`
	OriginalGrantType       string `json:"originalGrantType"`
	OriginalVersion         int    `json:"originalVersion"`
	DevToolsDescription     string `json:"devToolsDescription,omitempty"`
}

// tokenPropertyData contains information about the user, device, account, etc.
type tokenPropertyData struct {
	ClientID            string                          `json:"clientId"`                      // GUID assigned to the GO client, e.g. all AppleTV GO installs
	DeviceCode          string                          `json:"deviceCode"`                    // Type of device, e.g. "desktop"
	DeviceSerialNumber  string                          `json:"deviceSerialNumber"`            // ID of specific client device. Only unique within a deviceCode.
	Permissions         []int                           `json:"permissions"`                   // e.g. PLAY_VIDEO. See https://github.com/HBOCodeLabs/Hurley-AuthCheck/blob/master/lib/Permissions.js
	PlatformTenantCode  tenantconfig.PlatformTenantCode `json:"platformTenantCode"`            // Platform tenant, e.g. "hboTve", "hboD2c"
	ProductCode         tenantconfig.ProductCode        `json:"productCode"`                   // Product being sold, e.g. "hboGo", "hboNow", "hboMax"
	ServiceCode         string                          `json:"serviceCode"`                   // "HBO", "MAX"
	AccountProviderCode string                          `json:"accountProviderCode,omitempty"` // User data storage, "hurley" or "mlbam"
	ProfileID           string                          `json:"hurleyProfileId,omitempty"`     // Hurley unique profile ID
	UserID              string                          `json:"userId,omitempty"`              // Hurley unique user ID
	UserTKey            string                          `json:"userTkey,omitempty"`            // V1-style user ID (deprecated, use userId unless tkey is specified)
	Affiliate           string                          `json:"affiliate,omitempty"`           // (Deprecated, use affiliateCode)
	AffiliateCode       string                          `json:"affiliateCode,omitempty"`       // User’s identity provider for authn/authz, e.g. "COMCAST"
	CountryCode         string                          `json:"countryCode,omitempty"`         // e.g. "US"
	ParentalControls    parentalControls                `json:"parentalControls,omitempty"`    // e.g. { "movie": "NC-17" }
	OpaqueID            string                          `json:"opaqueId,omitempty"`            // Affiliate-provided user ID
	StreamTrackingID    string                          `json:"streamTrackingId,omitempty"`    // ID for a group of users sharing the same stream limit. Sent to Groot
	PlatformType        string                          `json:"platformType"`                  // "tv", "desktop", "phone"
	ClientDeviceData    clientDeviceData                `json:"clientDeviceData"`              // {"paymentProviderCode": "google-play"}
	CustSvcOrgID        string                          `json:"custSvcOrgId,omitempty"`        // Customer Service Org ID. ie harteHanksCustSvcOrgId
	CustSvcAgentRole    string                          `json:"custSvcAgentRole,omitempty"`    // "Tier 1", "Tier 2"
	CustSvcAgentUserID  string                          `json:"custSvcAgentUserId,omitempty"`  // UserId of the customer agent
	HurleyAccountID     string                          `json:"hurleyAccountId,omitempty"`     // Hurley unique account ID
}

// parentalControls represents a maximum rating allowed for several types of
// videos (movies, tv shows, etc.).
//
// As of 07-Sep-2017, the decoded value can be a map of string ->
// {string|number}, hence the empty interface typedef.
//
// Example:
//
// "parentalControls": {
// 	 "ratingSystem": "US",
// 	 "movie": "NC-17",
// 	 "movieRestriction": 1000,
// 	 "tv": "TV-MA",
// 	 "tvRestriction": 1000
// 	},
type parentalControls map[string]interface{}

// clientDeviceData will either be an empty object or have the single property "paymentProviderCode" currenlty on 1/2018.
// but we could add more properties in the future.
type clientDeviceData map[string]interface{}
