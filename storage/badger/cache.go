package badger

import (
	"bytes"
	"io"
)

// ------------------------------------------------------------------------

type stgCache struct {
	s *stgBase
}

// ------------------------------------------------------------------------

// NewCacheStorage returns a pointer to a newly created BadgerDB cache storage.
func NewCacheStorage(path string, keepData bool) (*stgCache, error) {
	cfg := config{
		prefix:      []byte{byte(TYPE_CACHE), 0},
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &stgCache{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the BadgerDB cache storage.
func (s *stgCache) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all items from the BadgerDB cache storage.
func (s *stgCache) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of items in the BadgerDB cache storage.
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

	return s.s.Set([]byte(key), data)
}

// ------------------------------------------------------------------------

// Fetch retrieves a cached item from the storage.
func (s *stgCache) Fetch(key string) (io.Reader, error) {
	data, err := s.s.Get([]byte(key))
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}

// ------------------------------------------------------------------------

// Has returns true if the key exists in the storage.
func (s *stgCache) Has(key string) bool {
	data, err := s.s.Get([]byte(key))
	if err == nil && data != nil {
		return true
	}

	return false
}

// ------------------------------------------------------------------------

// Remove deletes a stored item by key.
func (s *stgCache) Remove(key string) error {
	return s.s.DropPrefix([]byte(key))
}
