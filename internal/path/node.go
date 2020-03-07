package path

import (
	"container/list"
	"strings"

	"github.com/gopub/gox"

	"github.com/gopub/log"
)

type NodeType int

const (
	StaticNode   NodeType = 0 // /users
	ParamNode    NodeType = 1 // /users/{id}
	WildcardNode NodeType = 2 // /users/{id}/photos/*
)

func (n NodeType) String() string {
	switch n {
	case StaticNode:
		return "StaticNode"
	case ParamNode:
		return "ParamNode"
	case WildcardNode:
		return "WildcardNode"
	default:
		return ""
	}
}

func getNodeType(segment string) NodeType {
	switch {
	case IsStatic(segment):
		return StaticNode
	case IsParam(segment):
		return ParamNode
	case IsWildcard(segment):
		return WildcardNode
	default:
		logger.Panicf("Invalid segment: %s" + segment)
		// Suppress compiling error because compiler doesn't know logger.Panicf equals to built-in panic function
		return WildcardNode
	}
}

type Node struct {
	typ       NodeType
	path      string // E.g. /items/{id}
	segment   string // E.g. items or {id}
	paramName string // E.g. id
	handlers  *list.List
	children  []*Node
}

func NewNodeList(path string, handlers *list.List) *Node {
	path = Normalize(path)
	segments := strings.Split(path, "/")
	var head, p *Node
	for i, s := range segments {
		path := strings.Join(segments[:i+1], "/")
		node := NewNode(path, s)
		if p != nil {
			p.children = []*Node{node}
		} else {
			head = node
		}
		p = node
	}
	if p != nil {
		p.handlers = handlers
	}
	return head
}

func NewNode(path, segment string) *Node {
	if len(strings.Split(segment, "/")) > 1 {
		logger.Panicf("Invalid segment: " + segment)
	}
	n := &Node{
		typ:      getNodeType(segment),
		path:     path,
		segment:  segment,
		handlers: list.New(),
	}
	switch n.typ {
	case ParamNode:
		n.paramName = segment[1 : len(segment)-1]
	case WildcardNode:
		n.segment = segment[1:]
	default:
		break
	}
	return n
}

func NewEmptyNode() *Node {
	return &Node{
		typ: StaticNode,
	}
}

func (n *Node) Type() NodeType {
	return n.typ
}

func (n *Node) Path() string {
	return n.path
}

func (n *Node) IsEndpoint() bool {
	return n.handlers != nil && n.handlers.Len() > 0
}

func (n *Node) ListEndpoints() []*Node {
	var l []*Node
	if n.IsEndpoint() {
		l = append(l, n)
	}

	for _, child := range n.children {
		l = append(l, child.ListEndpoints()...)
	}
	return l
}

func (n *Node) Handlers() *list.List {
	return n.handlers
}

func (n *Node) SetHandlers(l *list.List) {
	if n.handlers != nil {
		logger.Panicf("Cannot set again")
	}
	n.handlers = l
}

func (n *Node) Conflict(node *Node) *gox.Pair {
	if n.typ != node.typ {
		return nil
	}

	switch n.typ {
	case StaticNode:
		if n.segment != node.segment {
			return nil
		}

		if n.IsEndpoint() && node.IsEndpoint() {
			return &gox.Pair{
				First:  n,
				Second: node,
			}
		}
	case ParamNode:
		if n.IsEndpoint() && node.IsEndpoint() {
			return &gox.Pair{
				First:  n,
				Second: node,
			}
		}
	case WildcardNode:
		return &gox.Pair{
			First:  n,
			Second: node,
		}
	}
	for _, a := range n.children {
		for _, b := range node.children {
			if v := a.Conflict(b); v != nil {
				return v
			}
		}
	}
	return nil
}

func (n *Node) Add(node *Node) {
	var match *Node
	for _, child := range n.children {
		if v := child.Conflict(node); v != nil {
			log.Panicf("Conflict: %s, %s", v.First.(*Node).path, v.Second.(*Node).path)
		}

		if child.segment == node.segment {
			match = child
			break
		}
	}

	// Match: reuse the same node and append new nodes
	if match != nil {
		if len(node.children) == 0 {
			match.handlers = node.handlers
			return
		}

		for _, child := range node.children {
			match.Add(child)
		}
		return
	}

	// Mismatch: append new nodes
	switch node.typ {
	case StaticNode:
		n.children = append([]*Node{node}, n.children...)
	case ParamNode:
		i := len(n.children) - 1
		for i >= 0 {
			if n.children[i].typ != WildcardNode {
				break
			}
			i--
		}

		if i < 0 {
			n.children = append([]*Node{node}, n.children...)
		} else if i == len(n.children)-1 {
			n.children = append(n.children, node)
		} else {
			n.children = append(n.children, node)
			copy(n.children[i+2:], n.children[i+1:])
			n.children[i+1] = node
		}
	case WildcardNode:
		n.children = append(n.children, node)
	default:
		logger.Panicf("Invalid node type: %v", node.typ)
	}
}

func (n *Node) Match(segments ...string) (*Node, map[string]string) {
	if len(segments) == 0 {
		if n.typ == WildcardNode {
			return n, nil
		}
		return nil, nil
	}

	first := segments[0]
	switch n.typ {
	case StaticNode:
		if n.segment != first {
			return nil, nil
		}
		if len(segments) == 1 {
			if n.IsEndpoint() {
				return n, nil
			}
			// Perhaps some child nodes are wildcard Node which can match empty node
			for _, child := range n.children {
				if child.typ == WildcardNode {
					return child, nil
				}
			}
			return nil, nil
		}
		if segments[1] == "" && n.IsEndpoint() {
			return n, nil
		}
		for _, child := range n.children {
			match, params := child.Match(segments[1:]...)
			if match != nil {
				return match, params
			}
		}
	case ParamNode:
		var match *Node
		var params map[string]string
		if len(segments) == 1 || (segments[1] == "" && n.IsEndpoint()) {
			match = n
		} else {
			for _, child := range n.children {
				match, params = child.Match(segments[1:]...)
				if match != nil {
					break
				}
			}
		}

		if match != nil && match.IsEndpoint() {
			if params == nil {
				params = map[string]string{}
			}
			params[n.paramName] = first
			return match, params
		}
	case WildcardNode:
		if n.IsEndpoint() {
			return n, nil
		}
	}
	return nil, nil
}
