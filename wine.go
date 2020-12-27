package wine

import (
	"net/url"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/gopub/conv"
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

func JoinURL(segments ...string) string {
	escapes := make([]string, len(segments))
	for i, v := range segments {
		escapes[i] = url.PathEscape(v)
	}
	p := path.Join(escapes...)
	p = strings.Replace(p, ":/", "://", 1)
	return p
}

func NewUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

type LogStringer interface {
	LogString() string
}

type Validator = conv.Validator

func Validate(i interface{}) error {
	return conv.Validate(i)
}
