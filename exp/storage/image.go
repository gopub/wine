package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/gopub/errors"
	"github.com/gopub/wine"
	"github.com/gopub/wine/httpvalue"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)

type ThumbnailOption struct {
	Width   int
	Height  int
	Quality int
}

func (t *ThumbnailOption) Validate() error {
	if t.Width <= 0 {
		return fmt.Errorf("negative width %d", t.Width)
	}
	if t.Height <= 0 {
		return fmt.Errorf("negative height %d", t.Height)
	}
	if t.Quality <= 0 || t.Quality > 100 {
		return fmt.Errorf("invalid quality %d, expected (0, 100]", t.Quality)
	}
	return nil
}

type ImageWriter struct {
	w          Writer
	thumbnails []*ThumbnailOption
}

var _ wine.Handler = (*ImageWriter)(nil)

func NewImageWriter(w Writer) *ImageWriter {
	return &ImageWriter{
		w: w,
	}
}

func (w *ImageWriter) AddThumbnailOptions(thumbnails ...*ThumbnailOption) error {
	var err error
	for _, t := range thumbnails {
		if er := t.Validate(); er != nil {
			err = errors.Append(err, er)
		} else {
			w.thumbnails = append(w.thumbnails, t)
		}
	}
	return err
}

func (w *ImageWriter) Write(ctx context.Context, name string, data []byte) (string, error) {
	if name = strings.TrimSpace(name); name == "" {
		name = uuid.NewString()
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

func (w *ImageWriter) thumbnail(ctx context.Context, img image.Image, name string, t *ThumbnailOption) (string, error) {
	data, err := getImageThumbnail(img, t)
	if err != nil {
		return "", fmt.Errorf("get image thumbnail: %w", err)
	}
	obj := &Object{
		Name:    name,
		Content: data,
		Type:    httpvalue.JPEG,
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
	url, err := w.Write(ctx, uuid.NewString(), body)
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

			url, err := w.Write(ctx, uuid.NewString(), b)
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

func getImageThumbnail(img image.Image, t *ThumbnailOption) ([]byte, error) {
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
		return nil, fmt.Errorf("encode image to jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

func GenerateThumbnail(origin []byte, w, h int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(origin))
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return nil, errors.BadRequest("decode: %v", err)
		}
		return nil, fmt.Errorf("decode: %w", err)
	}
	return getImageThumbnail(img, &ThumbnailOption{Width: w, Height: h, Quality: 100})
}
