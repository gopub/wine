package vfs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gopub/errors"
	"github.com/gopub/log/v2"
	"github.com/gopub/types"
)

var logger *log.Logger

func init() {
	logger = log.Default().Derive("Wine.vfs")
}

func SetLogger(l *log.Logger) {
	logger = l
}

// Important constants, do not change
const (
	keySize         = 32
	DefaultPageSize = int64(4 * types.MB)
	MinPageSize     = int64(32 * types.KB)
)

const (
	ErrAuth errors.String = "invalid password"
)

const (
	keyFSRootDir    = "filesystem.root"
	keyFSUUID       = "filesystem.uuid"
	keyFSCredential = "filesystem.credential"
	keyFSConfig     = "filesystem.config"
	keyFSPageSize   = "filesystem.page_size"
	keyFSKeyChain   = "filesystem.keychain"
)

type Storage interface {
	// Get returns os.ErrNotExist if key doesn't exist
	Get(key string) ([]byte, error)
	Put(key string, val []byte) error
	Delete(key string) error
	Close() error
}

func cleanName(name string) string {
	name = filepath.Clean(name)
	name = strings.TrimPrefix(name, "/")
	name = strings.TrimSuffix(name, "/")
	return name
}

func splitPath(path string) []string {
	path = filepath.Clean(path)
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	paths := strings.Split(path, "/")
	return paths
}

type Flag int

const (
	ReadOnly  = Flag(os.O_RDONLY)
	WriteOnly = Flag(os.O_WRONLY)
	Create    = Flag(os.O_CREATE)
)
