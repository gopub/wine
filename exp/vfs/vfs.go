package vfs

import (
	"github.com/gopub/types"
	"path/filepath"
	"strings"

	"github.com/gopub/errors"
)

const (
	keySize  = 64
	pageSize = int64(32 * types.KB)
)

const (
	keyFSHome = "filesystem.home"
	keyFSKey  = "filesystem.key"
)

const ErrNotExist = errors.String("not exist")

type KVStorage interface {
	// Get returns ErrNotExist if key doesn't exist
	Get(key string) ([]byte, error)
	Put(key string, val []byte) error
	Delete(key string) error
}

func splitPath(path string) []string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	paths := strings.Split(path, "/")
	return paths
}
