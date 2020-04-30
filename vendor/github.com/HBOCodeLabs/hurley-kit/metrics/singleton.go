package metrics

import (
	"sync/atomic"
	"unsafe"
)

// Singleton the singleton instance of Collector, must be initialized with `metrics.Init`
var singleton unsafe.Pointer

// Init initialize the global collector instance. There are two ways to use metrics collector. You can create instances of `Collector`s
// with `metrics.NewCollector` function, which creates a collector with specific settings your provided. In many cases, you may want to
// use a global/singleton instance of collector without passing the instance around. Before you can use this global instance. You will
// need to initialize it with your settings.
func Init(opts *Config) error {
	// emit statsd metrics
	var err error
	var c *Collector
	if opts.Buffered == true {
		c, err = NewCollector(WithHostPort(opts.Host, opts.Port), WithPrefix(opts.Prefix), WithBuffering(opts.Interval))
	} else {
		c, err = NewCollector(WithHostPort(opts.Host, opts.Port), WithPrefix(opts.Prefix))
	}
	atomic.StorePointer(&singleton, unsafe.Pointer(c))

	return err
}

// Timing global helper function that uses the `Singleton` `Collector` to adds timing metrics
func Timing(name string, val int64) error {
	c := (*Collector)(atomic.LoadPointer(&singleton))
	if c == nil {
		return nil
	}

	return c.Timing(name, val)
}

// Incr global helper function that uses the `Singleton` `Collector` to adds counter metrics
func Incr(name string, val int64) error {
	c := (*Collector)(atomic.LoadPointer(&singleton))
	if c == nil {
		return nil
	}

	return c.Incr(name, val)
}

// Gauge global helper function that uses the `Singleton` `Collector` to adds gauage metrics
func Gauge(name string, val int64) error {
	c := (*Collector)(atomic.LoadPointer(&singleton))
	if c == nil {
		return nil
	}

	return c.Gauge(name, val)
}
