package vfs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/types"
)

type FileSystem struct {
	storage    KVStorage
	root       *FileInfo
	credential []byte
	key        []byte
	configs    types.M
	pageSize   int64
}

var _ http.FileSystem = (*FileSystem)(nil)

func NewFileSystem(storage KVStorage) (*FileSystem, error) {
	fs := &FileSystem{
		storage:  storage,
		pageSize: DefaultPageSize,
		configs:  types.M{},
		root:     newFileInfo(true, ""),
	}

	var err error
	if fs.pageSize, err = loadPageSize(storage); err != nil {
		return nil, fmt.Errorf("load page size: %w", err)
	}

	if fs.credential, err = storage.Get(keyFSCredential); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read credential: %w", err)
	}

	if fs.IsEncrypted() {
		return fs, nil
	}

	if err = fs.loadConfig(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err = fs.loadRoot(); err != nil {
		return nil, fmt.Errorf("loadRoot root: %w", err)
	}
	return fs, nil
}

func (fs *FileSystem) Format(pageSize int64) error {
	if len(fs.root.Files) != 0 {
		return os.ErrInvalid
	}

	if pageSize < MinPageSize {
		return os.ErrInvalid
	}
	fs.pageSize = pageSize
	return savePageSize(fs.storage, pageSize)
}

func (fs *FileSystem) IsEncrypted() bool {
	return len(fs.credential) > 0
}

func (fs *FileSystem) SetPassword(password string) error {
	if fs.IsEncrypted() {
		return errors.New("cannot set password to encrypted file system")
	}
	// this is a new file system, initialize key if password is provided
	passwordHash := conv.Hash32([]byte(password))
	key := conv.Hash32([]byte(uuid.New().String()))
	fs.credential = make([]byte, 2*keySize)
	copy(fs.credential, key[:])
	copy(fs.credential[keySize:], passwordHash[:])

	// encrypt
	if err := conv.AES(fs.credential, passwordHash[:], passwordHash[:16]); err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}
	if err := fs.storage.Put(keyFSCredential, fs.credential); err != nil {
		return fmt.Errorf("flag credential: %w", err)
	}
	fs.key = key[:]

	if err := fs.SaveConfig(); err != nil {
		return fmt.Errorf("saveConfig: %w", err)
	}

	if err := fs.SaveFileTree(); err != nil {
		return fmt.Errorf("saveFileTree: %w", err)
	}
	return nil
}

func (fs *FileSystem) ChangePassword(old, new string) error {
	if !fs.Auth(old) {
		logger.Error("Cannot change vfs password")
		return os.ErrPermission
	}
	fs.credential = nil
	fs.key = nil
	if err := fs.SetPassword(new); err != nil {
		return fmt.Errorf("set password: %w", err)
	}
	logger.Debug("Changed vfs password")
	return nil
}

func (fs *FileSystem) AuthPassed() bool {
	if !fs.IsEncrypted() {
		return true
	}
	return len(fs.key) != 0
}

func (fs *FileSystem) Auth(password string) bool {
	if password == "" {
		logger.Error("Missing vfs password")
		return false
	}

	if len(fs.credential) == 0 {
		logger.Error("No vfs credential")
		return false
	}
	passwordHash := conv.Hash32([]byte(password))
	credential := make([]byte, len(fs.credential))
	copy(credential, fs.credential)
	// decrypt
	if err := conv.AES(credential, passwordHash[:], passwordHash[:16]); err != nil {
		logger.Errorf("Cannot decrypt: %v", err)
		return false
	}

	if !bytes.Equal(credential[keySize:], passwordHash[:]) {
		logger.Error("VFS Auth failed")
		return false
	}
	fs.key = credential[:keySize]

	if err := fs.loadConfig(); err != nil {
		logger.Errorf("Cannot load vfs config: %v", err)
		return false
	}

	if err := fs.loadRoot(); err != nil {
		logger.Errorf("Cannot load vfs root: %v", err)
		return false
	}
	logger.Info("VFS auth passed")
	return true
}

func (fs *FileSystem) loadRoot() error {
	data, err := fs.storage.Get(keyFSRootDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("get %s: %w", keyFSRootDir, err)
		}

		// initialize
		fs.root = newFileInfo(true, "")
		return nil
	}

	if err = fs.DecryptPage(data); err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	if err = json.Unmarshal(data, &fs.root); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

func (fs *FileSystem) PageSize() int64 {
	return fs.pageSize
}

func (fs *FileSystem) Size() int64 {
	return fs.root.totalSize()
}

func (fs *FileSystem) Mkdir(name string) (*FileInfo, error) {
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}

	name = cleanName(name)
	if name == "" {
		return nil, os.ErrInvalid
	}
	f := fs.root.FindByPath(name)
	if f != nil {
		if f.IsDir() {
			return f, nil
		}
		return nil, fmt.Errorf("%s is not directory", name)
	}
	segments := splitPath(name)
	if len(segments) == 1 {
		f = newFileInfo(true, fs.root.DistinctName(name))
		fs.root.AddSub(f)
		return f, fs.SaveFileTree()
	}

	dirSegments := segments[:len(segments)-1]
	dir := fs.root.find(dirSegments...)
	if dir == nil {
		return nil, fmt.Errorf("dir %s does not exist", filepath.Join(dirSegments...))
	}
	if !dir.IsDir() {
		return nil, fmt.Errorf("%s is not directory", filepath.Join(dirSegments...))
	}

	f = newFileInfo(true, dir.DistinctName(name))
	dir.AddSub(f)
	return f, fs.SaveFileTree()
}

