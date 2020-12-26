package vfs

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gopub/conv"
	"github.com/gopub/errors"
	"github.com/gopub/log"
	"github.com/gopub/types"
)

type FileSystem struct {
	storage KVStorage
	root    *FileInfo
	key     []byte
	configs types.M
}

var _ http.FileSystem = (*FileSystem)(nil)

func NewFileSystem(storage KVStorage, password string) (*FileSystem, error) {
	empty, err := isEmptyStorage(storage)
	if err != nil {
		return nil, fmt.Errorf("detect empty storage: %w", err)
	}

	fs := &FileSystem{
		storage: storage,
	}

	if empty {
		if password != "" {
			fs.key, err = setupCredential(storage, password)
			if err != nil {
				return nil, fmt.Errorf("setup credential: %w", err)
			}
		}
	} else {
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

func (fs *FileSystem) Mkdir(name string) (*FileInfo, error) {
	if name == "" {
		return nil, os.ErrInvalid
	}
	f := fs.root.GetByPath(name)
	if f != nil {
		if f.IsDir() {
			return f, nil
		}
		return nil, fmt.Errorf("%s is not directory", name)
	}
	paths := splitPath(name)
	if len(paths) == 1 {
		f = newFileInfo(true, fs.root.DistinctName(name))
		fs.root.AddSub(f)
		return f, fs.SaveFileTree()
	}

	dirPaths := paths[:len(paths)-1]
	dir := fs.root.getByPathList(dirPaths)
	if dir == nil {
		return nil, fmt.Errorf("dir %s does not exist", filepath.Join(dirPaths...))
	}
	if !dir.IsDir() {
		return nil, fmt.Errorf("%s is not directory", filepath.Join(dirPaths...))
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
	f := dir.GetByName(base)
	if f == nil {
		if (flag & Create) == 0 {
			return nil, os.ErrNotExist
		}
		f = newFileInfo(false, dir.DistinctName(base))
	}
	dir.AddSub(f)
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
	fi := fs.root.GetByPath(name)
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

func (fs *FileSystem) loadConfig() error {
	data, err := fs.storage.Get(keyFSConfig)
	if err != nil {
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
