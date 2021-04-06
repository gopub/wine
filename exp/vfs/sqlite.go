package vfs

import (
	"fmt"
	"os"
	"sync"

	"github.com/gopub/log"
	"github.com/gopub/sql"
	"github.com/gopub/sql/sqlite"
)

type SQLiteStorage struct {
	db   *sql.DB
	mu   sync.RWMutex
	name string
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
		db:   db,
		name: name,
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

func (s *SQLiteStorage) MultiDelete(keys []string) error {
	for _, key := range keys {
		_, err := s.db.Exec(`DELETE FROM vfs WHERE k=?1`, key)
		if err != nil {
			return err
		}
	}
	_, err := s.db.Exec(`VACUUM`)
	if err != nil {
		log.Error(err)
	}
	return nil
}

func (s *SQLiteStorage) ListKeysByLength(keyLen int) ([]string, error) {
	rows, err := s.db.Query(`SELECT k FROM vfs WHERE LENGTH(k)=?`, keyLen)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) DB() *sql.DB {
	return s.db
}

func (s *SQLiteStorage) Size() (int64, error) {
	fi, err := os.Stat(s.name)
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}
