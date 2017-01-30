package common

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gopkg.in/vmihailenco/msgpack.v2"
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

	// RunGC will clean up expired files from the storage backend.
	// On automatically expiring backends, this returns an empty array
	// and a nil error.
	// Otherwise it will return an array of all deleted files minus their
	// content.
	// If Deleting a file fails, the function returns with an error
	// and a full list of files that were about to be deleted.
	RunGC(ctx context.Context) ([]File, error)
}

// Temporary Onion Interface
type OnionBackend interface {
	// GetFirstWith returns the first backend in the hierarchy
	// that provides a specific option and if none of the underlying
	// Backends provide that option, returns nil.
	//
	// The resulting Backend must first be converted into the target
	// interface before using the functions, be sure to check
	// if that is OK.
	//
	// This function is used because Go's type system handles nested
	// interface propagation with the agility and grace of a blue whale
	// being dropped out of a Boeing 747 at nominal flight altitude.
	GetFirstWith(options BackendOption) Backend

	// GetAllWith returns a list of all backends configured that
	// provide all given options.
	//
	// Some backends may filter this call if they cannot reasonably
	// provide underlying functionality.
	//
	// This works like GetFirstWith but does not stop once
	// a backend with the given options is found.
	GetAllWith(options BackendOption) []Backend

	// GetOptions returns all options of a backend as a OR'd value
	GetOptions() BackendOption
}

func BackendHasOptions(b Backend, opts BackendOption) bool {
	if ob, ok := b.(OnionBackend); ok {
		if ob.GetOptions()&opts == opts {
			return true
		}
	}
	return false
}

type BackendOption uint

const (
	// The Backend is able to provide a statistics snapshot
	BackendOptionStatistics BackendOption = 1 << iota
	// BackendOptionDirectBytesIO indicates the backends supports
	// storing byte slices directly via Write, Read and Delete Methods
	BackendOptionDirectBytesIO
	// BackendOptionDirectReaderIO indicates the backend supports
	// storing reader data directly via Write, Read and Delete Methods
	BackendOptionDirectReaderIO
	BackendOptionPingFile
)

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