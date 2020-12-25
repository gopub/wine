package vfs

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gopub/log"

	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/types"
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
		fs.home = newDirInfo("")
		buf := bytes.NewBuffer(nil)
		if err = gob.NewEncoder(buf).Encode(fs.home); err != nil {
			return fmt.Errorf("encode: %w", err)
		}

		data = buf.Bytes()
		if err = fs.EncryptPage(data); err != nil {
			return fmt.Errorf("encrypt: %w", err)
		}
		if err = storage.Put(keyFSHome, data); err != nil {
			return fmt.Errorf("put %s: %w", keyFSHome, err)
		}
		return nil
	}

	if err = fs.DecryptPage(data); err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	buf := bytes.NewBuffer(data)
	if err = gob.NewDecoder(buf).Decode(&fs.home); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	return nil
}

func (fs *FileSystem) CreateFile(dir *FileInfo, name string) (*File, error) {
	return fs.CreateMIMEFile(dir, name, "", nil)
}

func (fs *FileSystem) CreateMIMEFile(dir *FileInfo, name, mimeType string, location *types.Point) (*File, error) {
	name = strings.TrimSpace(name)
	if !validateFileName(name) {
		return nil, errors.New("invalid file name")
	}
	if dir == nil {
		dir = fs.home
	}
	if !dir.IsDir() {
		return nil, errors.New("not directory")
	}
	if dir.parent == nil && dir != fs.home {
		return nil, errors.New("unknown dir")
	}
	name = dir.DistinctName(name)
	f := newFileInfo(name, mimeType, location)
	dir.AddSub(f)
	err := fs.Save()
	if err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return newFile(fs, f, true), nil
}

func (fs *FileSystem) CreateDir(dir *FileInfo, name string) (*File, error) {
	name = strings.TrimSpace(name)
	if !validateFileName(name) {
		return nil, errors.New("invalid file name")
	}
	if dir == nil {
		dir = fs.home
	}
	if !dir.IsDir() {
		return nil, errors.New("not directory")
	}
	if dir.parent == nil && dir != fs.home {
		return nil, errors.New("unknown dir")
	}
	name = dir.DistinctName(name)
	f := newDirInfo(name)
	dir.AddSub(f)
	err := fs.Save()
	if err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return newFile(fs, f, true), nil
}

func (fs *FileSystem) OpenFile(path string, write bool) (*File, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return newFile(fs, fs.home, write), nil
	}
	fi := fs.home.GetByPath(path)
	if fi == nil {
		return nil, os.ErrNotExist
	}
	return newFile(fs, fi, write), nil
}

func (fs *FileSystem) Open(path string) (http.File, error) {
	return fs.OpenFile(path, false)
}

func (fs *FileSystem) Delete(f *FileInfo) error {
	if f == fs.home {
		return errors.New("cannot delete home")
	}
	if f.parent == nil {
		return nil
	}
	f.parent.RemoveSub(f)
	if err := fs.Save(); err != nil {
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

func (fs *FileSystem) Move(f *FileInfo, dir *FileInfo) error {
	if !dir.IsDir() {
		return errors.New("dst is not dir")
	}
	if f.parent == dir {
		return nil
	}
	dir.AddSub(f)
	err := fs.Save()
	return errors.Wrapf(err, "cannot save")
}

func (fs *FileSystem) Save() error {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(fs.home); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	if err := fs.storage.Put(keyFSHome, buf.Bytes()); err != nil {
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
