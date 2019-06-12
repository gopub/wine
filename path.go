package wine

import (
	"regexp"
)

var (
	_compactSlashRegexp = regexp.MustCompile("/{2,}")
	_staticPathRegexp   = regexp.MustCompile("^[^\\{\\}\\*]+$")
	_wildcardPathRegexp = regexp.MustCompile("^\\*[0-9a-zA-Z_\\-]*$")
	_paramPathRegexp    = regexp.MustCompile("^\\{([a-zA-Z][a-zA-Z_0-9]*|_[a-zA-Z_0-9]*[a-zA-Z0-9]+[a-zA-Z_0-9]*)\\}$")
)

func normalizePath(path string) string {
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

func isStaticPath(path string) bool {
	return _staticPathRegexp.MatchString(path)
}

func isWildcardPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	return _wildcardPathRegexp.MatchString(path)
}

func isParamPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	return _paramPathRegexp.MatchString(path)
}

func getNodeType(path string) nodeType {
	switch {
	case isStaticPath(path):
		return StaticNode
	case isParamPath(path):
		return ParamNode
	case isWildcardPath(path):
		return WildcardNode
	default:
		panic("invalid path: " + path)
	}
}
