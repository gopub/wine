package vfs

import (
	"os"
)

type MemoryStorage struct {
	m map[string][]byte
}

func NewMemoryStorage() *MemoryStorage {
	s := &MemoryStorage{
		m: make(map[string][]byte),
	}
	return s
}

func (s *MemoryStorage) Get(key string) ([]byte, error) {
	v, ok := s.m[key]
	if ok {
		b := make([]byte, len(v))
		copy(b, v)
		return b, nil
	}
	return nil, os.ErrNotExist
}

func (s *MemoryStorage) Put(key string, v []byte) error {
	b := make([]byte, len(v))
	copy(b, v)
	s.m[key] = b
	return nil
}

func (s *MemoryStorage) Delete(key string) error {
	delete(s.m, key)
	return nil
}
