package router

import (
	"container/list"
)

type Endpoint struct {
	Scope string
	node  *node
}

func (e *Endpoint) Path() string {
	return e.node.path
}

func (e *Endpoint) SetDescription(s string) *Endpoint {
	e.node.Description = s
	return e
}

func (e *Endpoint) Description() string {
	return e.node.Description
}

func (e *Endpoint) HandlerPath() string {
	return e.node.HandlerPath()
}

func (e *Endpoint) FirstHandler() *list.Element {
	return e.node.handlers.Front()
}

func (e *Endpoint) Model() interface{} {
	return e.node.Model
}

func (e *Endpoint) SetModel(m interface{}) *Endpoint {
	e.node.Model = m
	return e
}

func (e *Endpoint) Sensitive() bool {
	return e.node.Sensitive
}

func (e *Endpoint) SetSensitive(b bool) {
	e.node.Sensitive = b
}

func (e *Endpoint) Metadata() interface{} {
	return e.node.Metadata
}

func (e *Endpoint) SetMetadata(m interface{}) {
	e.node.Metadata = m
}
