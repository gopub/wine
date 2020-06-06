package wine

import (
	"encoding/json"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/gopub/log"
	"github.com/gopub/wine/internal/respond"
	"github.com/gopub/wine/router"
)

var logger *log.Logger

func init() {
	logger = log.Default().Derive("Wine")
	logger.SetFlags(log.LstdFlags - log.Lfunction - log.Lshortfile)
	router.SetLogger(logger)
	respond.SetLogger(logger)
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

func JSONString(i interface{}) string {
	b, _ := json.Marshal(i)
	return string(b)
}
