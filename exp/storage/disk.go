package storage

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
)

//DiskBucket implements reader/writer based on local file system
type DiskBucket struct {
	dir string
}

func NewDiskBucket(dir string) (*DiskBucket, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &DiskBucket{
		dir: dir,
	}, nil
}

func (b *DiskBucket) Write(ctx context.Context, o *Object) (string, error) {
	name := filepath.Join(b.dir, o.Name)
	errC := make(chan error, 1)
	go func() {
		defer close(errC)
		errC <- ioutil.WriteFile(name, o.Content, 0644)
	}()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errC:
		if err == nil {
			return name, nil
		}
		return "", err
	}
}

func (b *DiskBucket) Read(ctx context.Context, name string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(b.dir, name))
}
