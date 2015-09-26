package wine

import (
	"github.com/justintan/gox"
)

type nodeType int

const (
	staticNode nodeType = 0
	paramNode  nodeType = 1
)

type node struct {
	t        nodeType
	path     string
	handlers []Handler
	children []*node
}

func (this *node) conflict(n *node) bool {
	if n.t == staticNode {
		if n.path != this.path {
			return false
		}
	} else if this.t == paramNode {
		if len(n.handlers) > 0 && len(this.handlers) > 0 {
			return true
		}
	}

	for _, c1 := range n.children {
		for _, c2 := range this.children {
			if c1.conflict(c2) {
				return true
			}
		}
	}

	return false
}

func (this *node) addChild(pathSegments []string, fullPath string, handlers ...Handler) {
	if len(pathSegments) == 0 || len(pathSegments[0]) == 0 {
		panic("path segments can not be empty")
	}

	n := &node{}
	segment := pathSegments[0]
	if segment[0] == ':' {
		n.t = paramNode
		n.path = segment[1:]
	} else {
		n.path = segment
	}

	if len(pathSegments) == 1 {
		n.handlers = handlers
	} else {
		n.addChild(pathSegments[1:], fullPath, handlers...)
	}

	for _, child := range this.children {
		if child.conflict(n) {
			panic("duplicate path " + fullPath)
		}
	}

	if n.t == staticNode {
		this.children = append([]*node{n}, this.children...)
	} else {
		this.children = append(this.children, n)
	}
}

func (this *node) match(pathSegments []string, fullPath string) (handlers []Handler, params map[string]string) {
	if len(pathSegments) == 0 {
		panic("path segments is empty")
	}

	segment := pathSegments[0]
	if this.t == staticNode && this.path != segment {
		return
	}

	if len(pathSegments) == 1 {
		handlers = this.handlers
	} else {
		for _, child := range this.children {
			handlers, params = child.match(pathSegments[1:], fullPath)
			if len(handlers) > 0 {
				break
			}
		}
	}

	if len(handlers) > 0 {
		if len(params) == 0 {
			params = map[string]string{}
		}

		if this.t == paramNode {
			params[this.path] = segment
		}
	}
	return
}

func (this *node) Print(method string, parentPath string) {
	var path string
	if this.t == staticNode {
		path = parentPath + "/" + this.path
	} else {
		path = parentPath + "/:" + this.path
	}

	path = cleanPath(path)

	if len(this.handlers) > 0 {
		var hNames string
		for _, h := range this.handlers {
			if len(hNames) != 0 {
				hNames += ", "
			}

			hNames += h.Name()
		}
		gox.LInfof("%-5s %s\t%s", method, path, hNames)
	}

	for _, n := range this.children {
		n.Print(method, path)
	}
}
