package badger

// ------------------------------------------------------------------------

type stgCookie struct {
	s *stgBase
}

// ------------------------------------------------------------------------

var prefixCookie = []byte{2, 0}

// ------------------------------------------------------------------------

// NewCookieStorage returns a pointer to a newly created BadgerDB cookie storage.
func NewCookieStorage(path string, keepData bool) (*stgCookie, error) {
	cfg := config{
		prefix:      prefixCookie,
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &stgCookie{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the BadgerDB cookie storage.
func (s *stgCookie) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the BadgerDB cookie storage.
func (s *stgCookie) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of request cookies in the BadgerDB cookie storage.
func (s *stgCookie) Len() (uint, error) {
	return s.s.Len()
}

// ------------------------------------------------------------------------

// Set stores cookies for a given host.
func (s *stgCookie) Set(key string, cookies []byte) error {
	return s.s.Set([]byte(key), cookies)
}

// ------------------------------------------------------------------------

// Get retrieves stored cookies for a given host.
func (s *stgCookie) Get(key string) ([]byte, error) {
	return s.s.Get([]byte(key))
}

// ------------------------------------------------------------------------

// Remove deletes stored cookies for a given host.
func (s *stgCookie) Remove(key string) error {
	return s.s.DropPrefix([]byte(key))
}
