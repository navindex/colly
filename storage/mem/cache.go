package mem

import (
	"bytes"
	"colly/storage"
	"io"
	"sync"
)

// ------------------------------------------------------------------------

// In-memory cache storage
type stgCache struct {
	lock  *sync.RWMutex
	cache map[string][]byte
}

// ------------------------------------------------------------------------

// NewCacheStorage returns a pointer to a newly created in-memory cache storage.
func NewCacheStorage() *stgCache {
	return &stgCache{
		lock:  &sync.RWMutex{},
		cache: map[string][]byte{},
	}
}

// ------------------------------------------------------------------------

// Close closes the in-memory cache storage.
func (s *stgCache) Close() error {
	if s.cache == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	s.cache = nil
	s.lock.Unlock()

	return nil
}

// ------------------------------------------------------------------------

// Clear removes all entries from the in-memory cache storage.
func (s *stgCache) Clear() error {
	if s.cache == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	s.cache = map[string][]byte{}
	s.lock.Unlock()

	return nil
}

// ------------------------------------------------------------------------

// Len returns the number of items in the in-memory cache storage.
func (s *stgCache) Len() (uint, error) {
	if s.cache == nil {
		return 0, storage.ErrStorageClosed
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return uint(len(s.cache)), nil
}

// ------------------------------------------------------------------------

// Put stores an item in the cache storage.
func (s *stgCache) Put(key string, item io.Reader) error {
	if s.cache == nil {
		return storage.ErrStorageClosed
	}

	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}

	s.lock.Lock()
	s.cache[key] = data
	s.lock.Unlock()

	return nil
}

// ------------------------------------------------------------------------

// Fetch retrieves a cached item from the storage.
func (s *stgCache) Fetch(key string) (io.Reader, error) {
	if s.cache == nil {
		return nil, storage.ErrStorageClosed
	}

	s.lock.RLock()
	data, present := s.cache[key]
	s.lock.RUnlock()

	if !present {
		return nil, nil
	}

	return bytes.NewReader(data), nil
}

// ------------------------------------------------------------------------

// Has returns true if the key exists in the storage.
func (s *stgCache) Has(key string) bool {
	if s.cache == nil {
		return false
	}

	s.lock.RLock()
	_, present := s.cache[key]
	s.lock.RUnlock()

	return present
}

// ------------------------------------------------------------------------

// Remove deletes a stored item by key.
func (s *stgCache) Remove(key string) error {
	s.lock.Lock()
	delete(s.cache, key)
	s.lock.Unlock()

	return nil
}
