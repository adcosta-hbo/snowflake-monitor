# hurley-kit
A standard library for Go microservices at HBO

### Contact Info

The `#golang` channel is the most efficacious way to reach out to all the
developers of the packages in `hurley-kit`.  It's also a great place to post a
PR link, should you find yourself [contributing to the project](.github/CONTRIBUTING).

Various packages within the repo have different code owners, per the
`.github/CODEOWNERS` file.

## Background

`hurley-kit` is a central collection of individual packages that can be used to
build Go microservices (typically HTTP servers), using the conventions and
packages written by the Hurley engineers. The repo organization (as well as the
name) was directly insprired by [go-kit](https://github.com/go-kit/kit).

It is important to note that `hurley-kit` is a collection of _individual Go
packages_, each of which can be used independently in an application.

The fact that these packages are grouped into a single _repo_ does not mean
that `hurley-kit` is a single big package/library!  Rather than get into a
monolithic package situation (ala `Hurley-Common`), we have split each discrete
piece of functionality into its own package, which can be imported
independently into an application.

*Example*: to use the `metrics` package, you might have an `import` block that looks like:

```
import (
    "context"
    "fmt"
    "net/http"

    "github.com/HBOCodeLabs/hurley-kit/metrics"
)

```

For additional background, you might find [these slides](https://docs.google.com/presentation/d/1hjq_6gbz_Wyt3E5mf7cGEoix2vHDwijiyyYEHHjDA2s/edit?usp=sharing) helpful.

## Packages

For details about specific packages, see the individual package READMEs:

- [auth](auth/README.md)
- [auth/tokens](auth/tokens/README.md)
- [auth/tokens/btc](auth/tokens/btc/README.md)
- [btc](btc/README.md)
- [contextdefs](contextdefs/README.md)
- [http/request](http/request/README.md)
- [llog](llog/README.md)
- [metrics](metrics/README.md)
- [tenantconfig](tenantconfig/README.md)
- [schemavalidation](schemavalidation/README.md)
- [secrets](secrets/README.md)
- [strutil](strutil/README.md)
- [tracing](tracing/README.md)

## Contributing

Detailed contribution guidelines can be found [here](https://github.com/HBOCodeLabs/hurley-kit/blob/master/.github/CONTRIBUTING.md).

## Goals

We intend for the code in the `hurley-kit` packages to be:
- well written
- well tested
- well documented
- service-agnostic (e.g. not specific to service _XYZ_)

To those ends, please familarize yourself with the [contribution
guidelines](https://github.com/HBOCodeLabs/hurley-kit/blob/master/.github/CONTRIBUTING.md)
before opening a PR.

By grouping these packages together in one repo, we hope to achieve:
- better discoverability of Go packages
- less fragmentation/duplication across repos
- better interoperability
- more consistent code style and conventions
- ease of maintenance
  - Jenkins pipelines and "best practices"
  - Go version upgrades
