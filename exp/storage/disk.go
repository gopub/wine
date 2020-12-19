package storage

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

//DiskBucket implements reader/writer based on local file system
type DiskBucket struct {
	dir string
}

func NewDiskBucket(dir string) *DiskBucket {
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(fmt.Sprintf("storage: create dir %s: %v", dir, err))
	}
	return &DiskBucket{
		dir: dir,
	}
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
