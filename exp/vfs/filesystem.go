package vfs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/log"
)

type FileSystem struct {
	storage KVStorage
	home    *FileInfo
	key     []byte
}

var _ http.FileSystem = (*FileSystem)(nil)

func NewFileSystem(storage KVStorage) (*FileSystem, error) {
	return NewEncryptedFileSystem(storage, "")
}

func NewEncryptedFileSystem(storage KVStorage, password string) (*FileSystem, error) {
	key, err := loadKey(storage, password)
	if err != nil {
		return nil, fmt.Errorf("load key: %w", err)
	}

	fs := &FileSystem{
		storage: storage,
		key:     key,
	}

	if err = fs.mountHome(storage); err != nil {
		return nil, fmt.Errorf("mount home: %w", err)
	}
	return fs, nil
}

func loadKey(storage KVStorage, password string) ([]byte, error) {
	credentials, err := storage.Get(keyFSCredentials)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("get key: %w", err)
		}

		if password == "" {
			return nil, nil
		}

		// this is a new file system, initialize key if password is provided
		passHash := conv.Hash32([]byte(password))
		key := conv.Hash32([]byte(uuid.New().String()))

		credentials = make([]byte, 64)
		copy(credentials, key[:])
		copy(credentials[32:], passHash[:])

		// encrypt
		if err = conv.AES(credentials, passHash[:], passHash[:16]); err != nil {
			return nil, fmt.Errorf("encrypt: %w", err)
		}

		if err = storage.Put(keyFSCredentials, credentials); err != nil {
			return nil, fmt.Errorf("put: %w", err)
		}
		return key[:], nil
	}

	if password == "" {
		return nil, errors.New("missing password")
	}
	if len(credentials) != keySize {
		return nil, errors.New("corrupted file system")
	}

	passHash := conv.Hash32([]byte(password))
	if err = conv.AES(credentials, passHash[:], passHash[:16]); err != nil {
		return nil, fmt.Errorf("decript: %w", err)
	}

	if !bytes.Equal(credentials[32:], passHash[:]) {
		return nil, errors.New("invalid password")
	}
	return credentials[:32], nil
}

func (fs *FileSystem) mountHome(storage KVStorage) error {
	if fs.home != nil {
		log.Panic("Cannot mount home twice")
	}
	data, err := storage.Get(keyFSHome)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("get %s: %w", keyFSHome, err)
		}

		// initialize
		fs.home = newFileInfo(true, "")
		if err = fs.SaveFileTree(); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		return nil
	}

	if err = fs.DecryptPage(data); err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&fs.home); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

func (fs *FileSystem) Create(parentUUID string, isDir bool, name string) (*File, error) {
	name = strings.TrimSpace(name)
	if !validateFileName(name) {
		return nil, errors.New("invalid file name")
	}
	var dir *FileInfo
	if parentUUID == "" {
		dir = fs.home
	} else {
		dir = fs.home.GetByUUID(parentUUID)
		if dir == nil || !dir.IsDir() {
			return nil, errors.New("cannot find parent directory")
		}
	}
	name = dir.DistinctName(name)
	f := newFileInfo(isDir, name)
	dir.AddSub(f)
	err := fs.SaveFileTree()
	if err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return newFile(fs, f, true), nil
}

func (fs *FileSystem) OpenByUUID(id string, write bool) (*File, error) {
	if id == "" {
		return newFile(fs, fs.home, write), nil
	}
	fi := fs.home.GetByUUID(id)
	if fi == nil {
		return nil, os.ErrNotExist
	}
	if fi.busy {
		return nil, errors.New("busy")
	}
	return newFile(fs, fi, write), nil
}

func (fs *FileSystem) OpenByPath(path string, write bool) (*File, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return newFile(fs, fs.home, write), nil
	}
	fi := fs.home.GetByPath(path)
	if fi == nil {
		return nil, os.ErrNotExist
	}
	if fi.busy {
		return nil, errors.New("busy")
	}
	return newFile(fs, fi, write), nil
}

func (fs *FileSystem) Open(path string) (http.File, error) {
	return fs.OpenByPath(path, false)
}

func (fs *FileSystem) Delete(uuid string) error {
	f := fs.home.GetByUUID(uuid)
	if f == nil {
		return nil
	}
	if f == fs.home {
		return errors.New("cannot delete home")
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

func (fs *FileSystem) Move(uuid, parentUUID string) error {
	dir, err := fs.StatByUUID(parentUUID)
	if err != nil {
		return fmt.Errorf("stat by parentUUID %s: %w", parentUUID, err)
	}
	if !dir.IsDir() {
		return errors.New("dst is not dir")
	}
	f, err := fs.StatByUUID(uuid)
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

func (fs *FileSystem) SaveFileTree() error {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(fs.home); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	data := buf.Bytes()
	if err := fs.EncryptPage(data); err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := fs.storage.Put(keyFSHome, data); err != nil {
		return fmt.Errorf("put: %w", err)
	}
	return nil
}

func (fs *FileSystem) EncryptPage(data []byte) error {
	if len(fs.key) == 0 {
		return nil
	}
	return conv.AES(data, fs.key, fs.key[:16])
}

func (fs *FileSystem) DecryptPage(data []byte) error {
	if len(fs.key) == 0 {
		return nil
	}
	return conv.AES(data, fs.key, fs.key[:16])
}

func (fs *FileSystem) StatByUUID(uuid string) (*FileInfo, error) {
	if uuid == "" {
		return fs.home, nil
	}
	f := fs.home.GetByUUID(uuid)
	if f == nil {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (fs *FileSystem) CreateAndWrite(parentUUID, name string, data []byte) (*FileInfo, error) {
	f, err := fs.Create(parentUUID, false, name)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return nil, fmt.Errorf("write data: %w", err)
	}
	f.info.SetMIMEType(http.DetectContentType(data))
	return f.info, nil
}

func (fs *FileSystem) CreateAndCopy(parentUUID, srcFilePath string) (*FileInfo, error) {
	name := filepath.Base(srcFilePath)
	f, err := fs.Create(parentUUID, false, name)
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
