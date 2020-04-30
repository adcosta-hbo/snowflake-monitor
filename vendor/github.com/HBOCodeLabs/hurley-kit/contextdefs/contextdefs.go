// Package contextdefs contains definitions of shared/common context-related keys and functions
package contextdefs

// contextKey is a typedef that allows us to define keys to use when
// storing/retrieving values in a context object. This type is used instead of
// a string, because:
//
// (via bradfitz): You really don't want to use strings as the keys. That means
// there's no isolated namespaces between packages, so different packages can
// collide and use the same keys, since everybody has access to the string
// type. But by requiring people to use their own types, there can't be
// conflicts.
type contextKey string

const (
	// UserToken is the key used to store/retrieve a user's decoded and
	// unmarshalled Hurley OAuth token to/from the context
	UserToken = contextKey("usertoken")

	// TraceID is the key used to store/retrieve the ID of a request trace
	// to/from the context
	TraceID = contextKey("X-B3-Traceid")

	// SpanID is the key used to store/retrieve the ID of a request tracing span
	// to/from the context
	SpanID = contextKey("X-B3-Spanid")

	// ParentSpanID is the key used to store/retrieve the ID of a request trace's
	// parent span to/from the context
	ParentSpanID = contextKey("X-B3-Parentspanid")

	// ForwardedFor is the key used to store/retrieve the originating client IP
	// address for a request that has gone through an HTTP proxy or load balancer
	// (including Amazon ELBs).
	ForwardedFor = contextKey("X-Forwarded-For")

	// UserInfo is the key used to store/retrieve a blob of data about a Hurley
	// user to/from the context.  NOTE that this method of propagating user
	// information between services (using a special `X-Userinfo` HTTP) should be
	// considered deprecated. Services should prefer to propagate the
	// encoded auth token in the typical "Authorization" HTTP header instead.
	UserInfo = contextKey("X-Userinfo")

	// HBOCaller is the key used to store/retrieve the name of a service making
	// an RPC call to/from the context.  This data is has typically (but not
	// always) been propagated between Hurley services using the "X-Hbo-Caller"
	// HTTP header.
	HBOCaller = contextKey("X-Hbo-Caller")

	// UberHeader is the key used to store/retrieve opentracing information
	UberHeader = contextKey("uber-trace-id")

	// SignedSignature is the key used to store/retrieve the signed signature header
	SignedSignature = contextKey("signature")

	// PlatformTenantCode is the key used to store/retrieve the PlatformTenantCode
	PlatformTenantCode = contextKey("PlatformTenantCode")
)
