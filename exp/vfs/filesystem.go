package vfs

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/types"
	"net/http"
	"os"
	"path/filepath"
)

type FileSystem struct {
	storage  KVStorage
	root     *FileInfo
	key      []byte
	configs  types.M
	pageSize int64
}

var _ http.FileSystem = (*FileSystem)(nil)

func NewFileSystem(storage KVStorage, pageSize int64, password string) (*FileSystem, error) {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize < MinPageSize {
		return nil, fmt.Errorf("min page size is %d", MinPageSize)
	}
	empty, err := isEmptyStorage(storage)
	if err != nil {
		return nil, fmt.Errorf("detect empty storage: %w", err)
	}

	fs := &FileSystem{
		storage:  storage,
		pageSize: pageSize,
		configs:  types.M{},
	}

	if empty {
		if err = savePageSize(storage, pageSize); err != nil {
			return nil, fmt.Errorf("save page size: %w", err)
		}
		if password != "" {
			fs.key, err = setupCredential(storage, password)
			if err != nil {
				return nil, fmt.Errorf("setup credential: %w", err)
			}
		}
	} else {
		expectedSize, err := loadPageSize(storage)
		if err != nil {
			return nil, fmt.Errorf("load page size: %w", err)
		}
		if expectedSize != pageSize {
			return nil, fmt.Errorf("mismatch page size %d != %d", pageSize, expectedSize)
		}
		encrypted, err := isEncryptedStorage(storage)
		if err != nil {
			return nil, fmt.Errorf("detech encryption: %w", err)
		}
		if encrypted {
			fs.key, err = loadCredential(storage, password)
			if err != nil {
				return nil, fmt.Errorf("load credential: %w", err)
			}
		} else {
			if password != "" {
				return nil, fmt.Errorf("password is not required for unencrypted storage")
			}
		}
	}
	if err = fs.loadConfig(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if err = fs.mount(storage); err != nil {
		return nil, fmt.Errorf("mount root: %w", err)
	}
	return fs, nil
}

func (fs *FileSystem) mount(storage KVStorage) error {
	if fs.root != nil {
		log.Panic("Mounted")
	}
	data, err := storage.Get(keyFSRootDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("get %s: %w", keyFSRootDir, err)
		}

		// initialize
		fs.root = newFileInfo(true, "")
		if err = fs.SaveFileTree(); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		return nil
	}

	if err = fs.DecryptPage(data); err != nil {
		return fmt.Errorf("decrypt: %w", err)
	}
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&fs.root); err != nil {
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
			log.Error(name, base, paths)
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
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)

	if err := enc.Encode(fs.root); err != nil {
		return fmt.Errorf("encode: %w", err)
	}

	data := buf.Bytes()
	if err := fs.EncryptPage(data); err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	if err := fs.storage.Put(keyFSRootDir, data); err != nil {
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

func (fs *FileSystem) Write(name string, data []byte) (*FileInfo, error) {
	f, err := fs.OpenFile(name, WriteOnly|Create)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	_, err = f.Write(data)
	return f.info, err
}

func (fs *FileSystem) Read(name string) ([]byte, error) {
	f, err := fs.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}
	return fs.Wrapper().Read(f.UUID())
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

func (fs *FileSystem) VerifyPassword(password string) bool {
	_, err := loadCredential(fs.storage, password)
	return err == nil
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

func isEmptyStorage(storage KVStorage) (bool, error) {
	_, err := storage.Get(keyFSRootDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}
		return true, fmt.Errorf("read: %w", err)
	}
	return false, nil
}

func isEncryptedStorage(storage KVStorage) (bool, error) {
	credential, err := storage.Get(keyFSCredential)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return len(credential) > 0, nil
}

func setupCredential(storage KVStorage, password string) ([]byte, error) {
	if password == "" {
		log.Panic("Missing password")
	}

	credential, err := storage.Get(keyFSCredential)
	if err == nil {
		return nil, errors.New("credential exists")
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read credential: %w", err)
	}

	if password == "" {
		return nil, errors.New("missing password")
	}

	// this is a new file system, initialize key if password is provided
	passwordHash := conv.Hash32([]byte(password))
	key := conv.Hash32([]byte(uuid.New().String()))

	credential = make([]byte, 2*keySize)
	copy(credential, key[:])
	copy(credential[keySize:], passwordHash[:])

	// encrypt
	if err = conv.AES(credential, passwordHash[:], passwordHash[:16]); err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	if err = storage.Put(keyFSCredential, credential); err != nil {
		return nil, fmt.Errorf("flag credential: %w", err)
	}
	return key[:], nil
}

func loadCredential(storage KVStorage, password string) ([]byte, error) {
	if password == "" {
		return nil, ErrAuth
	}
	credential, err := storage.Get(keyFSCredential)
	if err != nil {
		return nil, fmt.Errorf("read credential: %w", err)
	}

	passwordHash := conv.Hash32([]byte(password))
	if err = conv.AES(credential, passwordHash[:], passwordHash[:16]); err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	if !bytes.Equal(credential[keySize:], passwordHash[:]) {
		return nil, ErrAuth
	}
	return credential[:keySize], nil
}
