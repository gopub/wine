package path

import (
	"regexp"

	"github.com/gopub/log"
)

var logger = log.Default()

func SetLogger(l *log.Logger) {
	logger = l
}

var (
	compactSlashRegexp = regexp.MustCompile(`/{2,}`)
	staticPathRegexp   = regexp.MustCompile(`^[^\\{\\}\\*]+$`)
	wildcardPathRegexp = regexp.MustCompile(`^*[0-9a-zA-Z_\\-]*$`)
	paramPathRegexp    = regexp.MustCompile(`^{([a-zA-Z][a-zA-Z_0-9]*|_[a-zA-Z_0-9]*[a-zA-Z0-9]+[a-zA-Z_0-9]*)}$`)
)

func Normalize(p string) string {
	p = compactSlashRegexp.ReplaceAllString(p, "/")
	if p == "" {
		return p
	}

	if p[0] == '/' {
		p = p[1:]
	}

	if len(p) > 1 && p[len(p)-1] == '/' {
		p = p[:len(p)-1]
	}
	return p
}

func IsStatic(p string) bool {
	return staticPathRegexp.MatchString(p)
}

func IsWildcard(p string) bool {
	if p == "" {
		return false
	}
	return wildcardPathRegexp.MatchString(p)
}

func IsParam(p string) bool {
	if p == "" {
		return false
	}
	return paramPathRegexp.MatchString(p)
}
