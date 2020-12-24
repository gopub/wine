package httpvfs

import "net/http"

type bytesFS []byte

var _ http.FileSystem = (bytesFS)(nil)

func (f bytesFS) Open(name string) (http.File, error) {
	return NewFile(name, f), nil
}

func NewFileSystem(b []byte) http.FileSystem {
	return bytesFS(b)
}
