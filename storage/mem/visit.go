package mem

import (
	"colly/storage"
	"sync"
)

// ------------------------------------------------------------------------

// In-memory visit storage
type stgVisit struct {
	lock   *sync.RWMutex
	visits map[uint64]bool
}

// ------------------------------------------------------------------------

// NewVisitStorage returns a pointer to a newly created in-memory visit storage.
func NewVisitStorage() *stgVisit {
	return &stgVisit{
		lock:   &sync.RWMutex{},
		visits: map[uint64]bool{},
	}
}

// ------------------------------------------------------------------------

// Close closes the in-memory visit storage.
func (s *stgVisit) Close() error {
	if s.visits == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.visits = nil

	return nil
}

// ------------------------------------------------------------------------

// Clear removes all entries from the in-memory visit storage.
func (s *stgVisit) Clear() error {
	if s.visits == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.visits = map[uint64]bool{}

	return nil
}

// ------------------------------------------------------------------------

// Len returns the number of request visits in the in-memory visit storage.
func (s *stgVisit) Len() (uint, error) {
	if s.visits == nil {
		return 0, storage.ErrStorageClosed
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return uint(len(s.visits)), nil
}

// ------------------------------------------------------------------------

// SetVisited receives and stores a request ID that is visited by the Collector.
func (s *stgVisit) SetVisited(requestID uint64) error {
	if s.visits == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.visits[requestID] = true

	return nil
}

// ------------------------------------------------------------------------

// IsVisited returns true if the request was visited before.
func (s *stgVisit) IsVisited(requestID uint64) (bool, error) {
	if s.visits == nil {
		return false, storage.ErrStorageClosed
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.visits[requestID], nil
}
