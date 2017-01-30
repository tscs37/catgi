package common

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
	// Options is a list of file options
	// This may be altered by the backend to indicate
	// certain file conditions
	Options []FileOption `json:"opts,omitempty"`
}

func (f File) Valid() bool {
	if f.DeleteAt.TTL() == 0 {
		return false
	}
	if len(f.Options) > 100 {
		return false
	}
	if f.CreatedAt == nil {
		return false
	}
	if f.DeleteAt == nil {
		return false
	}
	return true
}

func (f File) HasOption(opt FileOption) bool {
	for k := range f.Options {
		if opt == f.Options[k] {
			return true
		}
	}
	return false
}

type FileOption string

const (
	// Indicates that no cache should store this file
	// if possible.
	OptionDisableCache = "nocache"
)
