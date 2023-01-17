package badger

import (
	"encoding/binary"
)

// ------------------------------------------------------------------------

type visitStorage struct {
	s *stgBase
}

// ------------------------------------------------------------------------

// NewVisitMemoryStorage returns a pointer to a newly created BadgerDB visit storage.
func NewVisitStorage(path string, keepData bool) (*visitStorage, error) {
	cfg := config{
		prefix:      []byte{byte(TYPE_VISIT), 0},
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &visitStorage{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the BadgerDB visit storage.
func (s *visitStorage) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the BadgerDB visit storage.
func (s *visitStorage) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of request visits in the BadgerDB visit storage.
func (s *visitStorage) Len() (uint, error) {
	return s.s.Len()

}

// ------------------------------------------------------------------------

// AddVisit stores a request ID that is visited by the Collector.
func (s *visitStorage) AddVisit(key string) error {
	visits := uintToBytes(0)

	if b, err := s.s.Get([]byte(key)); err == nil || b != nil {
		visits = uintToBytes(bytesToUint(b) + 1)
	}

	return s.s.Set([]byte(key), visits)
}

// ------------------------------------------------------------------------

// PastVisits returns true if the request was visited before.
func (s *visitStorage) PastVisits(key string) (uint, error) {
	visits := uint(0)

	b, err := s.s.Get([]byte(key))
	if err == nil || b != nil {
		visits = bytesToUint(b)
	}

	return visits, err
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
