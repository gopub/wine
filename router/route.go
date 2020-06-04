package router

import (
	"strings"
)

type Route struct {
	Scope string
	node  *node
}

func (r *Route) Path() string {
	return r.node.path
}

func (r *Route) SetDescription(s string) {
	r.node.Description = s
}

func (r *Route) Description() string {
	return r.node.Description
}

func (r *Route) HandlerPath() string {
	return r.node.HandlerPath()
}

type sortRouteList []*Route

func (l sortRouteList) Len() int {
	return len(l)
}

func (l sortRouteList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l sortRouteList) Less(i, j int) bool {
	return strings.Compare(l[i].node.path, l[j].node.path) < 0
}
