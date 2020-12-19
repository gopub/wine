package storage

import (
	"net/http"

	"github.com/gopub/errors"
	"github.com/gopub/wine"
)

type Object struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
	Type    string `json:"type"`
}

func (o *Object) Validate() error {
	if len(o.Content) == 0 {
		return errors.BadRequest("missing content")
	}

	if o.Type == "" {
		o.Type = http.DetectContentType(o.Content)
	}

	if o.Name == "" {
		o.Name = wine.NewUUID()
	}
	return nil
}

func NewObject(content []byte) (*Object, error) {
	o := &Object{
		Content: content,
	}
	if err := o.Validate(); err != nil {
		return nil, err
	}
	return o, nil
}
