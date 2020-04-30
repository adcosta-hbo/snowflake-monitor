package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/HBOCodeLabs/hurley-kit/auth/tokens"
	"github.com/HBOCodeLabs/hurley-kit/contextdefs"
)

const (
	// AuthorizationHeaderName is the name of the HTTP header that is used to store/propagate
	// the Bearer token.
	AuthorizationHeaderName = "Authorization"

	// BearerPrefix is the prefix used on the value of an Authorization header.
	// Note that the trailing space character is intentional, as it needs to be
	// stripped off in order to obtain the token value.
	// Example: "Authorization: Bearer <token>"
	BearerPrefix = "Bearer "
)

// TokenDecoder encapsulates the functionality needed to decode a token,
// allowing for the middleware to use a real implementation, or a mock
// implementation for testing.
type TokenDecoder interface {
	// Decode takes an encoded token value (typically a "Bearer" value from an
	// "authorization" header), decodes it, and performs signature verification.
	Decode(tokenValue string) (tokens.Tokener, error)
}

// ConfigurationFunc is a typedef for a function for configuring custom
// behavior of the authtoken middleware
type ConfigurationFunc func(*Middleware) error

// RequireUserToken returns a ConfigurationFunc that can be used to configure
// the middleware to enforce that a user token be supplied in the auth header,
// rather than a client token. Typically, endpoints that need to access
// user-specific data obtain the necessary user information from the user
// token, and cannot function properly if a client token in supplied instead.
// In such a configuration the middleware will return a 400 if a non-user token
// is found on the request.
func RequireUserToken() ConfigurationFunc {
	return func(m *Middleware) error {
		m.requiresUserToken = true
		return nil
	}
}

// RequirePermissions returns a ConfigurationFunc that can be used to configure
// the middleware to enforce that a token be supplied that contains the specified
// set of permissions.  In such a configuration the middleware will return a
// 403 if a token is found on the request that does not contain the required permissions.
func RequirePermissions(permissions []int) ConfigurationFunc {
	return func(m *Middleware) error {
		m.requiredPermissions = permissions
		return nil
	}
}

// Middleware is an HTTP middleware that checks for the presence of an
// encoded token value in the "Authorization" header.
//
// If no header value is found, a 401 is returned.
//
// If a value is found, the middleware uses the provided token decoder to
// decode the value of the authorization header, returning a 401
// response if any errors occur while decoding.
//
// If the middleware is successful, it decodes the token and stores it in the
// context of the request.
type Middleware struct {
	decoder TokenDecoder
	next    http.Handler

	// requiresUserToken indicates that the middleware will require that a user
	// token be supplied in the auth header, rather than a client token.
	requiresUserToken bool

	// requiredPermissions stores any required permissions values that a token
	// must have in order to be allowed by the middleware.
	requiredPermissions []int
}

// NewMiddleware creates an initializes an instance of
// Middleware. Any number of ConfigurationFuncs can be provided to
// customize the behavior of the middleware.
func NewMiddleware(decoder TokenDecoder, next http.Handler, options ...ConfigurationFunc) (http.Handler, error) {

	m := &Middleware{
		decoder: decoder,
		next:    next,
	}

	for _, opt := range options {
		// apply the option
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// ServeHTTP allows the middleware to implement the http.Handler interface.
// When called, all of the logic and validation described in the Middleware
// documentation is performed.  If any token checks/assertsions fail, the
// "next" HTTP handler in the middleware chain will NOT be called, and an error
// response will be returned to the caller, along with the proper 4xx HTTP
// status code.
func (m *Middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	headerValue := req.Header.Get(AuthorizationHeaderName)
	if headerValue == "" {
		// No header present at all! Return a 401 and don't call any other
		// middleware; We're done here.
		http.Error(rw, "missing Authorization header", http.StatusUnauthorized)
		return
	}

	// ensure that the header is using the proper bearer strategy
	if !strings.HasPrefix(headerValue, BearerPrefix) {
		http.Error(rw, "", http.StatusUnauthorized)
		return
	}

	tokenValue := strings.TrimPrefix(headerValue, BearerPrefix)
	if tokenValue == "" {
		http.Error(rw, "token is empty", http.StatusUnauthorized)
		return
	}

	token, err := m.decoder.Decode(tokenValue)
	if err != nil {
		// invalid token! Rather than expose any details about why the token
		// couldn't be decoded, just return a 401.
		http.Error(rw, "", http.StatusUnauthorized)
		return
	}

	// the token was decoded successfully
	var newReq *http.Request

	// check to see if a user token is required/supplied
	if m.requiresUserToken && len(token.UserID()) == 0 {
		http.Error(rw, "requires a user token", http.StatusBadRequest)
		return
	}

	// check to see if the token is expired
	if token.IsExpired() {
		http.Error(rw, "", http.StatusUnauthorized)
		return
	}

	// check to see if any required permissions are present
	if len(m.requiredPermissions) > 0 && !token.HasAllPermissions(m.requiredPermissions) {
		http.Error(rw, "The token provided does not allow this operation", http.StatusForbidden)
		return
	}

	// success; add the token to the request's context
	newReq = AddTokenToContext(req, token)

	// call the next middleware
	m.next.ServeHTTP(rw, newReq)
}

// AddTokenToContext adds the `tokener` to the request's context and returns the
// newly-amended request
func AddTokenToContext(req *http.Request, token tokens.Tokener) *http.Request {
	newCtx := context.WithValue(req.Context(), contextdefs.UserToken, token)
	return req.WithContext(newCtx)
}

// GetTokenFromContext is the companion function to AddTokenToContext. It
// retrieves the tokens.Tokener interface from the request's context (new to Go
// 1.7+) and performs the necessary type assertion. This function returns nil
// if no token was found in the context.
func GetTokenFromContext(ctx context.Context) tokens.Tokener {
	val := ctx.Value(contextdefs.UserToken)
	if val == nil {
		return nil
	}

	t, ok := val.(tokens.Tokener)

	if ok == false {
		return nil
	}

	return t
}
