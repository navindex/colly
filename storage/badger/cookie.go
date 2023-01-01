package badger

import (
	"net/http"
	"net/url"

	"github.com/gocolly/colly/v2/storage"
)

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

// SetCookies stores cookies for a given host.
func (s *stgCookie) SetCookies(u *url.URL, cookies []*http.Cookie) error {
	data, err := storage.CookiesToBytes(cookies)
	if err != nil {
		return err
	}

	return s.s.Set([]byte(u.Host), data)
}

// ------------------------------------------------------------------------

// Cookies retrieves stored cookies for a given host.
func (s *stgCookie) Cookies(u *url.URL) ([]*http.Cookie, error) {
	data, err := s.s.Get([]byte(u.Host))
	if err != nil {
		return nil, err
	}

	return storage.BytesToCookies(data)
}
