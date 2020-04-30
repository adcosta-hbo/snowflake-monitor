# auth

This package contains an HTTP middleware that:
- the request contains an OAuth bearer token in the "Authorization" header
- the correct bearer strategy is used
- the token can be decoded successfully
- the token has a valid signature
- the token is not expired

Optionally, the middleware can be configured to check that:
- the token is a _user_ token (as opposed to a client token)
- the token contains a certain set of permissions

Additionally, the middleware will store the decoded token in the [context of
the request](https://golang.org/pkg/net/http/#Request.Context), for retrieval
by code further down the middleware chain.

The middleware returns the appropriate `4xx` responses if any of the above
assertions fail to hold.

## Versioning

### 09-Oct-2018 SHA 9bd6c954b1b7dbf8c9c6efde5cb9b4a926052e05
[PR #27](https://github.com/HBOCodeLabs/hurley-kit/pull/27)
* `GetTokenFromContext` now returns a Tokener interface instead of a decoded
  token object. It's done because the token's `claim` property is meant to be
  private, hidden behind the interface
* `AddTokenToContext` is now public and takes in the specific Tokener interface.
  We can now use this to mock in test.



### Example usage
```
// next is an `http.Handler`

decoder := tokens.NewDecoder("someS3cret")
handler := auth.NewMiddleware(decoder, next)

// handler is now an `http.Handler` that includes the token validation
```

*NOTE*: This middleware should ideally be installed _after_ any metrics/logging
middleware, so that the service can gather metrics and/or logs on the number of
`400`/`401` responses produced by this middleware.

### Requiring a user token

To require that each auth token be a _user_ token, rather than a client token,
simply use the `RequireUserToken` configuration func:

```
decoder := tokens.NewDecoder("someS3cret")
auth.NewMiddleware(decoder, next, auth.RequireUserToken())
```

### Requiring permissions

To require that each token contain a certain set of permissions, simply use the
`RequirePermissions` configuration func, and the desired permissions from the
[tokens](tokens/README.md) package:

```
perms := auth.RequirePermissions(tokens.PermissionPlayVideo, tokens.PermissionUpdateProfile)

decoder := tokens.NewDecoder("someS3cret")
auth.NewMiddleware(decoder, next, perms)
```

### Context support

To insert a Tokener object into the HTTP request context, use the `AddTokenToContext` method.
This is automatically called by the middleware.  But this method is handy for tests
```
...
testToken := &testTokenObject{}       // testTokenObject implements Tokener interface
req = AddTokenToContext(req, testToken)
...
```

HTTP handlers can then access the decoded tokener, by using the `GetTokenFromContext` method:
```
...
tokener := auth.GetTokenFromContext(req.Context())
if tokener != nil {
  // we have a decoded auth tokener!
}
...
```

or, if desired, the `contextdefs` package can be used to retrieve the token
from the context directly, using the proper type assertion:

```
val := ctx.Value(contextdefs.UserToken)
token := val.*(tokens.Token)
```
