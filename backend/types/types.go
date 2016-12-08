package types

import (
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
	Upload(name string, file *File) error
	// Exists returns nil if the name exists or an error if it does not.
	Exists(name string) error
	// Get returns the file associated with the flake in name
	Get(name string) (*File, error)
	// Delete Removes a file from the backend
	Delete(name string) error
	// LoadIndex will load any stored index data from the backend and
	// replace the index with that. It will utilize the Unserialize() function
	// of the Index for this
	// If no index is found, the index itself is not modified.
	LoadIndex(Index) error
	// StoreIndex uses the Index.Serialize() function to store the index
	// inside the backend
	StoreIndex(Index) error
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
	Serialize() ([]byte, error)

	// Unserialize parses the given byte structure into the index.
	// Serialization is index-driver specific.
	//
	// Unserialize will call Clear() before loading the index, any
	// existing index is replaced. Drivers should make an effort to
	// do this atomically.
	Unserialize([]byte) error

	// Get returns the file associated with the request using the provided
	// backend. "cached"" indidcates wether the file was retrieved from
	// cache. If "err" is set, it indicates negative caching.
	Get(GetRequest, Backend) (cached bool, file *File, err error)

	// Put stores the content of the request into the provided backend.
	// "cached" indicates if the put request is continuing in the background
	// such that the file is not uploaded yet but has been put into the cache
	Put(File, Backend) (cached bool, file *File, err error)

	// Del deletes a file from the given backend. "cached" indicates
	// if the deletion is pending in background.
	Del(DelRequest, Backend) (cached bool, err error)

	// Flush blocks until all cache operations are complete and blocks
	// new cache operations until it completes.
	Flush() error

	// Clear calls Flush and then deletes the entire cache
	Clear() error

	// Collect cleans out expired files from the backend and cache
	Collect(Backend) error
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

type GetRequest struct {
	Flake string
}

type DelRequest struct {
	Flake string
}

var defaultTTL = time.Hour * 24 * 7

// DefaultTTL is the default Time To Live for Objects
var DefaultTTL = &defaultTTL

const MaxTTL = time.Hour * 24 * 30
const MinTTL = time.Hour * 1
