package router

import (
	"container/list"
	"strings"
)

type Route struct {
	Scope string
	node  *node
}

func (r *Route) Path() string {
	return r.node.path
}

func (r *Route) SetDescription(s string) *Route {
	r.node.Description = s
	return r
}

func (r *Route) Description() string {
	return r.node.Description
}

func (r *Route) HandlerPath() string {
	return r.node.HandlerPath()
}

func (r *Route) FirstHandler() *list.Element {
	return r.node.handlers.Front()
}

func (r *Route) Model() interface{} {
	return r.node.Model
}

func (r *Route) SetModel(m interface{}) *Route {
	r.node.Model = m
	return r
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
