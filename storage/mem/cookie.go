package mem

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"github.com/gocolly/colly/v2/storage"
)

// ------------------------------------------------------------------------

// In-memory cookie storage
type stgCookie struct {
	lock *sync.Mutex
	jar  *cookiejar.Jar
}

// ------------------------------------------------------------------------

// NewCookieStorage returns a pointer to a newly created in-memory cookie storage.
func NewCookieStorage() (*stgCookie, error) {
	var err error

	s := &stgCookie{
		lock: &sync.Mutex{},
		jar:  nil,
	}

	if s.jar, err = cookiejar.New(nil); err != nil {
		return nil, err
	}

	return s, nil
}

// ------------------------------------------------------------------------

// Close closes the in-memory cookie storage.
func (s *stgCookie) Close() error {
	if s.jar == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.jar = nil

	return nil
}

// ------------------------------------------------------------------------

// Clear removes all entries from the in-memory cookie storage.
func (s *stgCookie) Clear() error {
	if s.jar == nil {
		return storage.ErrStorageClosed
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.jar = jar

	return nil
}

// ------------------------------------------------------------------------

// Len hasn't been implemented.
func (s *stgCookie) Len() (uint, error) {
	return 0, storage.ErrNotImplemented
}

// ------------------------------------------------------------------------

// SetCookies stores cookies for a given host.
func (s *stgCookie) SetCookies(u *url.URL, cookies []*http.Cookie) error {
	if s.jar == nil {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.jar.SetCookies(u, cookies)

	return nil
}

// ------------------------------------------------------------------------

// Cookies retrieves stored cookies for a given host.
func (s *stgCookie) Cookies(u *url.URL) ([]*http.Cookie, error) {
	if s.jar == nil {
		return nil, storage.ErrStorageClosed
	}

	return s.jar.Cookies(u), nil
}
