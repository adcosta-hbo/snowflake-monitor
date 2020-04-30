package tracing

import (
	"io"

	"github.com/opentracing/opentracing-go"
)

// CreateTracer creates an opentracing tracer, using the configured options. In
// a slight break with typical GO conventions, this function can return a
// non-nil tracer instance AND a non-nil error from the same invocation. In this case,
// the tracer implementation is a noop implementation.  This allows consuming code
// to operate as usual, without having to perform myriad nil reference checks before
// using the tracer.
func CreateTracer(serviceName string, opts Options) (opentracing.Tracer, io.Closer, error) {
	if opts.Disabled {
		// Just return no-op implementations of each interface.  This allows the
		// caller to use each as per usual, without having to keep checking whether
		// or not either of them is nil.
		return opentracing.NoopTracer{}, noopCloser{}, nil
	}

	fns := []TracerConfigurationFunc{
		ServiceName(serviceName),
		SamplerType(opts.Sampler.Type),
	}

	if opts.Reporter.Host != "" {
		fns = append(fns, AgentAddress(opts.Reporter.Host, opts.Reporter.Port))
	}

	if opts.Sampler.Param != 0.00 {
		fns = append(fns, SampleRate(opts.Sampler.Param))
	}

	tracer, closer, err := NewTracer(fns...)
	if err != nil {
		return opentracing.NoopTracer{}, noopCloser{}, err
	}

	return tracer, closer, nil
}

// A noopCloser is a trivial, minimum overhead implementation of io.Closer
// for which the Close operation is a no-op.
type noopCloser struct{}

func (noopCloser) Close() error { return nil }
