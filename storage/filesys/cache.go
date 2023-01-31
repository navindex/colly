package filesys

import (
	"bytes"
	"colly/storage"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// ------------------------------------------------------------------------

// stgCache is a filesystem cache storage
type stgCache struct {
	lock     *sync.RWMutex
	path     string
	filePerm fs.FileMode
	dirPerm  fs.FileMode
	closed   bool
}

// ------------------------------------------------------------------------

const (
	DIR_PERM  fs.FileMode = 0750
	FILE_PERM fs.FileMode = 0644
)

// ------------------------------------------------------------------------

// NewCacheStorage returns a pointer to a newly created filesystem cache storage.
// After the path, the first optional argument is the directory permission,
// the second is the file permission.
func NewCacheStorage(path string, dirAndFilePermissions ...fs.FileMode) (*stgCache, error) {
	if path == "" {
		return nil, storage.ErrBlankPath
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	dirPerm := DIR_PERM
	filePerm := FILE_PERM
	permLen := len(dirAndFilePermissions)

	if permLen >= 1 {
		dirPerm = dirAndFilePermissions[0]
		if permLen >= 2 {
			filePerm = dirAndFilePermissions[1]
		}
	}

	if err := os.MkdirAll(path, dirPerm); err != nil {
		return nil, err
	}

	return &stgCache{
		lock:     &sync.RWMutex{},
		path:     path,
		dirPerm:  dirPerm,
		filePerm: filePerm,
		closed:   false,
	}, nil

}

// ------------------------------------------------------------------------

// Close closes the filesystem cache storage.
func (s *stgCache) Close() error {
	if s.closed {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	s.closed = true
	s.lock.Unlock()

	return nil
}

// ------------------------------------------------------------------------

// Clear removes all entries from the filesystem cache storage.
func (s *stgCache) Clear() error {
	if s.closed {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if err := os.RemoveAll(s.path); err != nil {
		return err
	}

	return os.MkdirAll(s.path, s.dirPerm)
}

// ------------------------------------------------------------------------

// Len returns the number of items in the filesystem cache storage.
func (s *stgCache) Len() (uint, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return FileCount(s.path)
}

// ------------------------------------------------------------------------

// Put stores an item in the cache storage.
func (s *stgCache) Put(key string, item io.Reader) error {
	if s.closed {
		return storage.ErrStorageClosed
	}

	if len(key) < 4 {
		return storage.ErrInvalidKey
	}

	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}

	key = SanitizeFileName(key)
	dir := filepath.Join(s.path, key[:2])

	s.lock.RLock()
	defer s.lock.RUnlock()

	if err := os.MkdirAll(dir, s.dirPerm); err != nil {
		return err
	}

	path := filepath.Join(dir, key)
	file, err := os.Create(path + "~")
	if err != nil {
		return err
	}
	defer file.Close()

	if file.Chmod(s.filePerm); err != nil {
		return err
	}

	if _, err := file.Write(data); err != nil {
		return err
	}

	return os.Rename(path+"~", path)
}

// ------------------------------------------------------------------------

// Fetch retrieves a cached item from the storage.
func (s *stgCache) Fetch(key string) (io.Reader, error) {
	if s.closed {
		return nil, storage.ErrStorageClosed
	}

	if len(key) < 4 {
		return nil, storage.ErrInvalidKey
	}

	key = SanitizeFileName(key)
	path := filepath.Join(s.path, key[:2], key)

	s.lock.RLock()
	data, err := os.ReadFile(path)
	s.lock.RUnlock()

	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}

		return nil, err
	}

	return bytes.NewReader(data), nil
}

// ------------------------------------------------------------------------

// Has returns true if the key exists in the storage.
func (s *stgCache) Has(key string) bool {
	if len(key) < 4 {
		return false
	}

	path := filepath.Join(s.path, key[:2], key)

	s.lock.RLock()
	info, err := os.Stat(path)
	s.lock.RUnlock()

	return err == nil && !info.IsDir()
}

// ------------------------------------------------------------------------

// Remove deletes a stored item by key.
func (s *stgCache) Remove(key string) error {
	if len(key) < 4 {
		return storage.ErrInvalidKey
	}

	path := filepath.Join(s.path, key[:2], key)

	s.lock.RLock()
	defer s.lock.RUnlock()

	return os.Remove(path)
}
