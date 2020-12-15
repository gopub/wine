package respond

import (
	"context"
	"io"
	"net/http"

	"github.com/gopub/log"
	"github.com/gopub/wine/httpvalue"
)

// StreamFile creates a application/octet-stream response
func StreamFile(r io.Reader, name string) Func {
	return Func(func(ctx context.Context, w http.ResponseWriter) {
		logger := log.FromContext(ctx)
		w.Header().Set(httpvalue.ContentType, httpvalue.OctetStream)
		if name != "" {
			w.Header().Set(httpvalue.ContentDisposition, httpvalue.FileAttachment(name))
		}
		const size = 1024
		buf := make([]byte, size)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				if _, wErr := w.Write(buf[:n]); wErr != nil {
					logger.Errorf("Write: %v", wErr)
					return
				}
			}
			if err != nil {
				logger.Errorf("Read: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	})
}

// File creates a application/octet-stream response
func File(b []byte, name string) Func {
	return func(ctx context.Context, w http.ResponseWriter) {
		w.Header().Set(httpvalue.ContentType, httpvalue.OctetStream)
		if name != "" {
			w.Header().Set(httpvalue.ContentDisposition, httpvalue.FileAttachment(name))
		}
		if _, err := w.Write(b); err != nil {
			log.FromContext(ctx).Errorf("write: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// StaticFile serves static files
func StaticFile(req *http.Request, filePath string) Func {
	return Func(func(ctx context.Context, w http.ResponseWriter) {
		http.ServeFile(w, req, filePath)
	})
}

func Image(contentType string, content []byte) Func {
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}
	return func(ctx context.Context, w http.ResponseWriter) {
		w.Header().Set(httpvalue.ContentType, contentType)
		if _, err := w.Write(content); err != nil {
			log.FromContext(ctx).Errorf("write: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
