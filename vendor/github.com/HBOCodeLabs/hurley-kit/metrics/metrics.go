package metrics

import (
	"encoding/json"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/HBOCodeLabs/statsd"
)

// Statsder is an interface that covers the basic operations that a statsd
// client must support. In practice, we use a third-party statsd client that
// supports these operations. However for tests, we use a "mock" implementation
// that supports these same operations.
type Statsder interface {
	Timing(string, int64) error
	Incr(string, int64) error
	Gauge(string, int64) error
	io.Closer
}

// Collector is a type for abstracting the sending of captured API metrics to
// statsd
type Collector struct {
	Statsder
}

// Config represents the set of options for configuring the metrics middleware
type Config struct {
	Host     string
	Port     int
	Prefix   string
	Buffered bool
	Interval time.Duration `json:"interval"`

	// ReconnectInterval is the time interval with which to recreate the
	// underlying socket used to send metrics. This can help ensure that metrics
	// continue to be gathered from an application in the event that the metrics
	// collector is unavailable, without an application restart.
	ReconnectInterval time.Duration `json:"reconnectInterval"`
}

// MarshalJSON is part of the interface that allows our type to implement custom
// JSON marshalling (mainly for the "Interval" field, which is represented as a
// time.Duration type (which gets converted to a string for JSON).
func (c Config) MarshalJSON() ([]byte, error) {
	type Alias Config

	return json.Marshal(&struct {
		Interval          string `json:"interval"`
		ReconnectInterval string `json:"reconnectInterval"`
		*Alias
	}{
		Interval:          c.Interval.String(),
		ReconnectInterval: c.ReconnectInterval.String(),
		Alias:             (*Alias)(&c),
	})
}

// UnmarshalJSON is part of the interface that allows our type to implment
// custom JSON unmarshalling (mainly for the "Interval" field, which is
// represented as a time.Duration type (which gets parsed from a string when
// coming from JSON).
func (c *Config) UnmarshalJSON(data []byte) error {
	type Alias Config
	var err error

	aux := &struct {
		Interval          string `json:"interval"`
		ReconnectInterval string `json:"reconnectInterval"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Do our "custom" conversion from a plain string to a duration
	if c.Interval, err = time.ParseDuration(aux.Interval); err != nil {
		return err
	}

	if c.ReconnectInterval, err = time.ParseDuration(aux.ReconnectInterval); err != nil {
		return err
	}
	return nil
}

// NewCollector creates a Collector instance. Configuration functions may be
// supplied to override the default settings of the statsd client.
// The default settings are as follows:
//    Host:       localhost
//    Port:				8125
//		Buffered:   false
//
// In a slight break with typical GO conventions, this function can return a
// non-nil Collector instance AND a non-nil error from the same invocation.
// This can happen for instance, if the metrics collector was successfully
// created, but could not establish a socket connection to its underlying
// destination.  This leaves the caller with a choice of whether to continue
// using the client (and potentially call Reconnect() or let the automatic
// reconnects happen), or give up entirely.
func NewCollector(fns ...func(*Config) error) (*Collector, error) {

	cfg := &Config{
		Host:     "localhost",
		Port:     8125,
		Buffered: false,
		Interval: time.Second * 10,
	}

	// run each configuration function to arrive at the final config object
	var err error
	for _, fn := range fns {
		if err == nil {
			err = fn(cfg)
		}
	}

	if err != nil {
		return nil, err
	}

	var collector *Collector

	c := statsd.NewStatsdClient(
		net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		cfg.Prefix,
		statsd.WithReconnectInterval(cfg.ReconnectInterval),
	)

	// We still need to return a non-nil client to the caller here, even though
	// there was an error creating the socket. Since the client has periodic
	// reconnects, the underlying socket creation error might be resolved on
	// subsequent connection attempts. Returning a non-nil client in this case
	// allows the consuming application to continue using the client in the hopes
	// that it eventually reconnects.
	err = c.CreateSocket()

	if cfg.Buffered {
		stats := statsd.NewStatsdBuffer(cfg.Interval, c)
		collector = &Collector{
			Statsder: stats,
		}
		return collector, err
	}

	collector = &Collector{
		Statsder: c,
	}

	return collector, err
}

// WithHostPort returns a configuration function that sets the statsd host and port.
func WithHostPort(host string, port int) func(*Config) error {
	return func(c *Config) error {
		c.Host = host
		c.Port = port
		return nil
	}
}

// WithPrefix returns a configuration function that sets a prefix to add to the
// names of all the emitted metrics.
func WithPrefix(prefix string) func(*Config) error {
	return func(c *Config) error {
		c.Prefix = prefix
		return nil
	}
}

// WithBuffering returns a configuration function that sets the interval with
// which to send buffered metrics updates to statsd.
func WithBuffering(interval time.Duration) func(*Config) error {
	return func(c *Config) error {
		c.Buffered = true
		c.Interval = interval
		return nil
	}
}

// WithReconnectInterval returns a configuration function that causes the
// metrics collector to automatically recreate its underlying connection on the
// specified interval.
func WithReconnectInterval(interval time.Duration) func(*Config) error {
	return func(c *Config) error {
		c.ReconnectInterval = interval
		return nil
	}
}

// Stop makes the collector stop sending metrics.  Any buffered metrics will be flushed.
func (c *Collector) Stop() {
	c.Close()
}
