package vfs

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/gopub/errors"
)

type SecKeyItem struct {
	UUID       string `json:"uuid"`
	Name       string `json:"name"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	ModifiedAt int64  `json:"modified_at"`
}

type fileSystemKeyChain FileSystem

func (fs *FileSystem) KeyChain() *fileSystemKeyChain {
	return (*fileSystemKeyChain)(fs)
}

func (kc *fileSystemKeyChain) load() (map[string]*SecKeyItem, error) {
	fs := (*FileSystem)(kc)
	if !fs.AuthPassed() {
		return nil, os.ErrPermission
	}
	data, err := fs.storage.Get(keyFSKeyChain)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]*SecKeyItem{}, nil
		}
		return nil, fmt.Errorf("read data: %w", err)
	}

	err = fs.DecryptPage(data)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	var res map[string]*SecKeyItem
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data: %w", err)
	}

	return res, nil
}

func (kc *fileSystemKeyChain) save(items map[string]*SecKeyItem) error {
	fs := (*FileSystem)(kc)
	if !fs.AuthPassed() {
		return os.ErrPermission
	}

	data, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	err = fs.EncryptPage(data)
	if err != nil {
		return fmt.Errorf("encrypt: %w", err)
	}

	err = fs.storage.Put(keyFSKeyChain, data)
	return errors.Wrapf(err, "write: %w", err)
}

func (kc *fileSystemKeyChain) Save(item *SecKeyItem) error {
	items, err := kc.load()
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	if item.UUID == "" {
		item.UUID = uuid.NewString()
	}
	item.ModifiedAt = time.Now().Unix()
	items[item.UUID] = item

	return kc.save(items)
}

func (kc *fileSystemKeyChain) Delete(uuid string) error {
	items, err := kc.load()
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	delete(items, uuid)
	return kc.save(items)
}

func (kc *fileSystemKeyChain) List() ([]*SecKeyItem, error) {
	items, err := kc.load()
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	a := make([]*SecKeyItem, 0, len(items))
	for _, v := range items {
		a = append(a, v)
	}
	sort.Slice(a, func(i, j int) bool {
		return a[i].Name < a[j].Name
	})
	return a, nil
}
