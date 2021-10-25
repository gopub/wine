package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gopub/log/v2"
	"github.com/gopub/wine"
	"github.com/gopub/wine/httpvalue"
)

type JSONReadCloser interface {
	Read(v interface{}) error
	io.Closer
}

type JSONWriteCloser interface {
	Write(v interface{}) error
	io.Closer
}

type jsonReadCloser struct {
	textReadCloser
}

func newJSONReadCloser(body io.ReadCloser) *jsonReadCloser {
	r := newTextReadCloser(body)
	return &jsonReadCloser{textReadCloser: *r}
}

func (r *jsonReadCloser) Read(v interface{}) error {
	p, err := r.textReadCloser.Read()
	if err != nil {
		return fmt.Errorf("read text: %w", err)
	}
	err = json.Unmarshal([]byte(p), v)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

type jsonWriteCloser struct {
	textWriteCloser
}

func newJSONWriteCloser(w http.ResponseWriter, done chan<- interface{}) *jsonWriteCloser {
	r := newTextWriteCloser(w, done)
	return &jsonWriteCloser{textWriteCloser: *r}
}

func (w *jsonWriteCloser) Write(v interface{}) error {
	p, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	err = w.textWriteCloser.Write(string(p))
	if err != nil {
		return fmt.Errorf("write text: %w", err)
	}
	return nil
}

func NewJSONReader(client *http.Client, req *http.Request) (JSONReadCloser, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if err := checkStatus(resp); err != nil {
		return nil, err
	}
	r := newJSONReadCloser(resp.Body)
	var greeting interface{}
	err = r.Read(&greeting)
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("handshake: %w", err)
	}
	if greeting != Greeting {
		r.Close()
		return nil, fmt.Errorf("expect %s, got %s", Greeting, greeting)
	}
	return r, nil
}

func NewJSONHandler(serve func(context.Context, JSONWriteCloser)) wine.Handler {
	return wine.ResponderFunc(func(ctx context.Context, w http.ResponseWriter) {
		logger := log.FromContext(ctx)
		logger.Debugf("Start")
		defer logger.Debugf("Closed")
		w.Header().Set(httpvalue.ContentType, httpvalue.JsonUTF8)
		done := make(chan interface{})
		jw := newJSONWriteCloser(w, done)
		err := jw.Write(Greeting)
		if err != nil {
			logger.Errorf("Handshake: %v", err)
			return
		}
		go serve(ctx, jw)
		<-done
		return
	})
}
