# llog
> A simple, levelled-logging package for Go.

## Versioning

### 7-February-2019 SHA 1a95ec4ccdbcd6a47377102231a3a745bab3a645
*  Opentracing IDs are now present in logs.  `WithCtx()` now logs trace, span, parent, and uber trace ids.

### 13-April-2018 SHA e887dc0d9958392b34511533c0a2084560f9e49c
*  We introduced extended logging `With(keyvals ...interface{})` to allow static values to be included in every log

### 07-August-2017 SHA f5c14d545e8c25b4ff0f23298082c064f2b24364

As of this commit, there're a few of major changes
*  Instead of using [go-kit/log](https://github.com/go-kit/kit/tree/master/log) as the engine, we use Uber's
[Zap sugar log](https://github.com/uber-go/zap) as the engine.  See [Zap's GoDoc](https://godoc.org/go.uber.org/zap) for its APIs.
*  Removal of the feature that sends ERROR logs to `stderr`, DEBUG to `/dev/null`.  Everything (except logger's internal ERROR
log, goes to `stdout`.  However, user can create a logger that sends logs to `stderr` if they wishes like
`myLogger := llog.NewLogger(os.Stderr, INFO)`
*  Removal of the feature that [redirects stdlib logger to Go-kit logger](https://github.com/go-kit/kit/tree/master/log#interact-with-stdlib-logger)
*  We introduced context tracing logging `WithCtx(ctx Context)`


## Dependencies - service that uses `llog` also need to get the following packages
* [Uber's Zap](https://github.com/uber-go/zap) At the time of the writing, we're using zap 1.5.0

## Overview

This package wraps the zap's sugarlog standard logging functionality and adds several key features, namely:

#### Key/Value-style logging
Commonly referred to as ["logfmt"
logs](https://blog.codeship.com/logfmt-a-log-format-thats-easy-to-read-and-write/),
this style of log entry is used across the Hurley nodejs services, and
integrates nicely with Splunk.

#### Levelled logs
Again, to achieve parity with the current Hurley nodejs services, a levelled
logging package was needed.  The available levels are: `DEBUG`, `INFO`,
`WARNING` and `ERROR`.

#### Extended and Context trace logging
It provides `WithCtx` to extract the [B3 propagation standard's](https://github.com/openzipkin/b3-propagation/blob/master/README.md)
`traceId`, `spanId`, as well as our Hurley's custom `X-HBO-Caller` header from the [Context](https://blog.golang.org/context)
object and appends the information in the log.

It also provides `With` to add static values to a logger, which will be appended to the information in the log.



## Motivation

This package was written with a few goals in mind:
  - use basic log levels (DEBUG, INFO, WARN, ERROR)
  - arguments to disabled log levels are not evaluated
  - drop-in replacement for calls to the standard `log` package's
    `Printf`/`Println` functions
  - provides easy methods to allow tracing


## Context Definition
The IDs propagates in-process (via Context object), and eventually downstream (via http headers)

#### TraceId
An ID that originates from the original request from the root client.  It propagates through all the downstream calls.
It usually gets extracted from the `X─B3─TraceId` http header or generated via a middleware.

#### SpanId
An ID that originates from a service whenever it makes a call to a downstream service.  Usually a client library like
[apiclients](https://github.com/HBOCodeLabs/apiclients/) assigns it.

#### Caller
The name of the service that makes the request that's responsible for this log.  (ie. hedwig)
Usually, the middleware assigns it to the `X-HBO-Caller` field.  But in the log, the field name is `caller`


## Usage:

The basic usage of the `llog` package is as follows. As always, check the
package's GoDoc for the most up-to-date documentation.

```
import (
  "errors"
  "os"

  "github.com/HBOCodeLabs/hurley-kit/llog"
)

func main() {
  // to make sure the log is flushed
  defer llog.Sync()
  llog.Info("hello", "world", "username", "dcarney")

  if err := doSomething(); err != nil {
    llog.Error("event", "tryingToDoSomething", "err", err.Error())
  }

  llog.Debug("msg", "this won't show up by default!")

  // enable DEBUG logs, and send them to stdout
  llog.SetLevel(llog.DEBUG)
  llog.SetDebug(os.Stdout)

  llog.Debug("msg", "here I am!")

  // send all INFO logs to /dev/null
  llog.SetInfo(ioutil.Discard)

  llog.Info("msg", "this also won't show up")

  // create a struct that requires a logger instance
  v := StructRequiringLogger{
      Logger: &llog.Logger{},
  }
}

func doSomething() error {
  return errors.New("some error message!")
}
```

Running the above code produces the following log entries:
```
ts="2017-08-03T15:33:11.078-0700", level="INFO", hello="world", username="dcarney",
ts="2017-08-03T15:33:11.078-0700", level="ERROR", event="tryingToDoSomething", err="some error message"
ts="2017-08-03T15:33:11.078-0700", msg="here I am!"
```

If you want to add tracing information to your log
```
import (
  "context"

  "github.com/HBOCodeLabs/hurley-kit/llog"
)

func main() {

  // to make sure the log is flushed
  defer llog.Sync()

  // Usually you'd pass the context from the request.context.  But in this example, we'll create a fresh one
  var ctx = Context.Background()

  // setup tracing information.  (in a middleware)
  ctx = context.WithValue(ctx, HeaderTraceID, "testTrace123")
  ctx = context.WithValue(ctx, HeaderSpanID, "testSpan123")
  ctx = context.WithValue(ctx, HeaderCaller, "pickup")

  // if you want to reuse the ctx, then assign it to a variable
  ctxLog := llog.WithCtx(ctx)
  ctxLog.Info("username", "dcarney")
  ctxLog.Warn("postgres", "outdated")

  // You can also just do
  llog.WithCtx(ctx).Info("event", "videoServed")
}
```
Result is
```
ts="2017-08-03T15:33:11.078-0700", level="INFO", traceId="testTrace123", spanId="testSpan123", caller="pickup", username="dcarney"
ts="2017-08-03T15:33:11.078-0700", level="WARN", traceId="testTrace123", spanId="testSpan123", caller="pickup", postgres="outdated"
ts="2017-08-03T15:33:11.078-0700", level="INFO", traceId="testTrace123", spanId="testSpan123", caller="pickup", event="videoServed"
```

You can also globally init logging with certain values.

```
 // Init all logging to originate from payments-service
 llog.InitWith("service", "payments")

 // or

 // Init all logging to originate from payments-purchase-worker
 llog.InitWith("service", "payments-purchase-worker")
```

Calling llog.InitWith for example in `main.go` while bootstrapping your service makes all logging to have service-keyword included.
This is convenient if same codebase is used with multiple services.

## Example

Some typical comet log entries look like the following:

```
[2016-07-12 21:31:07.855 UTC] [level=INFO] module=redisClientUtil - event=Connected to Redis service=comet line=events.js:180 src=events.js:180:g
```

```
[2016-07-12 21:32:23.989 UTC] [level=INFO] module=comet - apiName=POST.content url=/content method=POST httpStatus=207 elapsed=1 event=serverSend traceId=e4e15385-eb86-4fbc-f61d-69fc04828fdc spanId=1a659a022d7739d3 service=comet
```

In order to be consistent with the `k=v` style logging, this package makes a slight tweak to the log format:

```
ts="2016-07-12 21:31:07.855 UTC", level="INFO", module="redisClientUtil", event="Connected to Redis", line="events.go:180", src="events.go:180:g"
```

This follows the same general `k=v` style, but eliminates the _inconsistent_
`[]` around the timestamp and log level that the current nodejs logging uses.


## Testing
Testing may require additional external dependencies.

### Local
Get needed dependencies

    $ go get -v -t -d

Run tests

    $ go test

### Docker
If you don't have `golang` installed (which you should) then you can use `docker` to build this library

    $ make build

`build` command is a meta target and consist of 2 sub-targets, which could be executed individually (useful for test iteration)

    $ make dependencies-prepare

This will generate a **temporary** `vendor` directory and populate it with needed dependencies

    $ make test

This will run `go test`

    $ make dependencies-clean

This will remove `vendor` directory. This target could be added to `build` if you'd like to clean-up vendor directory after every build.

## Misc.

Why the name `llog`?  I wanted to avoid naming collisions with the std lib's
`log` package (or force client code to use an import alias).  Since
types/functions from a package are always referred to using the package name in
Go, short package names are desirable (`log.Printf` vs
`mycoolloggingpackage.Printf`).  `llog` strikes a balance with both of these
requirements, and gives a subtle indication that the logs are levelled.

## Contributing

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for details and instructions.
