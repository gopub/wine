package path

import (
	"net/http"
	"path"
	"regexp"
	"strings"
)

var (
	_compactSlashRegexp = regexp.MustCompile(`/{2,}`)
	_staticPathRegexp   = regexp.MustCompile(`^[^\\{\\}\\*]+$`)
	_wildcardPathRegexp = regexp.MustCompile(`^*[0-9a-zA-Z_\\-]*$`)
	_paramPathRegexp    = regexp.MustCompile(`^{([a-zA-Z][a-zA-Z_0-9]*|_[a-zA-Z_0-9]*[a-zA-Z0-9]+[a-zA-Z_0-9]*)}$`)
)

func Normalize(p string) string {
	p = _compactSlashRegexp.ReplaceAllString(p, "/")
	if len(p) == 0 {
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
	return _staticPathRegexp.MatchString(p)
}

func IsWildcard(p string) bool {
	if len(p) == 0 {
		return false
	}
	return _wildcardPathRegexp.MatchString(p)
}

func IsParam(p string) bool {
	if len(p) == 0 {
		return false
	}
	return _paramPathRegexp.MatchString(p)
}

func Join(segment ...string) string {
	p := path.Join(segment...)
	p = strings.Replace(p, ":/", "://", 1)
	return p
}

func NormalizeRequestPath(req *http.Request) string {
	p := req.RequestURI
	i := strings.Index(p, "?")
	if i > 0 {
		p = req.RequestURI[:i]
	}

	return Normalize(p)
}
