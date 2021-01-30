package storage

import (
	"github.com/google/uuid"
	"github.com/gopub/errors"
	"github.com/gopub/wine/httpvalue"
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
		o.Type = httpvalue.DetectContentType(o.Content)
	}

	if o.Name == "" {
		o.Name = uuid.NewString()
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
