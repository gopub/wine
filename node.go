package wine

import (
	"github.com/justintan/gox"
	"strings"
)

type nodeType int

const (
	staticNode   nodeType = 0 // /users
	paramNode    nodeType = 1 // /users/:id
	wildcardNode nodeType = 2 // /users/:id/photos/*
)

type node struct {
	t        nodeType
	path     string
	handlers []Handler
	children []*node
}

func (n *node) conflict(nod *node) bool {
	if nod.t != n.t {
		return false
	}

	switch nod.t {
	case staticNode:
		if nod.path != n.path {
			return false
		} else {
			if len(nod.handlers) > 0 && len(n.handlers) > 0 {
				return true
			}
		}
	case paramNode:
		if len(nod.handlers) > 0 && len(n.handlers) > 0 {
			return true
		}
	case wildcardNode:
		return true
	default:
		panic("unknown node type")
	}

	for _, c1 := range nod.children {
		for _, c2 := range n.children {
			if c1.conflict(c2) {
				return true
			}
		}
	}

	return false
}

func (n *node) addChild(pathSegments []string, fullPath string, handlers ...Handler) {
	if len(pathSegments) == 0 || len(pathSegments[0]) == 0 {
		panic("path segments can not be empty")
	}

	nod := &node{}
	segment := pathSegments[0]
	if segment[0] == ':' {
		nod.t = paramNode
		nod.path = segment[1:]
		if len(nod.path) == 0 {
			panic("invalid path: " + fullPath)
		}
	} else if segment[0] == '*' {
		nod.t = wildcardNode
		nod.path = segment[1:]
		if len(pathSegments) > 1 {
			panic("wildcard only in the end segement")
		}
	} else {
		nod.path = segment
	}

	if len(pathSegments) == 1 {
		nod.handlers = handlers
	} else {
		nod.addChild(pathSegments[1:], fullPath, handlers...)
	}

	for _, child := range n.children {
		if child.conflict(nod) {
			panic("duplicate path " + fullPath)
		}
	}

	switch nod.t {
	case staticNode:
		n.children = append([]*node{nod}, n.children...)
		break
	case paramNode:
		i := len(n.children) - 1
		for i >= 0 {
			if n.children[i].t != wildcardNode {
				break
			}
			i--
		}

		if i < 0 {
			n.children = append([]*node{nod}, n.children...)
		} else if i == len(n.children)-1 {
			n.children = append(n.children, nod)
		} else {
			n.children = append(n.children, nod)
			copy(n.children[i+2:], n.children[i+1:])
			n.children[i+1] = nod
		}
		break
	default:
		n.children = append(n.children, nod)
		break
	}
}

func (n *node) match(pathSegments []string, fullPath string) (handlers []Handler, params map[string]string) {
	if len(pathSegments) == 0 {
		panic("path segments is empty")
	}

	segment := pathSegments[0]
	if n.t == staticNode && n.path != segment {
		return
	}

	if len(pathSegments) == 1 {
		handlers = n.handlers
	} else {
		for _, child := range n.children {
			handlers, params = child.match(pathSegments[1:], fullPath)
			if len(handlers) > 0 {
				break
			}
		}
	}

	//consider wildcard in the end
	if len(handlers) == 0 && n.t == wildcardNode {
		handlers = n.handlers
		segment = strings.Join(pathSegments, "/")
	}

	if len(handlers) > 0 { //matched
		if len(params) == 0 {
			params = map[string]string{}
		}

		if n.t == paramNode {
			params[n.path] = segment
		}
	}
	return
}

func (n *node) Print(method string, parentPath string) {
	var path string
	if n.t == staticNode {
		path = parentPath + "/" + n.path
	} else if n.t == paramNode {
		path = parentPath + "/:" + n.path
	} else {
		path = parentPath + "/*" + n.path
	}

	path = cleanPath(path)

	if len(n.handlers) > 0 {
		var hNames string
		for _, h := range n.handlers {
			if len(hNames) != 0 {
				hNames += ", "
			}

			if f, ok := h.(HandlerFunc); ok {
				hNames += gox.GetFuncName(f)
			} else {
				hNames += gox.GetTypeName(h)
			}
		}
		gox.LInfof("%-5s %s\t%s", method, path, hNames)
	}

	for _, nod := range n.children {
		nod.Print(method, path)
	}
}
