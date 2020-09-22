package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gopub/errors"
	"github.com/gopub/wine"
	mimepkg "github.com/gopub/wine/mime"
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

var _ wine.Handler = (*ImageWriter)(nil)

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
		Name:    name,
		Content: data,
		Type:    mime.TypeByExtension("." + format),
	}
	if err = o.Validate(); err != nil {
		return "", fmt.Errorf("validate: %w", err)
	}
	url, err := w.w.Write(ctx, o)
	if err != nil {
		return "", err
	}
	for _, t := range w.thumbnails {
		name := fmt.Sprintf("%s-%dx%d", o.Name, t.Width, t.Height)
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
	var obj *Object
	if path.Ext(name) == ".png" {
		if err := png.Encode(buf, tImg); err != nil {
			return "", fmt.Errorf("encode image to jpeg: %w", err)
		}
		obj = &Object{
			Name:    name,
			Content: buf.Bytes(),
			Type:    mimepkg.PNG,
		}
	} else {
		if err := jpeg.Encode(buf, tImg, &jpeg.Options{Quality: t.Quality}); err != nil {
			return "", fmt.Errorf("encode image to jpeg: %w", err)
		}
		obj = &Object{
			Name:    name,
			Content: buf.Bytes(),
			Type:    mimepkg.JPEG,
		}
	}
	return w.w.Write(ctx, obj)
}

func (w *ImageWriter) HandleRequest(ctx context.Context, req *wine.Request) wine.Responder {
	if req.Request().MultipartForm != nil {
		return w.saveMultipart(ctx, req.Request().MultipartForm)
	}
	return w.saveBody(ctx, req.Body())
}

func (w *ImageWriter) saveBody(ctx context.Context, body []byte) wine.Responder {
	url, err := w.Write(ctx, wine.NewUUID(), body)
	if err != nil {
		return wine.Error(err)
	}
	return wine.JSON(http.StatusOK, url)
}

func (w *ImageWriter) saveMultipart(ctx context.Context, form *multipart.Form) wine.Responder {
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

			url, err := w.Write(ctx, wine.NewUUID(), b)
			if err != nil {
				return wine.Error(err)
			}
			urls = append(urls, url)
		}
	}
	if len(urls) == 0 {
		return errors.BadRequest("missing image")
	}
	if len(urls) == 1 {
		return wine.JSON(http.StatusOK, urls[0])
	}
	return wine.JSON(http.StatusOK, urls)
}
