# Catgi Backend & Frontend Interlink

The backends and frontends of catgi need to seamlessly communicate
intends and parameters to properly function.

This document attempts to specify how backends and frontends should
pass data around or communicate configuration.

## Default Handler Context

The Default HTTP Handler will pass on any parameters beginning
with `ctx_` to the context and hands this over to the backends.

This allows backends to utilize URL parameters for operation, like de- or
encryption.

## ErrorHTTPOptions

The backend may return an ErrorHTTPOptions.

This is not an actual error but is used to signal the frontend that the
backend requires additional control over the HTTP response.

The function `PassOverHTTP(http.ResponseWriter)` **must** be called if
this error is present, this function will set Cookies or HTTP headers.

The function `WantTakeover()` returns `true` if the backend requires full
control over the response. In this case the caller must pass control
over the request, responsewriter and the context to the function
stored in `HTTPTakeover` and return after it finishes.

## BackendOptions HTTP Handler

Backends can provide a HTTP Handler along with a identifier.

The default frontend uses the identifiers `file` and `f`.

When the backend has been setup, the HTTP Server will ask
the underlying backend for the all backend that support
HTTP Handling.

Backends may filter lower-placed backends Handlers if necessary.

This will most likely affect backends that have to access several underlying
backends.
