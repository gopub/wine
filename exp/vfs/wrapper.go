package vfs

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gopub/errors"
)

type FileSystemWrapper FileSystem

func (w *FileSystemWrapper) Mkdir(parentUUID, dirName string) (*FileInfo, error) {
	fs := (*FileSystem)(w)
	if dirName == "" {
		return nil, os.ErrInvalid
	}
	var parent *FileInfo
	if parentUUID == "" {
		parent = fs.root
	} else {
		parent = fs.root.GetByUUID(parentUUID)
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
		dir = fs.root
	} else {
		dir = fs.root.GetByUUID(dirUUID)
		if dir == nil {
			return nil, fmt.Errorf("dir %s does not exist", dirUUID)
		}
	}
	return fs.Create(filepath.Join(dir.Path(), name))
}

func (w *FileSystemWrapper) Open(uuid string, flag Flag) (*File, error) {
	fs := (*FileSystem)(w)
	if uuid == "" {
		return newFile(fs, fs.root, flag), nil
	}
	fi := fs.root.GetByUUID(uuid)
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
	f := fs.root.GetByUUID(uuid)
	if f == nil {
		return nil
	}
	if f == fs.root {
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
		return fs.root, nil
	}
	f := fs.root.GetByUUID(uuid)
	if f == nil {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (w *FileSystemWrapper) ReadAll(uuid string) ([]byte, error) {
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
	f.info.SetMIMEType(http.DetectContentType(data))
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

func (w *FileSystemWrapper) CopyDiskFile(dirUUID, srcFilePath string) (*FileInfo, error) {
	fs := (*FileSystem)(w)
	name := filepath.Base(srcFilePath)
	f, err := w.Create(dirUUID, name)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", srcFilePath, err)
	}
	if _, err = io.Copy(f, srcFile); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}
	me, err := mimetype.DetectFile(srcFilePath)
	if err != nil {
		return nil, fmt.Errorf("detet file: %w", err)
	}
	f.info.SetMIMEType(me.String())
	if err = fs.SaveFileTree(); err != nil {
		return nil, fmt.Errorf("save file tree: %w", err)
	}
	return f.info, nil
}
