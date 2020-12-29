package vfs

import (
	"encoding/json"
	"github.com/gopub/types"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gopub/errors"
	"github.com/gopub/wine/httpvalue"
)

type fileResponseItem struct {
	UUID       string  `json:"uuid"`
	Name       string  `json:"name"`
	IsDir      bool    `json:"is_dir"`
	Files      []*File `json:"files,omitempty"`
	CreatedAt  int64   `json:"created_at"`
	ModifiedAt int64   `json:"modified_at"`
}

type fileSystemHandler FileSystem

func (h *fileSystemHandler) writeError(req *http.Request, rw http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	code := errors.GetCode(err)
	if httpvalue.IsValidStatus(code) {
		errors.Format(code, err.Error()).Respond(req.Context(), rw)
	} else {
		errors.InternalServerError(err.Error()).Respond(req.Context(), rw)
	}
}

func (h *fileSystemHandler) Upload(rw http.ResponseWriter, req *http.Request) {
	fs := (*FileSystem)(h)
	if err := req.ParseMultipartForm(int64(20 * types.MB)); err != nil {
		h.writeError(req, rw, err)
		return
	}
	if req.MultipartForm == nil {
		errors.BadRequest("expect multipart form").Respond(req.Context(), rw)
		return
	}
	if req.MultipartForm.File == nil {
		errors.BadRequest("no file").Respond(req.Context(), rw)
	}

	dirID := req.URL.Query().Get("diruuid")
	if dirID == "" {
		dirName := req.URL.Query().Get("dir")
		df, err := fs.Stat(dirName)
		if err != nil {
			h.writeError(req, rw, err)
			return
		}
		dirID = df.UUID()
	} else {
		_, err := fs.Wrapper().Stat(dirID)
		if err != nil {
			h.writeError(req, rw, err)
			return
		}
	}

	for _, fhs := range req.MultipartForm.File {
		for _, fh := range fhs {
			f, err := fh.Open()
			if err != nil {
				h.writeError(req, rw, err)
				return
			}

			name := filepath.Base(fh.Filename)
			if name == "" {
				name = time.Now().Format(time.RFC3339)
			}
			df, err := fs.Wrapper().Create(dirID, name)
			if err != nil {
				h.writeError(req, rw, err)
				f.Close()
				return
			}
			io.Copy(df, f)
			df.Close()
			f.Close()
			fs.SaveFileTree()
		}
	}
}

func (h *fileSystemHandler) Get(rw http.ResponseWriter, req *http.Request) {
	fs := (*FileSystem)(h)
	uuid := req.URL.Query().Get("uuid")
	f, err := fs.Wrapper().Stat(uuid)
	if err != nil {
		h.writeError(req, rw, err)
		return
	}

	if f.IsDir() {
		rw.Header().Set(httpvalue.ContentType, httpvalue.JsonUTF8)
		_, err = rw.Write(f.DirContent())
		if err != nil {
			logger.Errorf("Write data: %v", err)
		}
		return
	}
	req.URL.Path = f.Path()
	http.FileServer(fs).ServeHTTP(rw, req)
}

func (h *fileSystemHandler) Stat(rw http.ResponseWriter, req *http.Request) {
	fs := (*FileSystem)(h)
	uuid := req.URL.Query().Get("uuid")
	f, err := fs.Wrapper().Stat(uuid)
	if err != nil {
		h.writeError(req, rw, err)
		return
	}

	clone := func(fi *FileInfo) *FileInfo {
		new := new(FileInfo)
		*new = *fi
		new.Files = nil
		new.Pages = nil
		return new
	}

	info := clone(f)
	if f.IsDir() {
		info.Files = make([]*FileInfo, len(f.Files))
		for i, subFile := range f.Files {
			info.Files[i] = clone(subFile)
		}
	}
	data, err := json.Marshal(info)
	if err != nil {
		h.writeError(req, rw, err)
		return
	}
	rw.Header().Set(httpvalue.ContentType, httpvalue.JsonUTF8)
	_, err = rw.Write(data)
	if err != nil {
		logger.Errorf("Write data: %v", err)
		return
	}
}

func (h *fileSystemHandler) RunServer(addr string) {
	log := logger.With("addr", addr)
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", h.Upload)
	mux.HandleFunc("/get", h.Get)
	mux.HandleFunc("/stat", h.Stat)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Infof("HTTP server was closed")
		} else {
			log.Panicf("HTTP server was terminated: %v", err)
		}
	} else {
		log.Infof("HTTP server stopped")
	}
}
