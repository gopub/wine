package router

import (
	"container/list"
	"net/http"
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

func (r *Route) Header() http.Header {
	return r.node.Header
}

func (r *Route) SetHeader(key, value string) {
	r.node.Header.Set(key, value)
}

func (r *Route) AddHeader(key, value string) {
	r.node.Header.Add(key, value)
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
