package vfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gopub/errors"
	"github.com/gopub/wine/httpvalue"
)

type FileSystemWrapper FileSystem

func (w *FileSystemWrapper) Mkdir(parentUUID, dirName string) (*FileInfo, error) {
	fs := (*FileSystem)(w)
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}
	if dirName == "" {
		return nil, os.ErrInvalid
	}
	var parent *FileInfo
	if parentUUID == "" {
		parent = fs.root.FileInfo
	} else {
		parent = fs.root.FindByUUID(parentUUID)
		if parent == nil {
			return nil, fmt.Errorf("parent %s does not exist", parentUUID)
		}
	}
	return fs.Mkdir(filepath.Join(parent.Path(), dirName))
}

func (w *FileSystemWrapper) Create(dirUUID, name string) (*File, error) {
	fs := (*FileSystem)(w)
	if name == "" {
		return nil, os.ErrInvalid
	}
	var dir *FileInfo
	if dirUUID == "" {
		dir = fs.root.FileInfo
	} else {
		dir = fs.root.FindByUUID(dirUUID)
		if dir == nil {
			return nil, fmt.Errorf("dir %s does not exist", dirUUID)
		}
	}
	return fs.Create(filepath.Join(dir.Path(), name))
}

func (w *FileSystemWrapper) Open(uuid string, flag Flag) (*File, error) {
	fs := (*FileSystem)(w)
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}
	if uuid == "" {
		return newFile(fs, fs.root.FileInfo, flag), nil
	}
	fi := fs.root.FindByUUID(uuid)
	if fi == nil {
		return nil, os.ErrNotExist
	}
	if fi.busy {
		return nil, os.ErrPermission
	}
	return newFile(fs, fi, flag), nil
}

func (w *FileSystemWrapper) Remove(uuid string) error {
	fs := (*FileSystem)(w)
	if !fs.AuthPassed() {
		return os.ErrPermission
	}
	f := fs.root.FindByUUID(uuid)
	if f == nil {
		return nil
	}
	if f == fs.root.FileInfo {
		return errors.New("cannot delete root")
	}
	if f.parent == nil {
		return nil
	}
	f.parent.RemoveSub(f)
	if err := fs.SaveFileTree(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	fileNodes := []*FileInfo{f}
	pages := f.Pages
	for i := 0; i < len(fileNodes); i++ {
		nod := fileNodes[i]
		fileNodes = append(fileNodes, nod.Files...)
		pages = append(pages, nod.Pages...)
	}
	var err error
	for _, page := range pages {
		er := fs.storage.Delete(page)
		if er != nil {
			err = errors.Append(err, er)
		}
	}
	return err
}

func (w *FileSystemWrapper) Move(uuid, dirUUID string) error {
	fs := (*FileSystem)(w)
	if !fs.AuthPassed() {
		return os.ErrPermission
	}
	dir, err := w.Stat(dirUUID)
	if err != nil {
		return fmt.Errorf("stat by dirUUID %s: %w", dirUUID, err)
	}
	if !dir.IsDir() {
		return errors.New("dst is not dir")
	}
	f, err := w.Stat(uuid)
	if err != nil {
		return fmt.Errorf("stat by uuid %s: %w", uuid, err)
	}
	if f.parent == dir {
		return nil
	}
	dir.AddSub(f)
	err = fs.SaveFileTree()
	return errors.Wrapf(err, "cannot save")
}

func (w *FileSystemWrapper) Stat(uuid string) (*FileInfo, error) {
	fs := (*FileSystem)(w)
	if uuid == "" {
		return fs.root.FileInfo, nil
	}
	f := fs.root.FindByUUID(uuid)
	if f == nil {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (w *FileSystemWrapper) Read(uuid string) ([]byte, error) {
	f, err := w.Open(uuid, ReadOnly)
	if err != nil {
		return nil, fmt.Errorf("open by uuid: %w", err)
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}
	return data, nil
}

func (w *FileSystemWrapper) Write(uuid string, data []byte) (*FileInfo, error) {
	f, err := w.Open(uuid, WriteOnly)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return nil, fmt.Errorf("flag data: %w", err)
	}
	f.info.SetMIMEType(httpvalue.DetectContentType(data))
	return f.info, nil
}

func (w *FileSystemWrapper) CreateAndWrite(dirUUID, name string, data []byte) (*FileInfo, error) {
	f, err := w.Create(dirUUID, name)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}
	return f.info, nil
}

func (w *FileSystemWrapper) ImportDiskFile(dirUUID, diskFilePath string) (*FileInfo, error) {
	fs := (*FileSystem)(w)
	cleanPath := filepath.Clean(diskFilePath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", cleanPath, err)
	}

	src, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", cleanPath, err)
	}

	abs, err := filepath.Abs(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("get absolute path %s: %w", cleanPath, err)
	}
	base := filepath.Base(abs)
	if !info.IsDir() {
		dst, err := w.Create(dirUUID, base)
		if err != nil {
			return nil, fmt.Errorf("create file: %w", err)
		}
		defer dst.Close()
		if _, err = io.Copy(dst, src); err != nil {
			return nil, fmt.Errorf("copy: %w", err)
		}
		me, err := mimetype.DetectFile(cleanPath)
		if err != nil {
			return nil, fmt.Errorf("detet file type: %w", err)
		}
		dst.info.SetMIMEType(me.String())
		if err = fs.SaveFileTree(); err != nil {
			return nil, fmt.Errorf("save file tree: %w", err)
		}
		return dst.info, nil
	}

	dir, err := w.Mkdir(dirUUID, base)
	if err != nil {
		return nil, fmt.Errorf("create dir %s: %w", base, err)
	}

	names, err := src.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("read dir names: %w", err)
	}

	for _, name := range names {
		_, er := w.ImportDiskFile(dir.UUID(), filepath.Join(cleanPath, name))
		if er != nil {
			err = errors.Append(err, er)
		}
	}
	return dir, nil
}
