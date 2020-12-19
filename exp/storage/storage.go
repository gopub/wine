package storage

import "context"

type Reader interface {
	Read(ctx context.Context, name string) ([]byte, error)
}

type Writer interface {
	Write(ctx context.Context, o *Object) (url string, err error)
}
