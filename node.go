package wine

import (
	"log"
	"reflect"
	"regexp"
	"runtime"
	"strings"
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
	t          nodeType
	path       string
	paramNames []string
	handlers   []Handler
	children   []*node
}

func isValidStaticNode(path string) bool {
	matched, _ := regexp.MatchString("[^:\\*]+", path)
	return matched
}

func isValidWildcardNode(path string) bool {
	matched, _ := regexp.MatchString("\\*[0-9a-zA-Z_\\-]+", path)
	return matched
}

func isValidParamNode(path string) bool {
	matched, _ := regexp.MatchString(":[a-zA-Z_]([a-zA-Z_0-9]+,)*", path)
	return matched
}

func (n *node) conflict(nod *node) bool {
	if nod.t != n.t {
		return false
	}

	switch nod.t {
	case staticNode:
		if nod.path != n.path {
			return false
		}

		if len(nod.handlers) > 0 && len(n.handlers) > 0 {
			return true
		}
	case paramNode:
		if len(nod.paramNames) != len(n.paramNames) {
			return false
		}

		if len(nod.handlers) > 0 && len(n.handlers) > 0 {
			return true
		}
	case wildcardNode:
		return true
	default:
		panic("[WINE] invalid node type")
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
	if len(pathSegments) == 0 {
		panic("[WINE] pathSegments is empty")
	}

	if n.t == wildcardNode {
		panic("[WINE] forbidden to add child to wildcardNode")
	}

	nod := &node{}
	segment := pathSegments[0]
	switch {
	case isValidParamNode(segment):
		nod.t = paramNode
		nod.path = segment
		nod.paramNames = strings.Split(segment, ",")
		for i, pn := range nod.paramNames {
			nod.paramNames[i] = pn[1:]
		}
	case isValidWildcardNode(segment):
		nod.t = wildcardNode
		nod.path = segment[1:]
		if len(pathSegments) > 1 {
			if pathSegments[1] == "" {
				pathSegments = pathSegments[0:1]
			} else {
				panic("[WINE] wildcard node only allowed at the end")
			}
		}
	case len(segment) == 0 || isValidStaticNode(segment):
		nod.t = staticNode
		nod.path = segment
	default:
		panic("[WINE] invalid path: " + fullPath)
	}

	if len(pathSegments) == 1 {
		nod.handlers = handlers
	} else {
		nod.addChild(pathSegments[1:], fullPath, handlers...)
	}

	for _, child := range n.children {
		if child.conflict(nod) {
			panic("[WINE] duplicate path: " + fullPath)
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
	case wildcardNode:
		n.children = append(n.children, nod)
		break
	default:
		panic("[WINE] invalid node type")
	}
}

func (n *node) match(pathSegments []string, fullPath string) ([]Handler, map[string]string) {
	if len(pathSegments) == 0 {
		panic("[WINE] pathSegments is empty")
	}

	segment := pathSegments[0]
	switch n.t {
	case staticNode:
		switch {
		case n.path != segment:
			return nil, nil
		case len(pathSegments) == 1:
			return n.handlers, nil
		case pathSegments[1] == "" && len(n.handlers) > 0:
			return n.handlers, nil
		default:
			for _, child := range n.children {
				handlers, params := child.match(pathSegments[1:], fullPath)
				if len(handlers) > 0 {
					return handlers, params
				}
			}
			return nil, nil
		}
	case paramNode:
		switch {
		case len(pathSegments) == 1:
			return n.handlers, map[string]string{n.path: segment}
		case pathSegments[1] == "" && len(n.handlers) > 0:
			return n.handlers, map[string]string{n.path: segment}
		default:
			for _, child := range n.children {
				handlers, params := child.match(pathSegments[1:], fullPath)
				if len(handlers) > 0 {
					if params == nil {
						params = map[string]string{}
					}

					segs := strings.Split(segment, ",")
					if len(segs) != len(n.paramNames) {
						return nil, nil
					}
					for i, s := range n.paramNames {
						params[s] = segs[i]
					}
					return handlers, params
				}
			}

			return nil, nil
		}
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
		path = parentPath + "/" + n.path
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
				hNames += runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
			} else {
				hNames += reflect.TypeOf(h).Name()
			}
		}
		log.Printf("[WINE] %-5s %s\t%s", method, path, hNames)
	}

	for _, nod := range n.children {
		nod.Print(method, path)
	}
}
