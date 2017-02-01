package common

import (
	"context"
	"io"
	"net/http"
)

type BackendPingFile interface {
	Ping(string, context.Context) (bool, interface{}, error)
}
type BackendDirectIOByte interface {
	WriteBytes(string, []byte, context.Context) error
	ReadBytes(string, context.Context) ([]byte, error)
}

type BackendDirectIOReader interface {
	WriteReader(string, io.ReadSeeker, context.Context) error
	ReadReader(string, context.Context) (io.ReadCloser, error)
}

type BackendWithHTTPHandler interface {
	// GetHTTPHandler returns the HTTP handler that should respond to
	// queries. The HTTP Prefix is stripped from the URL but not the RequestURI.
	GetHTTPHandler() http.Handler
	// HTTPPrefix returns a slice of strings that resemble the prefix of the
	// URL path that is routed to this backend.
	//
	// Example:
	// "file" -> /file/
	// "res" -> /res/
	// "res/file" -> /res/file/
	HTTPPrefix() []string
}
