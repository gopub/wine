package wine

import (
	"log"
	"reflect"
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

func newNodeList(path string, handlers ...Handler) []*node {
	path = normalizePath(path)
	segs := strings.Split(path, "/")
	nodes := make([]*node, len(segs))
	for i, s := range segs {
		nodes[i] = newNode(s)
	}

	nodes[len(nodes)-1].handlers = handlers
	return nodes
}

func newNode(pathSegment string) *node {
	if len(strings.Split(pathSegment, "/")) > 1 {
		panic("invalid path segment: " + pathSegment)
	}
	n := &node{
		t:    getNodeType(pathSegment),
		path: pathSegment,
	}
	switch n.t {
	case paramNode:
		n.paramNames = strings.Split(pathSegment, ",")
		for i, pn := range n.paramNames {
			n.paramNames[i] = pn[1:]
		}
	case wildcardNode:
		n.path = pathSegment[1:]
	default:
		break
	}
	return n
}

func (n *node) add(nodes []*node) bool {
	var matchNode *node
	for _, cn := range n.children {
		nod := nodes[0]
		if cn.t == wildcardNode || nod.t == wildcardNode {
			return false
		}

		if cn.t != nod.t {
			continue
		}

		//cn.t == node.t
		if cn.path == nod.path {
			if len(cn.handlers) > 0 && len(nod.handlers) > 0 {
				return false
			} else {
				matchNode = cn
				break
			}
		}

		if cn.t == paramNode && len(cn.handlers) > 0 && len(nod.handlers) > 0 && len(cn.paramNames) == len(nod.paramNames) {
			return false
		}
	}

	if matchNode != nil {
		if len(nodes) > 1 {
			return matchNode.add(nodes[1:])
		} else {
			matchNode.handlers = nodes[0].handlers
		}
	} else {
		nod := nodes[0]
		for i := 1; i < len(nodes); i++ {
			nod.children = []*node{nodes[i]}
			nod = nodes[i]
		}

		nod = nodes[0]
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

	return true
}

func (n *node) match(pathSegments []string) ([]Handler, map[string]string) {
	if len(pathSegments) == 0 {
		panic("[WINE] pathSegments is empty")
	}

	//log.Println(n.t, n.path, n.paramNames, n.children, "==", pathSegments)
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
				handlers, params := child.match(pathSegments[1:])
				if len(handlers) > 0 {
					return handlers, params
				}
			}
			return nil, nil
		}
	case paramNode:
		var handlers []Handler
		var params map[string]string
		if len(pathSegments) == 1 || (pathSegments[1] == "" && len(n.handlers) > 0) {
			handlers = n.handlers
		} else {
			for _, child := range n.children {
				handlers, params = child.match(pathSegments[1:])
				if len(handlers) > 0 {
					break
				}
			}
		}

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
		}
		return handlers, params
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

	path = normalizePath(path)

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
