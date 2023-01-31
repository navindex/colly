package badger

import (
	"encoding/binary"
)

// ------------------------------------------------------------------------

type stgVisit struct {
	s *stgBase
}

// ------------------------------------------------------------------------

// NewVisitStorage returns a pointer to a newly created BadgerDB visit storage.
func NewVisitStorage(path string, keepData bool) (*stgVisit, error) {
	cfg := config{
		prefix:      []byte{byte(TYPE_VISIT), 0},
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &stgVisit{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the BadgerDB visit storage.
func (s *stgVisit) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the BadgerDB visit storage.
func (s *stgVisit) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of request visits in the BadgerDB visit storage.
func (s *stgVisit) Len() (uint, error) {
	return s.s.Len(nil)

}

// ------------------------------------------------------------------------

// AddVisit stores a request ID that is visited by the Collector.
func (s *stgVisit) AddVisit(key string) error {
	visits := uintToBytes(0)

	if b, err := s.s.Get([]byte(key)); err == nil || b != nil {
		visits = uintToBytes(bytesToUint(b) + 1)
	}

	return s.s.Set([]byte(key), visits)
}

// ------------------------------------------------------------------------

// PastVisits returns true if the request was visited before.
func (s *stgVisit) PastVisits(key string) (uint, error) {
	visits := uint(0)

	b, err := s.s.Get([]byte(key))
	if err == nil || b != nil {
		visits = bytesToUint(b)
	}

	return visits, err
}

// ------------------------------------------------------------------------

// Remove deletes a stored item by key.
func (s *stgVisit) Remove(key string) error {
	return s.s.DropPrefix([]byte(key))
}

// ------------------------------------------------------------------------

// uintToBytes converts uint to bytes
func uintToBytes(i uint) []byte {
	b := []byte{}
	binary.BigEndian.PutUint64(b, uint64(i))

	return b
}

// ------------------------------------------------------------------------

// bytesToUint converts bytes to uint
func bytesToUint(b []byte) uint {
	return uint(binary.BigEndian.Uint64(b))
}
