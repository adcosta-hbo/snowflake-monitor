package schemavalidation

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

var (
	// ErrSchemaValidation is the string used in the "code" property of an error response
	// from the middleware, which uses the comet error schema:
	// https://comet.staging.hurley.hbo.com/developer#header-6.-errors
	ErrSchemaValidation = "schema_validation_failure"
)

// ConfigurationFunc is a typedef for a function for configuring custom
// behavior of the schema validation middleware
type ConfigurationFunc func(*Middleware) error

// UseVerboseErrors returns a ConfigurationFunc that can be used to configure
// the middleware to return the actual JSON validation error messages in the
// response, rather than a generic "whoops, something went wrong" message.
// This option should typically only be used in a non-production environment,
// to prevent callers from being exposed to internal details of the service.
func UseVerboseErrors() ConfigurationFunc {
	return func(m *Middleware) error {
		m.verboseErrors = true
		return nil
	}
}

// WithSchemaFilePath returns a ConfigurationFunc that is used to configure the
// middleware with a filepath to a specific JSON schema file to use when
// validating requests.  Since this middleware accepts a single JSON schema
// file path, it is likely that it will be used to validate requests to a
// single HTTP endpoint.
func WithSchemaFilePath(schemaFilePath string) ConfigurationFunc {
	return func(m *Middleware) error {
		if len(schemaFilePath) == 0 {
			return errors.New("a file path to a JSON schema file is required")
		}
		m.schemaFilePath = schemaFilePath
		return nil
	}
}

// Middleware is an HTTP middleware that validates a request body against a
// JSON schema definition. The middleware returns a 400 w/ a plain-text body if
// the request body's JSON does not pass validation.
type Middleware struct {
	next http.Handler

	// schemaFilePath is the file path to a JSON schema definition to load/parse
	schemaFilePath string

	// schema is the parsed JSON schema object to use when validating requests
	schema *gojsonschema.Schema

	// verboseErrors indicates whether or not we return verbose, non-generic
	// error messages from the middleware
	verboseErrors bool
}

// NewMiddleware creates an initializes an instance of Middleware. Any number
// of ConfigurationFuncs can be provided to customize the behavior of the
// middleware.
func NewMiddleware(next http.Handler, options ...ConfigurationFunc) (http.Handler, error) {

	m := &Middleware{
		next: next,
	}

	for _, opt := range options {
		// apply the option
		if err := opt(m); err != nil {
			return nil, err
		}
	}

	schema, err := gojsonschema.NewSchema(gojsonschema.NewReferenceLoader(m.schemaFilePath))
	if err != nil {
		return nil, err
	}

	m.schema = schema
	return m, nil
}

// ServeHTTP allows the middleware to implement the http.Handler interface.
// When called, all of the logic and validation described in the Middleware
// documentation is performed.  If any validations fail, the "next" HTTP
// handler in the middleware chain will NOT be called, and an error response
// will be returned to the caller, along with a 400 HTTP status code.
func (m *Middleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Since you can only read the request Body once (it's an io.Reader that
	// has state associated w/ the read), we can use the data we read as a new
	// io.Reader for the request.
	content, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(content))

	result, err := m.schema.Validate(gojsonschema.NewBytesLoader(content))

	if err != nil || !result.Valid() {
		if !m.verboseErrors {
			// return the generic "oops" message
			handleValidationError(http.StatusText(http.StatusBadRequest), rw)
			return
		}

		// start with the generic "bad request" message
		msg := bytes.NewBufferString(http.StatusText(http.StatusBadRequest))

		if result != nil && len(result.Errors()) > 0 {
			msg.WriteString(" errors: ")
			// append a verbose string listing all the validation errors
			for _, e := range result.Errors() {
				msg.WriteString(e.Field() + ": ")
				msg.WriteString(e.Description() + ", ")
			}
		}

		handleValidationError(msg.String(), rw)
		// don't call any other middleware; we're done
		return
	}

	// call the next middleware in the chain
	m.next.ServeHTTP(rw, req)
}

// validationErrorMessage is the type used to marshal JSON-formatted error
// messages from the schema validation middleware. This type conforms to
// the comet error "spec": https://comet.staging.hurley.hbo.com/developer#header-6.-errors
// This is the preferred error response format, per the #hadron team:
// https://hbo.slack.com/archives/C03GX5A6B/p1504290579000190
type validationErrorMessage struct {
	Code    string `json:"code"` // always "schema_validation" for schema validation errors
	Message string `json:"message"`
}

// handleValidationError creates a JSON formatted error response containing the desired
// `message`. The format of this message was copied from the existing
// concierge API.
func handleValidationError(message string, rw http.ResponseWriter) {
	errMsg := validationErrorMessage{
		Code:    ErrSchemaValidation,
		Message: message,
	}

	byts, err := json.Marshal(errMsg)
	if err != nil {
		http.Error(rw, "", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write(byts)
}
