package request

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/sony/gobreaker"
)

var (

	// ErrCircuitOpen is returned by the transport when the downstream is
	// unavailable due to a broken circuit.
	ErrCircuitOpen = errors.New("circuit open")

	// DefaultRequestTimeout is a sensible default for a http.Client Timeout
	DefaultRequestTimeout = 3 * time.Second

	// DefaultRequestTransport is a sensible default for a http.Client Transport
	DefaultRequestTransport = &http.DefaultTransport

	// defaultBreakerOpts represents sensible defaults for a circuit
	// breaker
	defaultBreakerOpts = struct {
		Window            time.Duration
		MinObservations   int
		FailurePercentage int
	}{
		Window:            5 * time.Second,
		MinObservations:   10,
		FailurePercentage: 50,
	}
)

// CircuitBreaker returns a configuration function that sets up the
// behavior of the HTTP client's circuit breaker.
//
// `window` is the amount of time that the breaker will keep stats/metrics for,
// when deciding to open/close.
//
// `minObservations` is the minimum number of requests that have to be made in
// a single window before the breaker can trip. This prevents, say, a single
// request failure from opening the breaker, since 1 failure out of 1 total ==
// 100% failure!
//
// `failurePercentage is the % of requests (from 0 to 100) that have to fail
// in any given window before the breaker will open.
func CircuitBreaker(window time.Duration, minObservations, failurePercentage int) ConfigurationFunc {

	return func(opts *clientOpts) error {
		// validate the failure % argument
		if failurePercentage <= 0 || failurePercentage > 100 {
			return errors.New("failurePercentage must be in the range (0, 100]")
		}

		// readyToTrip will tell the circuit breaker to trip iff:
		// - the minimum # of requests have been made AND
		// - the failure rate is >= the configured failure rate
		readyToTrip := func(counts gobreaker.Counts) bool {
			if counts.Requests == 0 {
				return false
			}

			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			failureThreshold := float64(failurePercentage) / 100.0

			return counts.Requests >= uint32(minObservations) &&
				failureRatio >= (failureThreshold)
		}

		st := gobreaker.Settings{
			// use an empty name, so that we can (shudder) check the error message returned
			// by the underlying gobreaker package, and return a custom error instead
			Name:        "",
			Interval:    window,
			ReadyToTrip: readyToTrip,
		}

		opts.breaker = gobreaker.NewCircuitBreaker(st)
		return nil
	}
}

// Timeout returns a configuration function that sets up the
// request timeout to use for each request made by the client.
//
// From (https://golang.org/pkg/net/http/#Client):
//   Timeout specifies a time limit for requests made by this Client. The
//   timeout includes connection time, any redirects, and reading the
//   response body. A Timeout of zero means no timeout.
func Timeout(timeout time.Duration) ConfigurationFunc {
	return func(opts *clientOpts) error {
		opts.requestTimeout = timeout
		return nil
	}
}

// Transport returns a configuration function that sets the transport
// used by the client. The transport allows additional tuning such as
// MaxIdleConns, MaxIdleConnsPerHost, and  IdleConnTimeout.
// See https://golang.org/pkg/net/http/#Transport for full details.
func Transport(transport http.RoundTripper) ConfigurationFunc {
	return func(opts *clientOpts) error {
		opts.transport = transport
		return nil
	}
}

// clientOpts encapsulates all the various configuration options that can be
// modified by a user of the request package. Instances are only modified by
// the various ConfigurationFuncs, and used by the NewClient "constructor" to
// create the actual http.Client.
type clientOpts struct {
	responseValidator ResponseValidator
	breaker           *gobreaker.CircuitBreaker
	requestTimeout    time.Duration
	transport         http.RoundTripper
}

// A ConfigurationFunc is a typedef for the return values of the various
// functions that configure some aspect of a new HTTP client.
type ConfigurationFunc func(*clientOpts) error

