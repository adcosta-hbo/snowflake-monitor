package tracing

import (
	"context"
	"errors"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/uber/jaeger-client-go"

	"github.com/HBOCodeLabs/hurley-kit/contextdefs"
	"github.com/HBOCodeLabs/hurley-kit/tracing/defs"
)

const (
	// DefaultAgentPort is the default port number used to send tracing data to the
	// Jaeger agent. The agent and other Jaeger components are described in detail
	// here: https://jaeger.readthedocs.io/en/latest/architecture/
	DefaultAgentPort int = 6381

	// DefaultSampleRate is the sample rate used by each client (unless configured
	// otherwise) to decide how much tracing data to send to the agent.
	DefaultSampleRate float64 = 0.01
)

// InfoFromContext extracts and returns the traceID, spanID, and
// parentSpanID respectively, from the supplied context. Any of these values
// that don't exist in the context will be represented as empty strings.
func InfoFromContext(ctx context.Context) (string, string, string, string) {
	// We use longer form of type assertion (with two return values), to prevent
	// runtime panics.
	// We can ignore the result (using the underscore) because the language spec
	// states that an invalid type assertion will result in the default value for
	// type being assigned, with in this case is what we want (the empty string).
	traceID, _ := ctx.Value(contextdefs.TraceID).(string)
	spanID, _ := ctx.Value(contextdefs.SpanID).(string)
	parentSpanID, _ := ctx.Value(contextdefs.ParentSpanID).(string)
	uberTrace, _ := ctx.Value(contextdefs.UberHeader).(string)

	return traceID, spanID, parentSpanID, uberTrace
}

// ConfigurationFunc is a typedef for a function for configuring custom
// behavior of the middleware
type ConfigurationFunc func(*Middleware) error

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
	tracer opentracing.Tracer
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
	if m.tracer == nil {
		return nil, errors.New("A tracer is required in order to create the middleware")
	}

	return m, nil
}

// resumeOrStartNewSpan extracts tracing information from the supplied HTTP
// request (if such information exists) and creates a new Span/Child span.  If no
// such information is found, a new span is created with no parent. In either
// case, the Span object is returned.
func (m *Middleware) resumeOrStartNewSpan(r *http.Request) opentracing.Span {
	var span opentracing.Span

	spanContext, err := m.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	if err == nil {
		// create a span that "resumes" the trace sent in the HTTP headers
		span = m.tracer.StartSpan(defs.HTTPServer, ext.RPCServerOption(spanContext))
	} else {
		// start a brand-new span
		span = m.tracer.StartSpan(defs.HTTPServer)

		switch err {
		// Per the opentracing godoc:
		//   > If there was simply no SpanContext to extract in
		//   > `carrier`, Extract() returns (nil, opentracing.ErrSpanContextNotFound)
		//
		// So this error is "normal" when there's no incoming span data. However,
		// we may want to log for instances of any other error.
		case opentracing.ErrSpanContextNotFound:
		default:
			// TODO: log?  ignore?
		}
	}

	return span
}

