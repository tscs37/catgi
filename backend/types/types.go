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

	// ListGlob returns a list of all files that match the glob string.
	// If context is not nil, the list is not complete and another query
	// needs to be made, if the context is nil, no further queries are
	// required. Backends can use this to effectively prevent rate-limits
	// and load data more efficiently
	// ListGlob MUST NOT list any files not created through Upload but also
	// is NOT REQUIRED to list only files created through Upload.
	ListGlob(
		glob string, ictx context.Context) (
		files []*File, octx context.Context, err error)

	// CleanUp removes expired and incomplete data from the backend
	// if this is necessary or possible.
	CleanUp(ctx context.Context) error
}

// File contains the data of a file, if it's public and when it was created.
type File struct {
	// CreatedAt is the creation time of the file
	CreatedAt time.Time `json:"created_at"`
	// Public marks if the file is public or not
	Public bool `json:"public,omitempty"`
	// Data is the raw binary data of the file
	Data []byte `json:"data,omitempty"`
	// DeleteAt  is the expiry date of a file
	DeleteAt time.Time `json:"delete_at"`
	// Flake is a unique identifier for the file
	Flake string `json:"-"`
	// Content Type sets the Mime Header
	ContentType string `json:"mime"`
}

// DefaultTTL is the default Time-to-Live of new Objects
const DefaultTTL = time.Hour * 24 * 7

// MaxTTL is the maximum Lifetime of an Object
const MaxTTL = time.Hour * 24 * 30

// MinTTL is the minimum Lifetime of an Object
const MinTTL = time.Hour * 1

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
	// ErrorIndexNoSerialize is returned by index.Serialize() or index.Unserialize() when they
	// are not to be stored in the backend
	ErrorIndexNoSerialize = errors.New("Do not serialize this index")
)
