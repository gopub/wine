package vfs

import (
	"path/filepath"
	"strings"

	"github.com/gopub/conv"

	"github.com/gopub/types"
)

const (
	keySize  = 64
	pageSize = int64(32 * types.KB)
)

var (
	keyFSHome        = conv.SHA256("filesystem.home")
	keyFSCredentials = conv.SHA256("filesystem.credentials")
)

type KVStorage interface {
	// Get returns os.ErrNotExist if key doesn't exist
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

func validateFileName(name string) bool {
	if name == "" {
		return false
	}

	if strings.Contains(name, "/") {
		return false
	}

	return true
}