// ServeHTTP allows the middleware to implement the http.Handler interface.
// When called, all of the logic and validation described in the Middleware
// documentation is performed.
func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	span := m.resumeOrStartNewSpan(r)
	ctx := opentracing.ContextWithSpan(r.Context(), span)

	span.SetTag(string(ext.HTTPUrl), r.URL.String())
	span.SetTag(string(ext.HTTPMethod), r.Method)

	// Ensure that the span is "finished", to guarantee that its data gets collected
	defer span.Finish()

	existingTraceID := r.Header.Get(defs.TraceIDHeaderName)
	spanContext, isJaegerSpanContext := span.Context().(jaeger.SpanContext)

	// Adds the span itself, as well as the trace/span IDs to the context, so
	// they can be obtained/inspected by other middleware/handlers
	// Uber TraceID headers are used for all opentracing reports.
	// Order of preference is as follows:
	//    A.)  Incoming X-B3 TraceIDs are used for traceID, `uber-trace-id` is used for parent and span.
	//    B.)  `uber-trace-id` header containing trace, span, and parent ids is used
	//    C.)  random trace, span, and parent ids are generated
	if isJaegerSpanContext {
		// Opentracing is enabled
		if existingTraceID != "" {
			w.Header().Set(defs.TraceIDHeaderName, existingTraceID)
		} else {
			w.Header().Set(defs.TraceIDHeaderName, spanContext.TraceID().String())
		}
		w.Header().Set(defs.SpanIDHeaderName, spanContext.SpanID().String())
		w.Header().Set(defs.ParentSpanIDHeaderName, spanContext.ParentID().String())
		w.Header().Set(defs.UberOpentracingHeaderName, spanContext.String())

	} else {
		// OpenTracing is not enabled
		if existingTraceID != "" {

			w.Header().Set(defs.TraceIDHeaderName, existingTraceID)

			// spanID logic
			if existingSpanID := r.Header.Get(defs.SpanIDHeaderName); existingSpanID != "" {
				// copy the existing spanID header into the response
				w.Header().Set(defs.SpanIDHeaderName, existingSpanID)
			} else {
				// generate a new spanID
				w.Header().Set(defs.SpanIDHeaderName, NewTraceID())
			}

			// parentID logic
			if existingParentID := r.Header.Get(defs.ParentSpanIDHeaderName); existingParentID != "" {
				// copy the existing parentID header into the response
				w.Header().Set(defs.ParentSpanIDHeaderName, existingParentID)
			} else {
				// re-use the traceID as the parentID
				w.Header().Set(defs.ParentSpanIDHeaderName, existingTraceID)
			}

		} else {
			randomID := NewTraceID()
			w.Header().Set(defs.TraceIDHeaderName, randomID)
			w.Header().Set(defs.SpanIDHeaderName, randomID)
			w.Header().Set(defs.ParentSpanIDHeaderName, randomID)
		}
	}

	// Populate the ctx for logging, and further tracing.
	ctx = context.WithValue(ctx, contextdefs.TraceID, w.Header().Get(defs.TraceIDHeaderName))
	ctx = context.WithValue(ctx, contextdefs.SpanID, w.Header().Get(defs.SpanIDHeaderName))
	ctx = context.WithValue(ctx, contextdefs.ParentSpanID, w.Header().Get(defs.ParentSpanIDHeaderName))
	if isJaegerSpanContext {
		ctx = context.WithValue(ctx, contextdefs.UberHeader, spanContext.String())
	}

	// Wrap the response writer in our own type, allowing for capture of the
	// status code used in the response. We default the response code to 200
	// because of this behavior of the http.ResponseWriter:
	//
	//   > If WriteHeader has not yet been called, Write calls
	//   > WriteHeader(http.StatusOK) before writing the data.
	//   (https://golang.org/pkg/net/http/#ResponseWriter)
	//
	// Most (all?) implementations of http.ResponseWriter in the std lib call
	// an unexported method to write the 200 header, rather than making a call
	// to our .WriteHeader method.
	rw := &captureResponseCode{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	// Call the next handler in the request/response "chain"
	m.next.ServeHTTP(rw, r.WithContext(ctx))

	// finally, capture the response code that was sent by other HTTP handlers, and add it to the span
	span.SetTag(string(ext.HTTPStatusCode), rw.statusCode)

	// Inject the tracing data into the HTTP headers of the outgoing response.
	// Since we're using the Zipkin propagation scheme, this will use the
	// familiar X-B3-* headers that we're used to seeing.
	m.tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(rw.Header()))
}

// captureResponseCode is a type that embeds an http.ResponseWriter, in order
// to capture the response code sent by an http.Handler that is "further down
// the chain" from this middleware.
type captureResponseCode struct {
	http.ResponseWriter
	statusCode int
}

// Write is part of the http.ResponseWriter interface, which propagates calls
// to the underlying http.ResponseWriter.
func (r *captureResponseCode) Write(p []byte) (int, error) {
	return r.ResponseWriter.Write(p)
}

// WriteHeader is part of the http.ResponseWriter interface, and allows the
// captureResponseCode instance to capture the status code of the response.
func (r *captureResponseCode) WriteHeader(status int) {
	r.statusCode = status
	r.ResponseWriter.WriteHeader(status)
}

// CloseNotify is part of the http.CloseNotifier interface (implemented by
// http.ResponseWriter), which propagates calls to the underlying response
// writer, which allows detecting when the underlying connection has gone away.
func (r *captureResponseCode) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
