package mem

import (
	"colly/storage"
	"sync"
)

// ------------------------------------------------------------------------

// In-memory visit storage
type stgVisit struct {
	lock   *sync.RWMutex
	visits map[string]uint
}

// ------------------------------------------------------------------------

// NewVisitStorage returns a pointer to a newly created in-memory visit storage.
func NewVisitStorage() *stgVisit {
	return &stgVisit{
		lock:   &sync.RWMutex{},
		visits: map[string]uint{},
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

	s.visits = map[string]uint{}

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

// AddVisit receives and stores a request ID that is visited by the Collector.
func (s *stgVisit) AddVisit(key string) error {
	if s.visits == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	if visits, present := s.visits[key]; present {
		s.visits[key] = visits + 1
	} else {
		s.visits[key] = uint(1)
	}
	s.lock.Unlock()

	return nil
}

// ------------------------------------------------------------------------

// PastVisits returns true if the request was visited before.
func (s *stgVisit) PastVisits(key string) (uint, error) {
	if s.visits == nil {
		return 0, storage.ErrStorageClosed
	}

	visits := uint(0)

	s.lock.RLock()
	if v, present := s.visits[key]; !present {
		visits = v
	}
	s.lock.RUnlock()

	return visits, nil
}
