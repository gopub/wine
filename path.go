package wine

import (
	"regexp"
)

var compactSlashRegexp = regexp.MustCompile("/{2,}")
var staticPathRegexp = regexp.MustCompile("^[^:\\*]+$")
var wildcardPathRegexp = regexp.MustCompile("^\\*[0-9a-zA-Z_\\-]*$")
var paramPathRegexp = regexp.MustCompile("^:[a-zA-Z_]([a-zA-Z_0-9]+)*(,:[a-zA-Z_]([a-zA-Z_0-9]+,)*)*$")

func normalizePath(path string) string {
	if path == "" {
		return "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	if len(path) > 1 {
		if path[len(path)-1] == '/' {
			path = path[:len(path)-1]
		}

		//Why??
		//if path[len(path)-1] != '/' {
		//	path += "/"
		//}
	}

	path = compactSlashRegexp.ReplaceAllString(path, "/")
	return path
}

func isStaticPath(path string) bool {
	return staticPathRegexp.MatchString(path)
}

func isWildcardPath(path string) bool {
	if len(path) == 0 {
		return false
	}
	return wildcardPathRegexp.MatchString(path)
}

func isParamPath(path string) bool {
	if len(path) == 0 || path == ":_" { //make regex pattern simple
		return false
	}
	return paramPathRegexp.MatchString(path)
}
