package wine

import (
	"net/url"
	"path"
	"strings"

	"github.com/gopub/wine/exp/vfs"

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
	vfs.SetLogger(logger)
}

func Logger() *log.Logger {
	return logger
}

func JoinURL(segments ...string) string {
	if len(segments) == 0 {
		return ""
	}
	p := path.Join(segments...)
	p = strings.Replace(p, ":/", "://", 1)
	i := strings.Index(p, "://")
	s := p
	if i >= 0 {
		s = p[i:]
		l := strings.Split(s, "/")
		for i, v := range l {
			l[i] = url.PathEscape(v)
		}
		p = p[:i] + path.Join(l...)
	}
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
