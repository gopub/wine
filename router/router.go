package router

import (
	"container/list"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
)

// Router implements routing function
type Router struct {
	scopedRoot map[string]*node
	basePath   string
	handlers   *list.List
}

// New new a Router
func New() *Router {
	r := &Router{
		scopedRoot: make(map[string]*node, 4),
		handlers:   list.New(),
	}
	r.scopedRoot[""] = NewEmptyNode()
	return r
}

func (r *Router) clone() *Router {
	nr := &Router{
		scopedRoot: r.scopedRoot,
		basePath:   r.basePath,
		handlers:   list.New(),
	}
	nr.handlers.PushBackList(r.handlers)
	return nr
}

func (r *Router) BasePath() string {
	return r.basePath
}

// Group returns a new router whose basePath is r.basePath+path
func (r *Router) Group(path string) *Router {
	if path == "/" {
		log.Panic(`Not allowed to create group "/"`)
	}

	nr := r.clone()
	// support empty path
	if len(path) > 0 {
		nr.basePath = Normalize(r.basePath + "/" + path)
	}
	return nr
}

// UseHandlers returns a new router with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router) Use(handlers *list.List) *Router {
	nr := r.clone()
	for h := handlers.Front(); h != nil; h = h.Next() {
		name := fmt.Sprint(h)
		found := false
		for e := nr.handlers.Front(); e != nil; e = e.Next() {
			if fmt.Sprint(e.Value) == name {
				found = true
				break
			}
		}

		if !found {
			nr.handlers.PushBack(h.Value)
		}
	}
	return nr
}

// Match finds handlers and parses path parameters according to method and path
func (r *Router) Match(scope string, path string) (*Endpoint, map[string]string) {
	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}

	root := r.scopedRoot[scope]
	global := r.scopedRoot[""]
	if root == nil {
		root = global
	}

	n, params := root.Match(segments...)
	if n == nil && root != global {
		n, params = global.Match(segments...)
	}

	if n == nil {
		return nil, map[string]string{}
	}

	unescaped := make(map[string]string, len(params))
	for k, v := range params {
		uv, err := url.PathUnescape(v)
		if err != nil {
			logger.Errorf("Unescape path param %s: %v", v, err)
			unescaped[k] = v
		} else {
			unescaped[k] = uv
		}
	}
	return &Endpoint{
		Scope: scope,
		node:  n,
	}, unescaped
}

func (r *Router) MatchScopes(path string) []string {
	var a []string
	for m := range r.scopedRoot {
		if rt, _ := r.Match(m, path); rt != nil {
			a = append(a, m)
		}
	}
	return a
}

// bind binds scope, path with handlers
func (r *Router) Bind(scope, path string, handlers *list.List) *Endpoint {
	if path == "" {
		log.Panic("path is empty")
	}

	if handlers == nil || handlers.Len() == 0 {
		log.Panic("handlers cannot be empty")
	}

	scope = strings.ToUpper(scope)
	handlers.PushFrontList(r.handlers)
	root := r.createRoot(scope)
	global := r.scopedRoot[""]
	path = Normalize(r.basePath + "/" + path)
	if path == "" {
		if root.IsEndpoint() {
			log.Panicf("Conflict: %s, %s", scope, r.basePath)
		}

		if global.IsEndpoint() {
			log.Panicf("Conflict: %s", r.basePath)
		}
		root.SetHandlers(handlers)
	} else {
		nl := newNodeList(path, handlers)
		if pair := global.Conflict(nl); pair != nil {
			first := pair.First.(*node).Path()
			second := pair.Second.(*node).Path()
			log.Panicf("Conflict: %s, %s %s", first, scope, second)
		}
		root.Add(nl)
	}
	n, _ := root.MatchPath(path)
	return &Endpoint{
		Scope: scope,
		node:  n,
	}
}

func (r *Router) createRoot(scope string) *node {
	root := r.scopedRoot[scope]
	if root == nil {
		root = NewEmptyNode()
		r.scopedRoot[scope] = root
	}
	return root
}

// Print prints all path trees
func (r *Router) Print() {
	for method, root := range r.scopedRoot {
		nodes := root.ListEndpoints()
		for _, n := range nodes {
			logger.Debugf("%-5s %s\t%s", method, n.Path(), n.HandlerPath())
		}
	}
}

func (r *Router) ListRoutes() []*Endpoint {
	l := make([]*Endpoint, 0, 10)
	for scope, root := range r.scopedRoot {
		for _, e := range root.ListEndpoints() {
			l = append(l, &Endpoint{
				Scope: scope,
				node:  e,
			})
		}
	}
	sort.Sort(sortRouteList(l))
	return l
}

func (r *Router) ContainsHandler(h interface{}) bool {
	s := fmt.Sprint(h)
	for e := r.handlers.Front(); e != nil; e = e.Next() {
		if s == fmt.Sprint(e.Value) {
			return true
		}
	}
	return false
}

func (r *Router) Handlers() *list.List {
	return r.handlers
}
