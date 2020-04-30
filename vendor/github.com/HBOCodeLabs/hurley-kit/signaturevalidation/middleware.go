package signaturevalidation

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ConfigurationFunc is a typedef for a function for configuring custom
// behavior of the middleware
type ConfigurationFunc func(*Middleware) error

// WithSecret returns a ConfigurationFunc that is used to configure the
// middleware with a secret to validating requests.  Since this middleware accepts a single secret,
// it is likely that it will be used to validate requests to a
// single HTTP endpoint.
func WithSecret(secret string) ConfigurationFunc {
	return func(m *Middleware) error {
		if secret == "" {
			return errors.New("a secret is required to validate request signature")
		}
		m.secret = secret
		return nil
	}
}

// Middleware is an HTTP handler that creates a new opentracing-compatible
// Tracer which is used to capture distributed tracing data and send it to our
// Jaeger collectors.
//
// When an incoming HTTP request is encountered, a new Span is created, and a
// reference to it is stored in the request's context.
//
// Currently, the following tags are automatically added to each Span:
//   - HTTP method
//   - HTTP URL
//   - HTTP status code
//
// If the incoming HTTP request contains Zipkin-style "X-B3-*" headers with
// tracing/span data in them, the new Span will be created as a "child" of that
// existing span.
//
// For backwards-compatibility and interoperability with our other systems, the
// traceID and spanID values are added to headers on the outgoing response, and
// added to the request's context, for use by other middleware/handlers.
type Middleware struct {
	next   http.Handler
	secret string
}

// NewMiddleware creates an instance of Middleware that calls `next` after
// completing. Any number of ConfigurationFuncs can be provided to customize
// the behavior of the middleware.
func NewMiddleware(next http.Handler, options ...ConfigurationFunc) (http.Handler, error) {
	// set up some defaults
	m := &Middleware{
		next: next,
	}

	for _, opt := range options {
		// apply the options
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	// check for required configuration before proceeding
	if m.secret == "" {
		return nil, errors.New("A secret is required in order to validate request signature")
	}

	return m, nil
}

// ServeHTTP allows the middleware to implement the http.Handler interface.
// When called, all of the logic and validation described in the Middleware
// documentation is performed.
func (m *Middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	err := validateSignatureHeader(m.secret, req)
	if err != nil {
		handleValidationError(err.Error(), rw)
	}
	m.next.ServeHTTP(rw, req)
}

// handleValidationError creates a JSON formatted error response containing the desired
// `message`. The format of this message was copied from the existing
// concierge API.
func handleValidationError(message string, rw http.ResponseWriter) {
	errMsg := map[string]interface{}{
		"message": message,
	}

	byts, err := json.Marshal(errMsg)
	if err != nil {
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusForbidden)
	rw.Write(byts)
}
