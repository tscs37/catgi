package types

import (
	"context"
	"errors"
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

	// LoadIndex will load any stored index data from the backend and
	// replace the index with that. It will utilize the Unserialize() function
	// of the Index for this
	// If no index is found, the index itself is not modified.
	LoadIndex(Index, context.Context) error
	// StoreIndex uses the Index.Serialize() function to store the index
	// inside the backend
	StoreIndex(Index, context.Context) error

	// ListGlob returns a list of all files that match the glob string.
	// If context is not nil, the list is not complete and another query
	// needs to be made, if the context is nil, no further queries are
	// required. Backends can use this to effectively prevent rate-limits
	// and load data more efficiently
	// ListGlob MUST NOT list any files not created through Upload
	ListGlob(
		glob string, ictx context.Context) (
		files []*File, octx context.Context, err error)
}

// Index provides an interface to turn a HTTP Request into a snowflake id
// for the backend and caching of request data.
// The Index must be usuable without loading data from the backend
// as the index data may not be loaded upon first use.
type Index interface {
	// Name returns the name of the current driver
	Name() string

	// Serialize returns the current index contents. Serialization is
	// index-driver specific.
	Serialize(context.Context) ([]byte, error)

	// Unserialize parses the given byte structure into the index.
	// Serialization is index-driver specific.
	//
	// Unserialize will call Clear() before loading the index, any
	// existing index is replaced. Drivers should make an effort to
	// do this atomically.
	Unserialize([]byte, context.Context) error

	// Get returns the file associated with the request using the provided
	// backend. "cached"" indidcates wether the file was retrieved from
	// cache. If "err" is set, it indicates negative caching.
	Get(File, Backend, context.Context) (cached bool, file *File, err error)

	// Put stores the content of the request into the provided backend.
	// "cached" indicates if the put request is continuing in the background
	// such that the file is not uploaded yet but has been put into the cache
	Put(File, Backend, context.Context) (cached bool, file *File, err error)

	// Del deletes a file from the given backend. "cached" indicates
	// if the deletion is pending in background.
	Del(File, Backend, context.Context) (cached bool, err error)

	// Flush blocks until all cache operations are complete and blocks
	// new cache operations until it completes.
	Flush(context.Context) error

	// Clear calls Flush and then deletes the entire cache
	Clear(context.Context) error

	// Collect cleans out expired files from the backend and cache
	Collect(Backend, context.Context) error
}

// File contains the data of a file, if it's public and when it was created.
type File struct {
	// CreatedAt is the creation time of the file
	CreatedAt time.Time `json:"created_at"`
	// Public marks if the file is public or not
	Public bool `json:"public"`
	// Data is the raw binary data of the file
	Data []byte `json:"data"`
	// DeleteAt  is the expiry date of a file
	DeleteAt time.Time `json:"delete_at"`
	// Flake is the Name of the File
	Flake string `json:"-"`
}

// DefaultTTL is the default Time-to-Live of new Objects
const DefaultTTL = time.Hour * 24 * 7

// MaxTTL is the maximum Lifetime of an Object
const MaxTTL = time.Hour * 24 * 30

// MinTTL is the minimum Lifetime of an Object
const MinTTL = time.Hour * 1

var (
	// ErrorFileNotExist is returned when a requested file does not
	// exist. Non-fatal when returned from backend.Exists().
	ErrorFileNotExist = errors.New("File does not exist")
	// ErrorIndexNoSerialize is returned by index.Serialize() or index.Unserialize() when they
	// are not to be stored in the backend
	ErrorIndexNoSerialize = errors.New("Do not serialize this index")
)
