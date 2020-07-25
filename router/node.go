package router

import (
	"container/list"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gopub/log"
	"github.com/gopub/types"
)

type nodeType int

const (
	staticNode   nodeType = iota // /users
	paramNode                    // /users/{id}
	wildcardNode                 // /users/{id}/photos/*
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

func getNodeType(segment string) nodeType {
	switch {
	case IsStatic(segment):
		return staticNode
	case IsParam(segment):
		return paramNode
	case IsWildcard(segment):
		return wildcardNode
	default:
		log.Panicf("Invalid segment: %s" + segment)
		// Suppress compiling error because compiler doesn't know log.Panicf equals to built-in panic function
		return wildcardNode
	}
}

type node struct {
	typ       nodeType
	path      string // E.g. /items/{id}
	segment   string // E.g. items or {id}
	paramName string // E.g. id
	handlers  *list.List
	children  []*node

	Model       interface{}
	Description string

	Metadata interface{}
}

func newNodeList(path string, handlers *list.List) *node {
	path = Normalize(path)
	segments := strings.Split(path, "/")
	var head, p *node
	for i, s := range segments {
		path := strings.Join(segments[:i+1], "/")
		n := NewNode(path, s)
		if p != nil {
			p.children = []*node{n}
		} else {
			head = n
		}
		p = n
	}
	if p != nil {
		p.handlers = handlers
	}
	return head
}

func NewNode(path, segment string) *node {
	if len(strings.Split(segment, "/")) > 1 {
		log.Panicf("Invalid segment: " + segment)
	}
	n := &node{
		typ:      getNodeType(segment),
		path:     path,
		segment:  segment,
		handlers: list.New(),
	}
	switch n.typ {
	case paramNode:
		n.paramName = segment[1 : len(segment)-1]
	case wildcardNode:
		n.segment = segment[1:]
	default:
		break
	}
	return n
}

func NewEmptyNode() *node {
	return &node{
		typ: staticNode,
	}
}

func (n *node) Type() nodeType {
	return n.typ
}

func (n *node) Path() string {
	return n.path
}

func (n *node) IsEndpoint() bool {
	return n.handlers != nil && n.handlers.Len() > 0
}

func (n *node) ListEndpoints() []*node {
	var l []*node
	if n.IsEndpoint() {
		l = append(l, n)
	}

	for _, child := range n.children {
		l = append(l, child.ListEndpoints()...)
	}
	return l
}

func (n *node) Handlers() *list.List {
	return n.handlers
}

func (n *node) SetHandlers(l *list.List) {
	if n.handlers != nil {
		log.Panicf("Cannot set again")
	}
	n.handlers = l
}

func (n *node) Conflict(node *node) *types.Pair {
	if n.typ != node.typ {
		return nil
	}

	switch n.typ {
	case staticNode:
		if n.segment != node.segment {
			return nil
		}

		if n.IsEndpoint() && node.IsEndpoint() {
			return &types.Pair{
				First:  n,
				Second: node,
			}
		}
	case paramNode:
		if n.IsEndpoint() && node.IsEndpoint() {
			return &types.Pair{
				First:  n,
				Second: node,
			}
		}
	case wildcardNode:
		return &types.Pair{
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

func (n *node) Add(nod *node) {
	var match *node
	for _, child := range n.children {
		if v := child.Conflict(nod); v != nil {
			log.Panicf("Conflict: %s, %s", v.First.(*node).path, v.Second.(*node).path)
		}

		if child.segment == nod.segment {
			match = child
			break
		}
	}

	// Match: reuse the same node and append new nodes
	if match != nil {
		if len(nod.children) == 0 {
			match.handlers = nod.handlers
			return
		}

		for _, child := range nod.children {
			match.Add(child)
		}
		return
	}

	// Mismatch: append new nodes
	switch nod.typ {
	case staticNode:
		n.children = append([]*node{nod}, n.children...)
	case paramNode:
		i := len(n.children) - 1
		for i >= 0 {
			if n.children[i].typ != wildcardNode {
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
		log.Panicf("Invalid node type: %v", nod.typ)
	}
}

func (n *node) MatchPath(path string) (*node, map[string]string) {
	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}
	return n.Match(segments...)
}

func (n *node) Match(segments ...string) (*node, map[string]string) {
	if len(segments) == 0 {
		if n.typ == wildcardNode {
			return n, nil
		}
		return nil, nil
	}

	first := segments[0]
	switch n.typ {
	case staticNode:
		if n.segment != first {
			return nil, nil
		}
		if len(segments) == 1 {
			if n.IsEndpoint() {
				return n, nil
			}
			// Perhaps some child nodes are wildcard node which can match empty node
			for _, child := range n.children {
				if child.typ == wildcardNode {
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
	case paramNode:
		var match *node
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
	case wildcardNode:
		if n.IsEndpoint() {
			return n, nil
		}
	}
	return nil, nil
}

func (n *node) HandlerPath() string {
	reg := regexp.MustCompile(`\(\*([a-zA-Z0-9_]+)\)`)
	s := new(strings.Builder)
	for p := n.handlers.Front(); p != nil; p = p.Next() {
		if s.Len() > 0 {
			s.WriteString(", ")
		}

		var name string
		if s, ok := p.Value.(fmt.Stringer); ok {
			name = s.String()
		} else {
			name = reflect.TypeOf(p.Value).Name()
		}

		if strings.HasSuffix(name, "-fm") {
			name = name[:len(name)-3]
		}
		name = reg.ReplaceAllString(name, "$1")
		s.WriteString(log.ShortPath(name))
	}
	return s.String()
}
