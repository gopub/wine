package wine

import (
	"path"
	"reflect"
	"runtime"
	"strings"

	"github.com/gopub/log"
	pathutil "github.com/gopub/wine/internal/path"
)

type nodeType int

const (
	staticNode   nodeType = 0 // /users
	paramNode    nodeType = 1 // /users/{id}
	wildcardNode nodeType = 2 // /users/{id}/photos/*
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
	t         nodeType
	path      string
	paramName string
	handlers  *handlerList
	children  []*node
}

func newNodeList(path string, handlers *handlerList) []*node {
	path = pathutil.Normalize(path)
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
		t:        getNodeType(pathSegment),
		path:     pathSegment,
		handlers: &handlerList{},
	}
	switch n.t {
	case paramNode:
		n.paramName = pathSegment[1 : len(pathSegment)-1]
	case wildcardNode:
		n.path = pathSegment[1:]
	default:
		break
	}
	return n
}

func (n *node) conflict(nodes []*node) bool {
	if len(nodes) == 0 {
		return false
	}

	nod := nodes[0]
	// Allow wildcardNode and other kinds of nodes coexist
	//if n.t == wildcardNode || nod.t == wildcardNode {
	//	log.Errorf("Conflict: [%s %s], [%s %s]", n.t, n.path, nod.t, nod.path)
	//	return true
	//}

	if n.t != nod.t {
		return false
	}

	//n.t == node.t

	if n.t == staticNode {
		ok := n.path == nod.path && !n.handlers.Empty() && !nod.handlers.Empty()
		if ok {
			log.Errorf("Conflict: %v, %v", n.path, nod.path)
		}
		return ok
	}

	if n.t == paramNode {
		if !n.handlers.Empty() && !nod.handlers.Empty() {
			log.Errorf("Conflict: %s, %s", n.path, nod.path)
			return true
		}

		sn := nodes[1:]
		for _, cn := range n.children {
			if cn.conflict(sn) {
				return true
			}
		}
	}

	return false
}

func (n *node) add(nodes []*node) bool {
	var matchNode *node
	for _, cn := range n.children {
		if cn.conflict(nodes) {
			return false
		}

		if cn.path == nodes[0].path {
			matchNode = cn
			break
		}
	}

	if matchNode != nil {
		if len(nodes) > 1 {
			return matchNode.add(nodes[1:])
		}
		matchNode.handlers = nodes[0].handlers
		return true
	}

	nod := nodes[0]
	for i := 1; i < len(nodes); i++ {
		nod.children = []*node{nodes[i]}
		nod = nodes[i]
	}

	nod = nodes[0]
	switch nod.t {
	case staticNode:
		n.children = append([]*node{nod}, n.children...)
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
	case wildcardNode:
		n.children = append(n.children, nod)
	default:
		logger.Panic("invalid node type")
	}

	return true
}

func (n *node) matchPath(pathSegments []string) (*handlerList, map[string]string) {
	if len(pathSegments) == 0 {
		if n.t == wildcardNode {
			return n.handlers, nil
		}
		return nil, nil
	}

	segment := pathSegments[0]
	switch n.t {
	case staticNode:
		switch {
		case n.path != segment:
			return nil, nil
		case len(pathSegments) == 1:
			if n.handlers.Empty() {
				// Perhaps some child nodes are wildcard node, which could match empty path
				for _, child := range n.children {
					handlers, params := child.matchPath(pathSegments[1:])
					if !handlers.Empty() {
						return handlers, params
					}
				}
			}
			return n.handlers, nil
		case pathSegments[1] == "" && !n.handlers.Empty():
			return n.handlers, nil
		default:
			for _, child := range n.children {
				handlers, params := child.matchPath(pathSegments[1:])
				if !handlers.Empty() {
					return handlers, params
				}
			}
			return nil, nil
		}
	case paramNode:
		var handlers *handlerList
		var params map[string]string
		if len(pathSegments) == 1 || (pathSegments[1] == "" && !n.handlers.Empty()) {
			handlers = n.handlers
		} else {
			for _, child := range n.children {
				handlers, params = child.matchPath(pathSegments[1:])
				if !handlers.Empty() {
					break
				}
			}
		}

		if handlers != nil && !handlers.Empty() {
			if params == nil {
				params = map[string]string{}
			}

			params[n.paramName] = segment
		}
		return handlers, params
	case wildcardNode:
		return n.handlers, nil
	default:
		return nil, nil
	}
}

func (n *node) matchNodes(nodes []*node) *node {
	if len(nodes) == 0 {
		logger.Panic("nodes is empty")
	}

	nod := nodes[0]
	if nod.t != n.t {
		return nil
	}

	if n.t == staticNode && n.path != nod.path {
		return nil
	}

	if len(nodes) == 1 {
		return n
	}

	childNodes := nodes[1:]
	for _, child := range n.children {
		if v := child.matchNodes(childNodes); v != nil {
			return v
		}
	}

	return nil
}

func (n *node) handlerNames() string {
	s := new(strings.Builder)
	for h := n.handlers.Head(); h != nil; h = h.next {
		if s.Len() > 0 {
			s.WriteString(", ")
		}

		var name string
		if f, ok := h.handler.(HandlerFunc); ok {
			name = runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		} else {
			name = reflect.TypeOf(h).Name()
		}

		if strings.HasSuffix(name, "-fm") {
			name = name[0 : len(name)-3]
		}

		if ShortHandlerNameFlag {
			s.WriteString(getShortFileName(name))
		} else {
			s.WriteString(name)
		}
	}
	return s.String()
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

	path = pathutil.Normalize(path)

	if !n.handlers.Empty() {
		logger.Infof("%-5s %s\t%s", method, path, n.handlerNames())
	}

	for _, nod := range n.children {
		nod.Print(method, path)
	}
}

func (n *node) Endpoints(method string) endpointInfoList {
	var list endpointInfoList
	var p string
	switch n.t {
	case staticNode, paramNode:
		p = "/" + n.path
	default:
		p = "/*" + n.path
	}

	p = pathutil.Normalize(p)
	if !n.handlers.Empty() {
		list = append(list, &endpointInfo{
			Method:       method,
			Path:         p,
			HandlerNames: n.handlerNames(),
		})
	}

	for _, nod := range n.children {
		subList := nod.Endpoints(method)
		for _, sp := range subList {
			sp.Path = path.Join(p, sp.Path)
			list = append(list, sp)
		}
	}

	return list
}

func getShortFileName(filename string) string {
	if len(log.PackagePath) > 0 {
		filename = strings.TrimPrefix(filename, log.PackagePath)
	} else {
		start := strings.Index(filename, log.GoSrc)
		if start > 0 {
			start += len(log.GoSrc)
			filename = filename[start:]
		}
	}

	names := strings.Split(filename, "/")
	for i := 0; i < len(names)-1; i++ {
		names[i] = names[i][0:1]
	}
	filename = strings.Join(names, "/")
	return filename
}

func getNodeType(path string) nodeType {
	switch {
	case pathutil.IsStatic(path):
		return staticNode
	case pathutil.IsParam(path):
		return paramNode
	case pathutil.IsWildcard(path):
		return wildcardNode
	default:
		panic("invalid path: " + path)
	}
}
