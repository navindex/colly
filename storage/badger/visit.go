package badger

import "colly/storage"

// ------------------------------------------------------------------------

type visitStorage struct {
	s *stgBase
}

// ------------------------------------------------------------------------

var prefixVisit = []byte{1, 0}

// ------------------------------------------------------------------------

// NewVisitMemoryStorage returns a pointer to a newly created BadgerDB visit storage.
func NewVisitStorage(path string, keepData bool) (*visitStorage, error) {
	cfg := config{
		prefix:      prefixVisit,
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

// SetVisited stores a request ID that is visited by the Collector.
func (s *visitStorage) SetVisited(requestID uint64) error {
	return s.s.SetBool(storage.Uint64ToBytes(requestID), true)
}

// ------------------------------------------------------------------------

// IsVisited returns true if the request was visited before.
func (s *visitStorage) IsVisited(requestID uint64) (bool, error) {
	return s.s.GetBool(storage.Uint64ToBytes(requestID))
}
