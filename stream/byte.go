package stream

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"

	"github.com/gopub/gox"
	"github.com/gopub/log"
	"github.com/gopub/wine"
	"github.com/gopub/wine/api"
	"github.com/gopub/wine/mime"
)

const packetHeadLen = 4

type ByteReadCloser interface {
	Read() (packet []byte, err error)
	io.Closer
}

type ByteWriteCloser interface {
	Write(packet []byte) error
	io.Closer
}

type byteReadCloser struct {
	body  io.ReadCloser
	buf   *bytes.Buffer
	block []byte
	err   error
}

func newByteReadCloser(body io.ReadCloser) *byteReadCloser {
	r := new(byteReadCloser)
	r.body = body
	r.buf = new(bytes.Buffer)
	r.block = make([]byte, 1024)
	return r
}

func (r *byteReadCloser) Read() (packet []byte, err error) {
	for {
		p := r.readPacket()
		if p != nil {
			return p, nil
		}
		if r.err != nil {
			return nil, r.err
		}
		n, err := r.body.Read(r.block)
		if n > 0 {
			r.buf.Write(r.block[:n])
		}
		if err != nil {
			r.err = err
		}
	}
}

func (r *byteReadCloser) Close() error {
	return r.body.Close()
}

func (r *byteReadCloser) readPacket() []byte {
	if r.buf.Len() < packetHeadLen {
		return nil
	}
	head := r.buf.Bytes()[:packetHeadLen]
	n := int(binary.BigEndian.Uint32(head))
	if r.buf.Len() < n+packetHeadLen {
		return nil
	}
	b := r.buf.Next(n + packetHeadLen)
	return b[packetHeadLen:]
}

type byteWriteCloser struct {
	w    http.ResponseWriter
	done chan<- interface{}
}

func newByteWriteCloser(w http.ResponseWriter, done chan<- interface{}) *byteWriteCloser {
	return &byteWriteCloser{
		w:    w,
		done: done,
	}
}

func (w *byteWriteCloser) Write(p []byte) error {
	head := make([]byte, packetHeadLen)
	binary.BigEndian.PutUint32(head, uint32(len(p)))
	err := gox.WriteAll(w.w, head)
	if err != nil {
		return fmt.Errorf("write header: %w", err)
	}
	err = gox.WriteAll(w.w, p)
	if err != nil {
		return fmt.Errorf("write packet: %w", err)
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

func (w *byteWriteCloser) Close() error {
	close(w.done)
	return nil
}

func NewByteReader(client *http.Client, req *http.Request) (ByteReadCloser, error) {
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
	r := newByteReadCloser(resp.Body)
	greeting, err := r.Read()
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("handshake: %w", err)
	}

	if s := string(greeting); s != Greeting {
		r.Close()
		return nil, fmt.Errorf("expect %s, got %s", Greeting, s)
	}
	return r, nil
}

func NewByteHandler(serve func(context.Context, ByteWriteCloser)) wine.Handler {
	return wine.HandlerFunc(func(ctx context.Context, req *wine.Request, next wine.Invoker) wine.Responder {
		logger := log.FromContext(ctx)
		logger.Debugf("Start")
		defer logger.Debugf("Closed")
		w := wine.GetResponseWriter(ctx)
		w.Header().Set(mime.ContentType, mime.OctetStream)
		done := make(chan interface{})
		bw := newByteWriteCloser(w, done)
		err := bw.Write([]byte(Greeting))
		if err != nil {
			logger.Errorf("Handshake: %v", err)
			return wine.Status(http.StatusOK)
		}
		go serve(ctx, bw)
		<-done
		return wine.Status(http.StatusOK)
	})
}
