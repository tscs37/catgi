package types

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

// Backend allows the server to retrieve single files based on snowflake
// and also load and store the index data.
type Backend interface {
	// Name returns an identifier for the backend type
	// This does not need to be equivalent to the driver name.
	Name() string

	// Upload creates a file. If the file already exist, it must abort
	// with an error
	// Upload needs to ensure that no two files with the same name are
	// uploaded, ie operate atomically.
	Upload(name string, file *File, ctx context.Context) error
	// Exists returns nil if the name exists.
	// If the name does not exists, returns ErrorFileNotExist,
	// if an actual error occurs, returns an error instance that is not
	// equal to ErrorFileNotExist
	Exists(name string, ctx context.Context) error
	// Get returns the file associated with the flake in name
	Get(name string, ctx context.Context) (*File, error)
	// Delete Removes a file from the backend
	// This is required to happen atomically and if the file
	// cannot be delete now but only later, it *must* still return an error
	Delete(name string, ctx context.Context) error

	// ListGlob returns a list of all files with the given prefix
	// The returned file structs need to contain the name but need not
	// contain the file data
	ListGlob(ctx context.Context, glob string) (files []*File, err error)
}

// Publisher is implemented by backends that support clear name publishing
// of files.
type Publisher interface {
	Backend
	// Publish associated a name with a flake for clearname publishing
	// If the name is already taken, this fails
	// A name may be associated with more than one flake
	Publish(flake []string, name string, ctx context.Context) error
	// Unpublish disassociates a name from any flakes it's associated with.
	Unpublish(name string, ctx context.Context) error
	// Resolves takes a name and returns a set of flakes for that name
	Resolve(name string, ctx context.Context) ([]string, error)
}

// KVBackend is for small and indexed data, large data
// should not be saved in here.
// Structs that are saved in a KVBackend need to be fully serializable.
// Keys will be formated using slash seperated prefix sets.
//
// Example: /prefix1/prefix1/key
//
// Some keys may also use a "<name>:" prefix with following additional
// prefixes.
type KVBackend interface {
	Backend
	// Set returns the saved interface
	KVSet(ctx context.Context, key string, value interface{}) error
	// Get returns the saved interface
	KVGet(ctx context.Context, key string) (interface{}, error)
	// Has returns true if the key exists
	KVHas(ctx context.Context, key string) (bool, error)
	// Ls returns a list of all keys with a given prefix
	KVLs(ctx context.Context, prefix string) ([]string, error)
}

// ContentBackend implements a simple KV-store
// Unlike the KVBackend it's entries need not be indexed
// and it is supposed to handle much larger files than
// KVBackend
//
// Unlike the methods provided by the standard backend, these methods
// should allow for a more raw access, including accessing files by
// internal names (although this is optional)
type ContentBackend interface {
	Backend
	// Write will read data from reader and write it to the specified named
	// file
	CntWrite(ctx context.Context, name string, reader io.Reader) error
	// Read will write any data from the named file into writer.
	CntRead(ctx context.Context, name string, writer io.Writer) error
}

// File contains the data of a file, if it's public and when it was created.
type File struct {
	// CreatedAt is the creation time of the file
	CreatedAt *DateOnlyTime `json:"created_at"`
	// Public marks if the file is public or not
	Public bool `json:"public,omitempty"`
	// Data is the raw binary data of the file
	Data []byte `json:"data,omitempty"`
	// DeleteAt  is the expiry date of a file
	DeleteAt *DateOnlyTime `json:"delete_at"`
	// Flake is a unique identifier for the file
	Flake string `json:"name"`
	// Content Type sets the Mime Header
	ContentType string `json:"mime"`
	// If this is not empty, use this extension for downloads.
	FileExtension string `json:"ext,omitempty"`
	// Username of who uploaded the file
	User string `json:"usr,omitempty"`
}

// DefaultTTL is the default Time-to-Live of new Objects
const DefaultTTL = time.Hour * 24 * 7

// MaxTTL is the maximum Lifetime of an Object
const MaxTTL = time.Hour * 24 * 30

// MinTTL is the minimum Lifetime of an Object
const MinTTL = time.Hour * 1

// MaxDataSize is the maximum size of a file
const MaxDataSize = 25 * 1024 * 1024

// SkipSize marks how many characters should be grouped when
// splitting filenames. If filenames aren't split, this can be
// ignored.
const SkipSize = 2

// ErrorFileNotExist is returned when a requested file does not
// exist. Non-fatal when returned from backend.Exists().
type ErrorFileNotExist struct {
	Object     string
	InnerError error
}

// Error returns a nested Error Message regarding a missing file
func (e ErrorFileNotExist) Error() string {
	return "ErrFileNotExist(" + e.Object + "): " + e.InnerError.Error()
}

// NewErrorFileNotExists returns a ErrFileNotExist typed error.
func NewErrorFileNotExists(name string, err error) error {
	if err == nil {
		err = errors.New("generic file not exist")
	}
	return ErrorFileNotExist{
		Object:     name,
		InnerError: err,
	}
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
	// ErrorIndexNoSerialize is returned by index.Serialize() or index.Unserialize() when they
	// are not to be stored in the backend
	ErrorIndexNoSerialize = errors.New("Do not serialize this index")
)

type DateOnlyTime struct {
	time.Time
}

func (dot *DateOnlyTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if len(s) == 2 {
		return errors.New("Cannot parse empty date")
	}
	s = s[1 : len(s)-1]

	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	dot.Time = t
	return nil
}

func (dot *DateOnlyTime) MarshalJSON() ([]byte, error) {
	s := dot.Format("2006-01-02")
	s = fmt.Sprintf("\"%s\"", s)
	return []byte(s), nil
}

func FromTime(t time.Time) *DateOnlyTime {
	return &DateOnlyTime{
		Time: t,
	}
}

func FromString(s string) (*DateOnlyTime, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, err
	}
	return FromTime(t), nil
}
