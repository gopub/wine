package storage

import (
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gopub/errors"
	"github.com/gopub/wine"
)

type FileReader struct {
	r Reader
}

func (r *FileReader) HandleRequest(ctx context.Context, req *wine.Request) wine.Responder {
	segments := strings.Split(req.Request().URL.Path, "/")
	var name string
	for i := len(segments) - 1; i >= 0; i-- {
		name = segments[i]
		if name != "" {
			break
		}
	}
	data, err := r.r.Read(ctx, name)
	if err != nil {
		return wine.Error(err)
	}
	return wine.File(data, name)
}

type FileWriter struct {
	w Writer
}

func (w *FileWriter) HandleRequest(ctx context.Context, req *wine.Request) wine.Responder {
	if req.Request().MultipartForm != nil {
		return w.saveMultipart(ctx, req.Request().MultipartForm)
	}
	return w.saveBody(ctx, req.Body())
}

func (w *FileWriter) saveBody(ctx context.Context, body []byte) wine.Responder {
	o, err := NewObject(body)
	if err != nil {
		return wine.Error(err)
	}
	url, err := w.w.Write(ctx, o)
	if err != nil {
		return wine.Error(err)
	}
	return wine.JSON(http.StatusOK, []string{url})
}

func (w *FileWriter) saveMultipart(ctx context.Context, form *multipart.Form) wine.Responder {
	urls := make([]string, 0, 1)
	for _, fileHeaders := range form.File {
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				return wine.Error(err)
			}

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return wine.Error(err)
			}
			o, err := NewObject(b)
			if err != nil {
				return wine.Error(err)
			}
			url, err := w.w.Write(ctx, o)
			if err != nil {
				return wine.Error(err)
			}
			urls = append(urls, url)
		}
	}
	if len(urls) == 0 {
		return errors.BadRequest("missing image")
	}
	return wine.JSON(http.StatusOK, urls)
}