// NewClient is a constructor function that creates a new HTTP client for
// interacting with various services. It accepts 0 or more configuration
// functions, to customize the behavior of the returned client.
//
// If no configuration functions are used, the client defaults to using
// a 3 second timeout for all requests, and a circuit breaker that opens/trips
// if a 50% failure rate is observed over a 5 second window, with a minimum
// of 10 requests.
func NewClient(configFuncs ...ConfigurationFunc) (*http.Client, error) {
	var err error

	// start with the defaults
	opts := &clientOpts{
		requestTimeout:    DefaultRequestTimeout,
		responseValidator: DefaultResponseValidator,
		transport:         *DefaultRequestTransport,
	}

	// apply the default circuit breaker
	breakerFunc := CircuitBreaker(defaultBreakerOpts.Window, defaultBreakerOpts.MinObservations, defaultBreakerOpts.FailurePercentage)
	if err = breakerFunc(opts); err != nil {
		return nil, err
	}

	// apply each configuration function that was supplied
	for _, configFunc := range configFuncs {
		if err := configFunc(opts); err != nil {
			return nil, err
		}
	}

	// Construct an http.Client, using our "layered" transports (the circuit
	// breaker-protected transport uses the transport in the options to make the
	// actual requests), and an overall request timeout. For more details and
	// possiblities, see the "Client Timeouts" section of this reference:
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	//
	// From (https://golang.org/pkg/net/http/#Client):
	//   Timeout specifies a time limit for requests made by this Client. The
	//   timeout includes connection time, any redirects, and reading the
	//   response body. A Timeout of zero means no timeout.
	return &http.Client{
		Transport: wrapTransport(opts.breaker, opts.responseValidator, &opts.transport),
		Timeout:   opts.requestTimeout,
	}, nil
}

// ResponseValidationError is the error that is returned when a response is
// judged to be invalid by a RespsonseValidator function.
type ResponseValidationError struct{}

func (rve ResponseValidationError) Error() string {
	return "response did not pass validation"
}

// ResponseValidator is a function that determines if an http.Response
// received by a circuit breaking Transport should count as a success or a
// failure. The DefaultResponseValidator can be used in most situations.
type ResponseValidator func(*http.Response) bool

// DefaultResponseValidator considers any status code less than 500 to be a
// success, from the perspective of a client. All other codes are failures.
func DefaultResponseValidator(resp *http.Response) bool {
	return resp.StatusCode < 500
}

// IsCircuitOpenError helps callers identify whether or not an error they
// receive from using the HTTP client is a result of the ciruit breaker being
// open, or something else entirely.
func IsCircuitOpenError(err error) bool {
	if err == nil {
		return false
	}
	// This is necessary since the Transport wraps the returned errors from
	// RoundTrip in a url.Error type. We have to type assert it and check
	// the underlying .Err field
	if urlError, ok := err.(*url.Error); ok {
		return urlError.Err == ErrCircuitOpen
	}
	return false
}

// IsTimeoutError helps callers identify whether or not an error they
// receive from using the HTTP client is a result of a request timeout,
// or something else entirely.
// NOTE: this only works on Go 1.6+!!
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}

// wrapTransport produces an http.RoundTripper that's governed by the passed
// Breaker and ResponseValidator. Responses that fail the validator signal
// failures to the breaker. Once the breaker opens, outgoing requests are
// terminated before being forwarded with ErrCircuitOpen.
func wrapTransport(breaker *gobreaker.CircuitBreaker, validator ResponseValidator, next *http.RoundTripper) http.RoundTripper {
	return &transport{
		breaker:   breaker,
		validator: validator,
		next:      *next,
	}
}

type transport struct {
	breaker   *gobreaker.CircuitBreaker
	validator ResponseValidator
	next      http.RoundTripper
}

// RoundTrip implements the http.RoundTripper interface, and allows the
// `transport` type to be used as the transport for an http.Client.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {

	response, err := t.breaker.Execute(func() (interface{}, error) {
		// execute the request using the "nested" transport
		resp, err := t.next.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// Check that the response passes the supplied "validator" function
		if !t.validator(resp) {
			return resp, ResponseValidationError{}
		}
		return resp, nil
	})

	if err == nil {
		return response.(*http.Response), nil
	}

	switch err.(type) {
	case ResponseValidationError:
		return response.(*http.Response), nil
	default:
		// TODO: gross.  If the breaker package would return a custom error type
		// here, we wouldn't have to do this
		if err.Error() == "circuit breaker is open" {
			return nil, ErrCircuitOpen
		}
		return nil, err
	}

}
