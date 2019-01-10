package wine

import (
	"github.com/gopub/log"
	"reflect"
	"runtime"
	"strings"
)

type nodeType int

const (
	_StaticNode   nodeType = 0 // /users
	_ParamNode    nodeType = 1 // /users/{id}
	_WildcardNode nodeType = 2 // /users/{id}/photos/*
)

func (n nodeType) String() string {
	switch n {
	case _StaticNode:
		return "_StaticNode"
	case _ParamNode:
		return "_ParamNode"
	case _WildcardNode:
		return "_WildcardNode"
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
		t:        getNodeType(pathSegment),
		path:     pathSegment,
		handlers: &handlerList{},
	}
	switch n.t {
	case _ParamNode:
		n.paramName = pathSegment[1 : len(pathSegment)-1]
	case _WildcardNode:
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
	if n.t == _WildcardNode || nod.t == _WildcardNode {
		return true
	}

	if n.t != nod.t {
		return false
	}

	//n.t == node.t

	if n.t == _StaticNode {
		return n.path == nod.path && !n.handlers.Empty() && !nod.handlers.Empty()
	}

	if n.t == _ParamNode {
		if !n.handlers.Empty() && !nod.handlers.Empty() {
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
	case _StaticNode:
		n.children = append([]*node{nod}, n.children...)
		break
	case _ParamNode:
		i := len(n.children) - 1
		for i >= 0 {
			if n.children[i].t != _WildcardNode {
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
	case _WildcardNode:
		n.children = append(n.children, nod)
		break
	default:
		log.Panic("invalid node type")
	}

	return true
}

func (n *node) matchPath(pathSegments []string) (*handlerList, map[string]string) {
	if len(pathSegments) == 0 {
		log.Panic("pathSegments is empty")
	}

	segment := pathSegments[0]
	switch n.t {
	case _StaticNode:
		switch {
		case n.path != segment:
			return nil, nil
		case len(pathSegments) == 1:
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
	case _ParamNode:
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

		if !handlers.Empty() {
			if params == nil {
				params = map[string]string{}
			}

			params[n.paramName] = segment
		}
		return handlers, params
	case _WildcardNode:
		return n.handlers, nil
	default:
		return nil, nil
	}
}

func (n *node) matchNodes(nodes []*node) *node {
	if len(nodes) == 0 {
		log.Panic("nodes is empty")
	}

	nod := nodes[0]
	if nod.t != n.t {
		return nil
	}

	if n.t == _StaticNode && n.path != nod.path {
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

func (n *node) Print(method string, parentPath string) {
	var path string
	switch n.t {
	case _StaticNode:
		path = parentPath + "/" + n.path
	case _ParamNode:
		path = parentPath + "/" + n.path
	default:
		path = parentPath + "/*" + n.path
	}

	path = normalizePath(path)

	if !n.handlers.Empty() {
		var hNames string
		for h := n.handlers.Head(); h != nil; h = h.next {
			if len(hNames) != 0 {
				hNames += ", "
			}

			var handlerName string
			if f, ok := h.handler.(HandlerFunc); ok {
				handlerName = runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
			} else {
				handlerName = reflect.TypeOf(h).Name()
			}

			if ShortHandlerNameFlag {
				hNames += getShortFileName(handlerName)
			} else {
				hNames += handlerName
			}
		}
		log.Infof("%-5s %s\t%s", method, path, hNames)
	}

	for _, nod := range n.children {
		nod.Print(method, path)
	}
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
	for i := 1; i < len(names)-1; i++ {
		names[i] = names[i][0:1]
	}
	filename = strings.Join(names, "/")
	return filename
}
