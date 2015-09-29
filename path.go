package wine

import "strings"

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

	n := len(path)
	path = strings.Replace(path, "//", "/", -1)
	for n != len(path) {
		n = len(path)
		path = strings.Replace(path, "//", "/", -1)
	}

	return path
}

//过滤`//`
func filterSlashPath(path string) string{
	if path == ""{
		return "/"
	}
	for strings.Contains(path,"//") {
		path=strings.Replace(path,"//","/",-1)
	}

	return path
}
