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

func Normalize(path string) string {
	path = _compactSlashRegexp.ReplaceAllString(path, "/")
	if len(path) == 0 {
		return path
	}

	if path[0] == '/' {
		path = path[1:]
	}

	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

func IsStatic(path string) bool {
	return _staticPathRegexp.MatchString(path)
}

func IsWildcard(path string) bool {
	if len(path) == 0 {
		return false
	}
	return _wildcardPathRegexp.MatchString(path)
}

func IsParam(path string) bool {
	if len(path) == 0 {
		return false
	}
	return _paramPathRegexp.MatchString(path)
}

func Join(segment ...string) string {
	res := path.Join(segment...)
	res = strings.Replace(res, ":/", "://", 1)
	return res
}

func NormalizeRequestPath(req *http.Request) string {
	path := req.RequestURI
	i := strings.Index(path, "?")
	if i > 0 {
		path = req.RequestURI[:i]
	}

	return Normalize(path)
}
