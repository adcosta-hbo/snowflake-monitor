# request
> An HTTP client with sane default behavior and a few nice features

## Motivation

The "Hurley" system, being composed of a multitude of "microservices", relies
heavily on communication over HTTP.  We've done a good job breaking up services
into small-ish pieces that scale indendently, but we _haven't_ done as good of
a job ensuring that the communication between them is well-behaved.

In particular, we've seen issues with:

- no use of HTTP keepalive
- _very_ long request timeouts
- no "circuit breaking" when an upstream dependency starts to fail
- bloated/coupled dependency chains

In light of the above, it's become clear that we need to have a lightweight
HTTP client with sane default behavior, and a few features built-in that can
help make our service <--> service communication more robust and fool-proof.

## Design goals:

- shorter default request timeouts
- use of HTTP keepalive by default
- a built-in [circuit breaker](https://martinfowler.com/bliki/CircuitBreaker.html), to help guard
against cascading failures
- up-to-date dependencies
- a cleaner, more composable API

## Usage

The basic usage is almost _identical_ to the code that utilizes the default Go
HTTP client.  The only difference is, the `http.Client` instance is created
using `request.NewClient()`:

```
import "github.com/HBOCodeLabs/hurley-kit/http/request"

...
client := request.NewClient()

req, err := http.NewRequest(http.MethodGet, someURL, nil)
...

resp, err := client.Do(req)
if err != nil {
  // do something
}

// close the response body, to ensure HTTP keepalive is used!
defer resp.Body.Close()
```

Since the `request.NewClient` function simply returns an instance of
`http.Client`, it can be used as a "drop-in replacement" for Go's
default HTTP client!

For control over the default request timeout, circuit breaking behavior, etc.,
use the various configuration functions to create a customized HTTP client:

```
import "github.com/HBOCodeLabs/hurley-kit/http/request"

...

// open the breaker if we see > 25% failure in a 1-minute window, with a
minimum of 100 requests
cb := request.CircuitBreaker(60 * time.Second, 100, 25)

// use a 1 second request timeout
rt := request.Timeout(1 * time.Second)

// use a transport with maximum of 100 idle connections per host
transport := &http.Transport{
		MaxIdleConnsPerHost:   100
}
tr := request.Transport(transport)

client := request.NewClient(cb, rt, tr)

// proceed to use the `client` just as before
```

### Keepalive

Using HTTP keepalive can make a dramatic improvement in the performance of any
HTTP client, since the overhead of establishing a new TCP connection (doing a
DNS lookup, connecting to the host, peforming the TCP 3-way handshake, etc.)
does *not* have to be done for every new HTTP request.

Luckily, Go's `http.Client` has keepalive enabled by default, as long as the
following conditions are met:

- the `Body` of an `http.Response` is read and closed after use, so it's
  connection can be re-used
- the connection is reused because it fits in the `MaxIdleConnsPerHost` setting
  (default of 2)

The first point can be illustrated with a simple example:
```
resp, err := client.Do(req)
if err != nil {
    return nil, err
}

// do what you want with resp.Body, but DON'T FORGET
// TO CLOSE IT, so the connection can be reused.
defer resp.Body.Close()
```

The second point only comes into play when many concurrent requests need to be
made to the same hostname.  Because DNS names and load balancers count as a
single host, the `http.Client` will only open 2 keep-alive connections to
each host by default. If requests reach a high enough rate, new connections
will be established and destroyed for every request. This connection thrashing
results in a very high spike in CPU load for TLS and GC, and a spike in
latencies for establishing connections before every request.

If many concurrent requests need to be made to the same host, this setting can
be changed.
