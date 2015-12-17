package wine

import (
	"github.com/justintan/gox"
)

type nodeType int

const (
	staticNode   nodeType = 0 // /users
	paramNode    nodeType = 1 // /users/:id
	wildcardNode nodeType = 2 // /users/:id/photos/*
)

func (n nodeType) String() string {
	switch n {
	case staticNode:
		return "staticNode"
	case paramNode:
		return "paramNode"
	case wildcardNode:
		return "wildcardNode"
	default:
		return ""
	}
}

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
	switch segment[0] {
	case ':':
		nod.t = paramNode
		nod.path = segment[1:]
		if len(nod.path) == 0 {
			panic("invalid path: " + fullPath)
		}
	case '*':
		nod.t = wildcardNode
		nod.path = segment[1:]
		if len(pathSegments) > 1 {
			panic("wildcard only in the end segement")
		}
	default:
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

func (n *node) match(pathSegments []string, fullPath string) ([]Handler, map[string]string) {
	if len(pathSegments) == 0 {
		panic("path segments is empty")
	}

	//gox.LDebug(pathSegments, n.t, n.path, n.children[0].t, n.children[0].path, fullPath)

	segment := pathSegments[0]
	switch n.t {
	case staticNode:
		if n.path != segment {
			return nil, nil
		}

		if len(pathSegments) == 1 {
			return n.handlers, nil
		}

		for _, child := range n.children {
			handlers, params := child.match(pathSegments[1:], fullPath)
			if len(handlers) > 0 {
				return handlers, params
			}
		}

		return nil, nil
	case paramNode:
		if len(pathSegments) == 1 {
			return n.handlers, map[string]string{n.path:segment}
		}

		for _, child := range n.children {
			handlers, params := child.match(pathSegments[1:], fullPath)
			if len(handlers) > 0 {
				if params == nil {
					params = map[string]string{}
				}
				params[n.path] = segment
				return handlers, params
			}
		}

		return nil, nil
	case wildcardNode:
		return n.handlers, nil
	default:
		return nil, nil
	}
}

func (n *node) Print(method string, parentPath string) {
	var path string
	switch n.t {
	case staticNode:
		path = parentPath + "/" + n.path
	case paramNode:
		path = parentPath + "/:" + n.path
	default:
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
