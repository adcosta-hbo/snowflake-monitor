# gometrics

A Golang libray StatsD client and HTTP middleware that collects RESTful API metrics.

# Collector

`Statsder` is an interface that covers the basic operations that a statsd
client must support. In practice, we use a third-party statsd client that
supports these operations. However for tests, we use a "mock" implementation
that supports these same operations.

```
type Statsder interface {
	Timing(string, int64) error
	Incr(string, int64) error
	Gauge(string, int64) error
	io.Closer
}
```

`Collector` implements the `Statsder` interface. To build an instance of
`Collector`, you call `NewCollector` and then provide configurations with
`WithHostPort` and `WithBuffering`. The following code snippet creates a
metrics collector and sends the metrics to `StatsD` running on `statsd.hbo.com`
that listens to default port `1234`.

```
// Configures StatsD host and port
addr := WithHostPort("statsd.hbo.com", 1234)

// Configures collector to buffer metrics and send every 30 seconds
buff := WithBuffering(time.Second * 30)

// Add a prefix to all the metrics emitted
prefix := WithPrefix("foobar")

// Recreate the metrics socket every so often
reconnect := WithReconnectInterval(time.Minute * 10)

_, err := NewCollector(addr, buff, prefix, reconnect)
if err != nil {
	panic(err)
}
```


# Singleton

A global instance of `Collector` can be accessed by using `metrics.Singleton`. However, you must initialize the singleton instance with
`metrics.Init` function before you can use it.

```

// Load statsd configuration from config file
config := loadStatsdConfig()

// Iniialize the singleton collector
metrics.Init(config)

// use the singleton collector to update metrics
metrics.Incr("request.count", 1)

```

# Configuration
`metrics.Config` has special unmarshal logic for parsing "time duration" values. One such configuration example is:
```
{
	"host": "localhost",
	"port": 8125,
	"buffered": true,
	"interval": "10s"
	"reconnectInterval": "10m"
}
```

# Dependencies

* github.com/HBOCodeLabs/statsd
