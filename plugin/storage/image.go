package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gopub/wine"

	"github.com/disintegration/imaging"
	"github.com/gopub/errors"
	wm "github.com/gopub/wine/mime"
)

type Thumbnail struct {
	Width   int
	Height  int
	Quality int
}

type ImageWriter struct {
	w          Writer
	thumbnails []*Thumbnail
}

func NewImageWriter(w Writer) *ImageWriter {
	return &ImageWriter{
		w: w,
	}
}

func (w *ImageWriter) AddThumbnails(thumbnails ...*Thumbnail) {
	for i, t := range thumbnails {
		if t.Width <= 0 {
			panic(fmt.Sprintf("storage: thumbnails[%d].Width=%d", i, t.Width))
		}
		if t.Height <= 0 {
			panic(fmt.Sprintf("storage: thumbnails[%d].Height=%d", i, t.Height))
		}
		if t.Quality <= 0 {
			panic(fmt.Sprintf("storage: thumbnails[%d].Quality=%d, expect (0,100]", i, t.Quality))
		}
		if t.Quality > 100 {
			t.Quality = 100
		}
	}
	w.thumbnails = append(w.thumbnails, thumbnails...)
}

func (w *ImageWriter) Write(ctx context.Context, name string, data []byte) (string, error) {
	if name = strings.TrimSpace(name); name == "" {
		name = wine.NewUUID()
	}
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return "", errors.BadRequest("decode: %v", err)
		}
		return "", fmt.Errorf("decode: %w", err)
	}
	o := &Object{
		Name:     name,
		Content:  data,
		MIMEType: mime.TypeByExtension("." + format),
	}
	if err = o.Validate(); err != nil {
		return "", fmt.Errorf("validate: %w", err)
	}
	url, err := w.w.Write(ctx, o)
	if err != nil {
		return "", err
	}
	for _, t := range w.thumbnails {
		name := fmt.Sprintf("%s_%dx%d.%s", o.Name, t.Width, t.Height, format)
		if _, err = w.thumbnail(ctx, img, name, t); err != nil {
			return "", fmt.Errorf("thumbnail: %#v, %w", t, err)
		}
	}
	return url, nil
}

func (w *ImageWriter) thumbnail(ctx context.Context, img image.Image, name string, t *Thumbnail) (string, error) {
	dx := img.Bounds().Dx()
	dy := img.Bounds().Dy()

	var tImg image.Image
	if dx < t.Width || dy < t.Height {
		tImg = img
	} else {
		if dx*t.Height > dy*t.Width {
			tImg = imaging.Resize(img, 0, t.Height, imaging.Lanczos)
		} else {
			tImg = imaging.Resize(img, t.Width, 0, imaging.Lanczos)
		}
	}

	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, tImg, &jpeg.Options{Quality: t.Quality})
	if err != nil {
		return "", fmt.Errorf("encode image to jpeg: %w", err)
	}
	obj := &Object{
		Name:     name,
		Content:  buf.Bytes(),
		MIMEType: wm.JPEG,
	}
	return w.w.Write(ctx, obj)
}

type SaveImageHandler struct {
	w *ImageWriter
}

var _ wine.Handler = (*SaveImageHandler)(nil)

func NewImageHandler(w Writer) *SaveImageHandler {
	return &SaveImageHandler{w: NewImageWriter(w)}
}

func (h *SaveImageHandler) AddThumbnails(thumbnails ...*Thumbnail) {
	h.w.AddThumbnails(thumbnails...)
}

func (h *SaveImageHandler) HandleRequest(ctx context.Context, req *wine.Request) wine.Responder {
	if req.Request().MultipartForm != nil {
		return h.saveMultipart(ctx, req.Request().MultipartForm)
	}
	return h.saveBody(ctx, req.Body())
}

func (h *SaveImageHandler) saveBody(ctx context.Context, body []byte) wine.Responder {
	url, err := h.w.Write(ctx, wine.NewUUID(), body)
	if err != nil {
		return wine.Error(err)
	}
	return wine.JSON(http.StatusOK, []string{url})
}

func (h *SaveImageHandler) saveMultipart(ctx context.Context, form *multipart.Form) wine.Responder {
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

			url, err := h.w.Write(ctx, wine.NewUUID(), b)
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
