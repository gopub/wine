package stream

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine"
	"github.com/gopub/wine/api"
	"github.com/gopub/wine/mime"
)

const textPacketDelimiter = 0x01

type TextReadCloser interface {
	Read() (string, error)
	io.Closer
}

type TextWriteCloser interface {
	Write(s string) error
	io.Closer
}

type textReadCloser struct {
	body  io.ReadCloser
	buf   *bytes.Buffer
	block []byte
	err   error
}

func newTextReadCloser(body io.ReadCloser) *textReadCloser {
	r := new(textReadCloser)
	r.body = body
	r.buf = new(bytes.Buffer)
	r.block = make([]byte, 1024)
	return r
}

func (r *textReadCloser) Read() (string, error) {
	for {
		p, ok := r.readPacket()
		if ok {
			return p, nil
		}
		if r.err != nil {
			return "", r.err
		}
		n, err := r.body.Read(r.block)
		if err != nil {
			r.err = err
			continue
		}
		r.buf.Write(r.block[:n])
	}
}

func (r *textReadCloser) Close() error {
	return r.body.Close()
}

func (r *textReadCloser) readPacket() (string, bool) {
	for _, b := range r.buf.Bytes() {
		if b == textPacketDelimiter {
			p, err := r.buf.ReadBytes(textPacketDelimiter)
			if err != nil {
				log.Errorf("Read bytes: %v", err)
				return "", false
			}
			// Exclude the last byte which is packet delimiter
			return string(p[:len(p)-1]), true
		}
	}
	return "", false
}

type textWriteCloser struct {
	w    http.ResponseWriter
	done chan<- interface{}
}

func newTextWriteCloser(w http.ResponseWriter, done chan<- interface{}) *textWriteCloser {
	return &textWriteCloser{
		w:    w,
		done: done,
	}
}

func (w *textWriteCloser) Write(s string) error {
	p := []byte(s)
	p = append(p, textPacketDelimiter)
	err := gox.WriteAll(w.w, p)
	if err != nil {
		return fmt.Errorf("write all: %w", err)
	}
	if flusher, ok := w.w.(http.Flusher); ok {
		flusher.Flush()
	}
	if e, ok := w.w.(interface{ Error() error }); ok {
		if e.Error() != nil {
			return e.Error()
		}
	}
	return nil
}

func (w *textWriteCloser) Close() error {
	close(w.done)
	return nil
}

func NewTextReader(client *http.Client, req *http.Request) (TextReadCloser, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		err = api.ParseResult(resp, nil, true)
		if err != nil {
			return nil, fmt.Errorf("parse result: %w", err)
		}
		return nil, gox.NewError(resp.StatusCode, "unknown error")
	}
	return newTextReadCloser(resp.Body), nil
}

func NewTextHandler(serve func(context.Context, TextWriteCloser)) wine.Handler {
	return wine.HandlerFunc(func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responsible {
		logger := log.FromContext(ctx)
		logger.Debugf("Receive stream")
		w := wine.GetResponseWriter(ctx)
		w.Header().Set(mime.ContentType, mime.Plain)
		done := make(chan interface{})
		go serve(ctx, newTextWriteCloser(w, done))
		<-done
		logger.Debugf("Close stream")
		return wine.Status(http.StatusOK)
	})
}
