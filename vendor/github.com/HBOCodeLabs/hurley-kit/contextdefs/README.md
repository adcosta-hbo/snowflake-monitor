# contextdefs

> A Go package containing definitions of context-related keys and functions

## Background

The [context package](https://golang.org/pkg/context/) for Go provides a
facility for sharing request-scoped data throughout the HTTP lifecycle.

In the words of the package's documentation:

> Package context defines the Context type, which carries deadlines,
> cancelation signals, and other request-scoped values across API boundaries
> and between processes.

This package, coupled with the fact that every http.Request has [a .Context()
method](https://golang.org/pkg/net/http/#Request.Context) that returns the
context of an HTTP request, means that context is the de facto location for
storing and retrieving request-scoped data.

### Issues with Context Usage

The context object stores arbitrary key/value pairs of any type, making it
incredibly easy to get started using.

There is one wrinkle however: [the official documentation](https://golang.org/pkg/context/#WithValue)
recommends against using plain-old `string` keys when using the context:

> The provided key must be comparable and should not be of type string or any
> other built-in type to avoid collisions between packages using context. Users
> of WithValue should define their own types for keys.

Based on this advice, packages will typically do something like the following:

```
package foo

type ContextKey string

const (
  MySpecialContextKey = ContextKey("foobar")

  AnotherContextKey = ContextKey("fizzbuzz")
)
```

This prevents two packages from ever accidentally using the same context key
and stomping on each other's values, since only package `foo` defines the
ContextKey type.

However, this leads to one glaring problem: excessive code coupling.  Any code
that wants to retrieve one of these values from the context has to import the
`foo` package in order to reference the context keys:

```
package main

import "github.com/HBOCodeLabs/foo"

func main() {
  ...
  ctx := req.Context()

  mySpecialValue := ctx.Value(foo.MySpecialContextKey)
  ...
}

```

This might not seem like such a big issue until you think about packages that
deal with very common functionality, like tokens.  If we used the above
strategy, such a package would need to be imported _everywhere_ that the token
value was to be pulled from a context, which would likely be...everywhere.

This same issue still arises even if the package (`foo` in this example) were to
provide "getter/setter" functions for reading/writing the context values.

### A Better Solution?

After much discussion and reading, a solution was proposed to create a package
that does nothing but define shared context keys (and possibly context-related
helper functions).

This package has no external dependencies, and can be imported in as many
locations as necessary, without a huge hassle or a messy dependency tree.  It
shouldn't require frequent upgrades in consuming code, or be difficult to
upgrade when the time comes.  Any code that wants to access a "shared" context
value still has to import a package in order to reference the context key
definition, but that import _only_ contains context key-related definitions.

This repo contains that package.

## Additional reading:

This solution to the problem of sharing context keys without excessing package
coupling is the best idea we've been able to come up with/glean from the
community to date, but we're open to additional ideas/approaches!

For additional resources about Go's context and its usage, check out:

- https://blog.golang.org/context
- https://medium.com/@matryer/context-keys-in-go-5312346a868d
- https://medium.com/@matryer/context-has-arrived-per-request-state-in-go-1-7-4d095be83bd8
- https://medium.com/golangspec/globally-unique-key-for-context-value-in-golang-62026854b48f
