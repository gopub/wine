package vfs

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
		return v, nil
	}
	return nil, ErrNotExist
}

func (s *MemoryStorage) Put(key string, v []byte) error {
	s.m[key] = v
	return nil
}

func (s *MemoryStorage) Delete(key string) error {
	delete(s.m, key)
	return nil
}
