package tracing

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

// TracerConfigurationFunc is a typedef for a function for configuring custom
// behavior of the tracer
type TracerConfigurationFunc func(*TracerConfig) error

// ServiceName returns a TracerConfigurationFunc that is used to set the name
// of the service. This name is present in the captured tracing data, and
// is a required configuration in order to successfully create the middleware.
func ServiceName(name string) TracerConfigurationFunc {
	return func(c *TracerConfig) error {
		if len(name) == 0 {
			return errors.New("service name cannot be empty")
		}
		c.serviceName = name
		return nil
	}
}

// AgentAddress returns a TracerConfigurationFunc that is used to set the host/port
// of the Jaeger agent used to collect the tracing data. The middleware is configured
// to use "localhost" as the host and 6381 as the port by default.
func AgentAddress(host string, port int) TracerConfigurationFunc {
	return func(c *TracerConfig) error {
		if len(host) == 0 {
			return errors.New("agent host cannot be empty")
		}

		if port < 1 || port > 65535 {
			return errors.New("port must be a in the range [1,65535]")
		}

		c.agentHostPort = net.JoinHostPort(host, strconv.Itoa(port))
		return nil
	}
}

// SampleRate returns a TracerConfigurationFunc that is used to set the sample rate
// of the middleware.  The supplied rate must be on the interval [0.0, 1.0],
// with 0.0 indicating that no traces should be collected, and 1.0 meaning that
// all traces should be collected. If no sample rate is configured, the middleware
// will default to using a sample rate of 0.01.
func SampleRate(rate float64) TracerConfigurationFunc {
	return func(c *TracerConfig) error {
		if rate < 0.0 || rate > 1.0 {
			return errors.New("sample rate must be on the interval [0.0, 1.0]")
		}
		c.sampleRate = rate
		return nil
	}
}

var supportedSamplerTypes = []string{jaeger.SamplerTypeRemote, jaeger.SamplerTypeProbabilistic, jaeger.SamplerTypeRateLimiting, jaeger.SamplerTypeConst}

func isSamplerTypeSupported(samplerType string) bool {
	for _, sampler := range supportedSamplerTypes {
		if samplerType == sampler {
			return true
		}
	}
	return false
}

// SamplerType returns a TracerConfigurationFunc that is used to set the type of
// sampler which should be
func SamplerType(samplerType string) TracerConfigurationFunc {
	return func(c *TracerConfig) error {
		if len(samplerType) == 0 {
			return errors.New("sampler type cannot be empty")
		}
		if !isSamplerTypeSupported(samplerType) {
			return fmt.Errorf("sampler type must be one of %s", supportedSamplerTypes)
		}
		c.samplerType = samplerType
		return nil
	}
}

// UseTracer returns a ConfigurationFunc that instructs the middleware to use
// an existing tracer, rather than create a new one.  This is useful for
// creating multiple, route-specific middleware stacks, without having to
// create multiple, identical tracers.
func UseTracer(tracer opentracing.Tracer) ConfigurationFunc {
	return func(m *Middleware) error {
		m.tracer = tracer
		return nil
	}
}

// TracerConfig is used by the various TracerConfigurationFunc calls to configure
// various aspects of the tracer's behavior.
type TracerConfig struct {
	serviceName   string
	agentHostPort string
	sampleRate    float64
	samplerType   string
}

// NewTracer creates a Tracer instance. Any number of TracerConfigurationFuncs
// can be provided to customize the behavior of the tracer.
//
// NOTE that this function returns 3 values: the tracer itself, an
// "io.Closer", and an error.  In non-error cases, an io.Closer will be
// returned that should be closed by the calling process during shutdown.  This
// allows any buffered tracing data to be flushed to the collector. Example:
//
//    tracer, closer, err := NewTracer(...)
//    if err != nil { /* handle error */ }
//    defer closer.Close()
func NewTracer(options ...TracerConfigurationFunc) (opentracing.Tracer, io.Closer, error) {
	var err error

	// set up some defaults
	config := &TracerConfig{
		sampleRate:    DefaultSampleRate,
		agentHostPort: net.JoinHostPort("localhost", strconv.Itoa(DefaultAgentPort)),
		samplerType:   jaeger.SamplerTypeRemote,
	}

	for _, opt := range options {
		// apply the options
		if err := opt(config); err != nil {
			return nil, nil, err
		}
	}

	// check for required config before proceeding
	if len(config.serviceName) == 0 {
		return nil, nil, errors.New("A service name is required in order to create the tracer")
	}

	// use the env variables to set up the tracers
	// The K8S architecture dictates the jaeger endpoint.
	// Try to grab the JAEGER environment variable, use the default later if not provided.
	jaegerConfiguration, err := jaegercfg.FromEnv()
	if err != nil {
		return nil, nil, err
	}

	jaegerConfiguration.ServiceName = config.serviceName

	jaegerConfiguration.Sampler.Type = getEnvValueOrDefault(jaegerConfiguration.Sampler.Type, config.samplerType).(string)
	jaegerConfiguration.Sampler.Param = getEnvValueOrDefault(jaegerConfiguration.Sampler.Param, config.sampleRate).(float64)
	jaegerConfiguration.Reporter.LocalAgentHostPort = getEnvValueOrDefault(jaegerConfiguration.Reporter.LocalAgentHostPort, config.agentHostPort).(string)

	tracer, closer, err := jaegerConfiguration.NewTracer()
	if err != nil {
		return nil, nil, err
	}

	return tracer, closer, nil
}

// getEnvValueOrDefault returns the environment variable if present, otherwise the `fallback`
func getEnvValueOrDefault(configValue, fallback interface{}) interface{} {
	switch configValue.(type) {
	case float64:
		if configValue.(float64) == 0.0 {
			return fallback
		}
	default:
		if configValue.(string) == "" {
			return fallback
		}
	}
	return configValue
}
