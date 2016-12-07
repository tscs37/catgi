package types

import (
	"net/http"
	"time"
)

// Backend allows the server to retrieve single files based on snowflake
// and also load and store the index data.
type Backend interface {
	Name() string
	// Upload creates a file. If the file already exist, it must abort
	// with an error
	Upload(name string, ttl *time.Duration, file *File) error
	// Exists returns nil if the name exists or an error if it does not.
	Exists(name string) error
	Get(name string) (*File, error)

	LoadIndex(Index) error
	StoreIndex(Index) error
	Delete(name string) error
}

// Index provides an interface to turn a HTTP Request into a snowflake id
// for the backend and caching of request data.
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

	// Resolve takes an http.Request and resolves it to a proper
	// snowflake of a file. It does not ensure if the file exists
	Resolve(http.Request) (string, error)

	// Get returns the file associated with the request using the provided
	// backend. "cached"" indidcates wether the file was retrieved from
	// cache. If "err" is set, it indicates negative caching.
	Get(GetRequest, Backend) (cached bool, file *File, err error)

	// Put stores the content of the request into the provided backend.
	// "cached" indicates if the put request is continuing in the background
	// such that the file is not uploaded yet but has been put into the cache
	Put(PutRequest, Backend) (cached bool, file *File, err error)

	// Del deletes a file from the given backend. "cached" indicates
	// if the deletion is pending in background.
	Del(DelRequest, Backend) (cached bool, err error)

	// Flush blocks until all cache operations are complete and blocks
	// new cache operations until it completes.
	Flush() error

	// Clear calls Flush and then deletes the entire cache
	Clear() error
}

type File struct {
	CreatedAt time.Time `json:"created_at"`
	Public    bool      `json:"public"`
	Data      []byte    `json:"data"`
}

type PutRequest struct {
	Data   []byte
	Public bool
	TTL    time.Duration
}

type GetRequest struct {
	Flake string
}

type DelRequest struct {
	Flake string
}

var defaultTTL = time.Hour * 24 * 30
var DefaultTTL = &defaultTTL
