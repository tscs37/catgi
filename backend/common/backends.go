package common

import "context"
import "io"

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