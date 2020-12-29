package vfs

import (
	"fmt"
	"os"
	"sync"

	"github.com/gopub/sql/sqlite"

	"github.com/gopub/sql"
)

type SQLiteStorage struct {
	db *sql.DB
	mu sync.RWMutex
}

var _ Storage = (*SQLiteStorage)(nil)

func NewSQLiteStorage(name string) (*SQLiteStorage, error) {
	db := sqlite.Open(name)
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS vfs(
k VARCHAR(255) PRIMARY KEY, 
v BLOB NOT NULL
)`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}
	return &SQLiteStorage{
		db: db,
	}, nil
}

func (s *SQLiteStorage) Put(key string, data []byte) error {
	s.mu.Lock()
	_, err := s.db.Exec("REPLACE INTO vfs(k,v) VALUES(?1,?2)", key, data)
	s.mu.Unlock()
	return err
}

func (s *SQLiteStorage) Get(key string) ([]byte, error) {
	var v []byte
	s.mu.RLock()
	err := s.db.QueryRow("SELECT v FROM vfs WHERE k=?", key).Scan(&v)
	s.mu.RUnlock()

	if err == sql.ErrNoRows {
		return nil, os.ErrNotExist
	}
	return v, err
}

func (s *SQLiteStorage) Delete(key string) error {
	s.mu.RLock()
	_, err := s.db.Exec("DELETE FROM vfs WHERE k=?", key)
	s.mu.RUnlock()
	return err
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
