package wine

import (
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/gopub/log"
	pathpkg "github.com/gopub/wine/internal/path"
)

var logger *log.Logger

func init() {
	logger = log.Default().Derive("Wine")
	logger.SetFlags(log.LstdFlags - log.Lfunction - log.Lshortfile)
	pathpkg.SetLogger(logger)
}

func Logger() *log.Logger {
	return logger
}

func JoinURL(segment ...string) string {
	p := path.Join(segment...)
	p = strings.Replace(p, ":/", "://", 1)
	return p
}

func NewUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
