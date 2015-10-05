package wine

import (
	"regexp"
)

var slashCleanRegexp = regexp.MustCompile("/{2}")

func cleanPath(path string) string {
	if path == "" {
		return "/"
	}

	if path[0] != '/' {
		path = "/" + path
	}

	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[0 : len(path)-1]
	}

	path = slashCleanRegexp.ReplaceAllString(path, "/")
	return path
}
