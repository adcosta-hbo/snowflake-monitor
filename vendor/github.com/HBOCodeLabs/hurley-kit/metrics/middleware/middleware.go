package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/HBOCodeLabs/hurley-kit/metrics"
)

type codeCapture struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader implements (a portion of) http.ResponseWriter, and is used to
// capture the status code being written to the response
func (c *codeCapture) WriteHeader(status int) {
	c.statusCode = status
	c.ResponseWriter.WriteHeader(status)
}

// CloseNotify propagates call to the CloseNotify
// implemented by ResponseWriters which allow detecting when the underlying connection has gone away.
func (c *codeCapture) CloseNotify() <-chan bool {
	return c.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// NewHTTPLatencyMiddleware creates a http.Handler that delegates request handling to
// the given http.Handler and records the duration of the delegate's ServeHTTP
// method as a statsd 'timing' metric. The name of the timing metric follows the convention:
// 	`http.[<metricBase>].<fn>[.<fn2>... .<fnN>].<code>.elapsed`
// where `metricBase` is an optional string argument, `code` is the HTTP
// response code, and `fn1` - `fnN` are the return values from the specified
// `fns`.
//
// The middleware also emits a counter metric that follows the same naming
// convention, with the omission of the ".elapsed" suffix.
//
// Typically, the "metricBase" argument would be the name of some endpoint in
// the service/server.  For example, for a "GET /data" HTTP endpoint, the
// metricBase argument would be "get.data", yielding final metric names like:
//
//    http.get.data.200         (counter)
//    http.get.data.200.elapsed (timer)
//
// Since the metric names emitted by the middleware begin with "http", a client
// with a configured prefix should be used in order to prevent cluttering the
// "global" statsd namespace.  In the previous example, if the supplied client
// had a prefix of "foobar", the metrics would ultimately show up in statsd as:
//
//    foobar.http.get.data.200         (counter)
//    foobar.http.get.data.200.elapsed (timer)
func NewHTTPLatencyMiddleware(s metrics.Statsder, metricBase string, h http.Handler, fns ...func(*http.Request) string) http.Handler {
	var partialMetricName string

	if metricBase != "" {
		partialMetricName = fmt.Sprintf("http.%s", metricBase)
	} else {
		partialMetricName = "http"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		cc := &codeCapture{ResponseWriter: w, statusCode: http.StatusOK}
		h.ServeHTTP(cc, r)
		d := time.Now().Sub(start)

		// Build the full metric name, using the return values from all the
		// supplied fns
		components := []string{partialMetricName}
		for _, fn := range fns {
			if rv := fn(r); rv != "" {
				components = append(components, rv)
			}
		}
		components = append(components, strconv.Itoa(cc.statusCode))

		counterMetricName := strings.Join(components, ".")
		timerMetricName := counterMetricName + ".elapsed"

		ms := d.Nanoseconds() / int64(1e6)
		s.Timing(timerMetricName, ms)
		s.Incr(counterMetricName, 1)
	})
}

// NewHTTPHeaderMiddleware creates a http.Handler that delegates request
// handling to the given http.Handler and increments counters based on names
// and values of headers which contain the prefix "X-Increment-". The name of
// the metric follows the convention:
//  `<name>.<header>`
// where `c` is the supplied StatsD collector, `name` is the name supplied to this middleware,
// and `header` is the portion of the header name after the 'X-Increment-' prefix. If the header
// specifies multiple values, they will be aggregated and sent in a single metric.
func NewHTTPHeaderMiddleware(s metrics.Statsder, name string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		findAndLogIncrements(s, name, w)
	})
}

func findAndLogIncrements(s metrics.Statsder, name string, w http.ResponseWriter) {
	for k, vv := range w.Header() {
		if strings.HasPrefix(k, "X-Increment-") {
			var value int64
			for _, v := range vv {
				if val, err := strconv.ParseInt(v, 10, 64); err == nil {
					value += int64(val)
				}
			}
			s.Incr(name+k[len("X-Increment"):], value)
		}
	}
}
