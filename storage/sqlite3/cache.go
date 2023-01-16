package sqlite3

import (
	"bytes"
	"database/sql"
	"io"
)

// ------------------------------------------------------------------------

type stgCache struct {
	s *stgBase
}

// ------------------------------------------------------------------------

const defaultCacheName = "cache"

// ------------------------------------------------------------------------

var (
	cmdCache = map[string]string{
		"create": `CREATE TABLE IF NOT EXISTS "<table>" ("key" TEXT PRIMARY KEY, "response" BLOB) WITHOUT ROWID`,
		"drop":   `DROP TABLE IF EXISTS "<table>"`,
		"trim":   `DELETE FROM "<table>"`,
		"insert": `INSERT INTO "<table>" ("key", "response") VALUES (?, ?) ON CONFLICT("key") DO UPDATE SET "response" = "excluded"."response"`,
		"select": `SELECT "caches" FROM "<table>" WHERE "key" = ?`,
		"delete": `DELETE FROM "<table>" WHERE "key" = ?`,
		"count":  `SELECT COUNT(*) FROM "<table>"`,
		"check":  `SELECT COUNT(*) FROM "<table>" WHERE "key" = ?`,
	}
)

// ------------------------------------------------------------------------

// NewCacheStorage returns a pointer to a newly created SQLite3 cache storage.
func NewCacheStorage(path string, table string, keepData bool) (*stgCache, error) {
	cfg := config{
		table:       setTable(table, defaultCacheName),
		dropOnClose: false,
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg, cmdCache)
	if err != nil {
		return nil, err
	}

	return &stgCache{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the SQLite3 cache storage.
func (s *stgCache) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the SQLite3 cache storage.
func (s *stgCache) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of hosts in the SQLite3 cache storage.
func (s *stgCache) Len() (uint, error) {
	return s.s.Len()
}

// ------------------------------------------------------------------------

// Put stores an item in the cache storage.
func (s *stgCache) Put(key string, item io.Reader) error {
	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}

	s.s.lock.Lock()
	_, err = s.s.stmts["insert"].Exec(key, data)
	s.s.lock.Unlock()

	return err
}

// ------------------------------------------------------------------------

// Fetch retrieves a cached item from the storage.
func (s *stgCache) Fetch(key string) (io.Reader, error) {
	var data = []byte{}

	s.s.lock.Lock()
	err := s.s.stmts["select"].QueryRow(key).Scan(&data)
	s.s.lock.Unlock()
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}

		return nil, err
	}

	return bytes.NewReader(data), nil
}

// ------------------------------------------------------------------------

// Has returns true if the key exists in the storage.
func (s *stgCache) Has(key string) bool {
	var count int

	s.s.lock.Lock()
	err := s.s.stmts["check"].QueryRow(key).Scan(&count)
	s.s.lock.Unlock()

	return err == nil && count != 0
}

// ------------------------------------------------------------------------

// Remove deletes a stored item by key.
func (s *stgCache) Remove(key string) error {
	s.s.lock.Lock()
	_, err := s.s.stmts["delete"].Exec(key)
	s.s.lock.Unlock()

	return err
}
