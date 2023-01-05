package cache

import (
	"net/url"
	"time"
)

// ------------------------------------------------------------------------

type ExpirationController interface {
	Expired() bool
}

// Storage is a Storage to manage cached responses.
type Storage interface {
	Put(*url.URL, []byte, time.Time) error // Put adds the response to the cache.
	Fetch(*url.URL) ([]byte, error)        // Fetch retrieves the response from the cache.
	Has(*url.URL, bool) bool               // Has returns true if the URL exists in the cache.
	Remove(*url.URL) error                 // Remove deletes the response from the cache.
}

type cache struct {
	stg  Storage              // Data storage
	ctrl ExpirationController // Expiration controller
}

// ------------------------------------------------------------------------

func NewCache(s Storage, expCtrl ExpirationController) (*cache, error) {
	if s == nil {
		return nil, nil
	}
	c := &cache{
		stg:  s,
		ctrl: expCtrl,
	}

	return c, nil
}

// ------------------------------------------------------------------------

func (c *cache) Save(key string, data []byte) error {
	return nil
}