func (fs *FileSystem) MkdirAll(path string) (*FileInfo, error) {
	paths := splitPath(path)
	for i := range paths {
		f, err := fs.Mkdir(filepath.Join(paths[:i+1]...))
		if err != nil {
			return nil, err
		}
		if i == len(paths)-1 {
			return f, nil
		}
	}
	return nil, os.ErrInvalid
}

func (fs *FileSystem) Create(name string) (*File, error) {
	return fs.OpenFile(name, WriteOnly|Create)
}

func (fs *FileSystem) OpenFile(name string, flag Flag) (*File, error) {
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}

	paths := splitPath(name)
	if len(paths) == 0 {
		return nil, os.ErrInvalid
	}

	var dir *FileInfo
	if len(paths) == 1 {
		dir = fs.root
	} else {
		var err error
		dirPath := filepath.Join(paths[:len(paths)-1]...)
		dir, err = fs.MkdirAll(dirPath)
		if err != nil {
			return nil, fmt.Errorf("cannot make dir %s: %w", dirPath, err)
		}
	}

	base := paths[len(paths)-1]
	f := dir.Find(base)
	if f == nil {
		if (flag & Create) == 0 {
			logger.Error(name, base, paths)
			return nil, os.ErrNotExist
		}
		f = newFileInfo(false, dir.DistinctName(base))
		dir.AddSub(f)
	}
	if err := fs.SaveFileTree(); err != nil {
		return nil, fmt.Errorf("save file tree: %w", err)
	}
	return newFile(fs, f, flag), nil
}

func (fs *FileSystem) Open(name string) (http.File, error) {
	return fs.OpenFile(name, ReadOnly)
}

func (fs *FileSystem) Remove(name string) error {
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	fi, err := fs.Stat(name)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	if (fi.IsDir() && len(fi.Files) > 0) || fi == fs.root {
		return os.ErrPermission
	}
	return fs.Wrapper().Remove(fi.UUID())
}

func (fs *FileSystem) RemoveAll(path string) error {
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	fi, err := fs.Stat(path)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	if fi == fs.root {
		return os.ErrPermission
	}
	return fs.Wrapper().Remove(fi.UUID())
}

func (fs *FileSystem) Stat(name string) (*FileInfo, error) {
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}
	if name == "" || name == "/" {
		return fs.root, nil
	}
	fi := fs.root.FindByPath(name)
	if fi == nil {
		return nil, fmt.Errorf("%s: %w", name, os.ErrNotExist)
	}
	return fi, nil
}

func (fs *FileSystem) SaveFileTree() error {
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	data, err := json.Marshal(fs.root)
	if err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	if err := fs.EncryptPage(data); err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := fs.storage.Put(keyFSRootDir, data); err != nil {
		return fmt.Errorf("put: %w", err)
	}
	return nil
}

func (fs *FileSystem) EncryptPage(data []byte) error {
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	if fs.IsEncrypted() {
		return conv.AES(data, fs.key, fs.key[:16])
	}
	return nil
}

func (fs *FileSystem) DecryptPage(data []byte) error {
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	if fs.IsEncrypted() {
		return conv.AES(data, fs.key, fs.key[:16])
	}
	return nil
}

func (fs *FileSystem) Write(name string, data []byte) (*FileInfo, error) {
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}

	f, err := fs.OpenFile(name, WriteOnly|Create)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Write(data)
	return f.info, err
}

func (fs *FileSystem) Read(name string) ([]byte, error) {
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}

	f, err := fs.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	return fs.Wrapper().Read(f.UUID())
}

func (fs *FileSystem) WriteJSON(name string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = fs.Write(name, data)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

func (fs *FileSystem) ReadJSON(name string, v interface{}) error {
	data, err := fs.Read(name)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	err = json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}

func (fs *FileSystem) SetPermission(name string, permission int) error {
	f, err := fs.Stat(name)
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	f.Permission = permission
	err = fs.SaveFileTree()
	if err != nil {
		return fmt.Errorf("save file tree: %w", err)
	}
	return nil
}

func (fs *FileSystem) loadConfig() error {
	data, err := fs.storage.Get(keyFSConfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("get %s: %w", keyFSConfig, err)
	}

	if err = fs.DecryptPage(data); err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}

	err = json.Unmarshal(data, &fs.configs)
	if err != nil {
		return fmt.Errorf("unmarshal %s: %w", data, err)
	}
	return nil
}

func (fs *FileSystem) Config() types.M {
	return fs.configs
}

func (fs *FileSystem) SaveConfig() error {
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	data, err := json.Marshal(fs.configs)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = fs.EncryptPage(data); err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err = fs.storage.Put(keyFSConfig, data); err != nil {
		return fmt.Errorf("flag: %w", err)
	}
	return nil
}

func (fs *FileSystem) ListByPermission(p int) []*FileInfo {
	return fs.root.ListByPermission(p)
}

func (fs *FileSystem) Wrapper() *FileSystemWrapper {
	return (*FileSystemWrapper)(fs)
}

func (fs *FileSystem) Root() *FileInfo {
	return fs.root
}

func loadPageSize(storage KVStorage) (size int64, err error) {
	data, err := storage.Get(keyFSPageSize)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DefaultPageSize, nil
		}
		return size, fmt.Errorf("cannot read page size: %w", err)
	}
	err = json.Unmarshal(data, &size)
	if err != nil {
		return size, fmt.Errorf("cannot unmarshal page size: %w", err)
	}
	return size, nil
}

func savePageSize(storage KVStorage, size int64) error {
	data, err := json.Marshal(size)
	if err != nil {
		return fmt.Errorf("cannot marshal page size %d: %v", size, err)
	}
	err = storage.Put(keyFSPageSize, data)
	if err != nil {
		return fmt.Errorf("cannot save page size %d: %v", size, err)
	}
	return nil
}
