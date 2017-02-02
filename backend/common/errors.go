package common

import (
	"context"
	"errors"
	"net/http"
)

// ErrorFileNotExist is returned when a requested file does not
// exist. Non-fatal when returned from backend.Exists().
type ErrorFileNotExist struct {
	Object     string
	InnerError error
}

// Error returns a nested Error Message regarding a missing file
func (e ErrorFileNotExist) Error() string {
	if e.InnerError != nil {
		return "ErrFileNotExist(" + e.Object + "): " + e.InnerError.Error()
	}
	return "ErrFileNotExist(" + e.Object + ")"
}

// NewErrorFileNotExists returns a ErrFileNotExist typed error.
func NewErrorFileNotExists(name string, err error) error {
	return ErrorFileNotExist{
		Object:     name,
		InnerError: err,
	}
}

func IsFileNotExists(err error) bool {
	_, ok := err.(ErrorFileNotExist)
	return ok
}

type ErrorHTTPOptions struct {
	Cookies      []*http.Cookie
	Headers      map[string]string
	HTTPTakeover func(r *http.Request, w http.ResponseWriter, ctx context.Context)
}

func (e ErrorHTTPOptions) Error() string {
	return "Frontend does not accept HTTP Options"
}

func (e ErrorHTTPOptions) PassOverHTTP(w http.ResponseWriter) {
	for _, cookie := range e.Cookies {
		http.SetCookie(w, cookie)
	}
	for name, value := range e.Headers {
		w.Header().Add(name, value)
	}
}

func (e ErrorHTTPOptions) WantsTakeover() bool {
	if e.HTTPTakeover != nil {
		return true
	}
	return false
}

func IsHTTPOption(err error) bool {
	_, ok := err.(ErrorHTTPOptions)
	return ok
}

var (
	// ErrorNotImplemented is returned if the underlying interface has
	// not implemented a function. The presence of ErrorNotImplemented is
	// not acceptable for any production-ready backend.
	ErrorNotImplemented = errors.New("Request without Implementation")
	// ErrorExpired is returned when the file that was requested has been
	// found but was deleted because it expired
	ErrorExpired = errors.New("The requested file has expired")
	// ErrorQuotaExceeded is returned when the request issued exceeded a
	// quota in the backend, for example if a file is too large or a publish
	// contains too many flakes.
	ErrorQuotaExceeded = errors.New("Backend aborted because a quota was exceeded")
	// ErrorIncompleteWrite is returned when the underlying data was not
	// written to the backend entirely and may be in an inconsistent state.
	ErrorIncompleteWrite = errors.New("Backend could not complete write")
	// ErrorFileExists is returned when the backend tried to upload
	// a file over an existing one.
	ErrorFileExists = errors.New("File already present in backend")
	// ErrorSerializationFailure is returned when the given file
	// could not be serialized and no other error information is available.
	ErrorSerializationFailure = errors.New("Could not serialize file data")
)
