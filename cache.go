package colly

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// ------------------------------------------------------------------------

// Cache is a collection of functions to managed cached HTTP reponses.
type Cache interface {
	Set(*Response) error               // Set writes a response to the cache.
	Get(url string) (*Response, error) // Get retrieves a cached response.
	Remove(url string) error           // Remove removes a cache item by key.
	RemoveAll() error                  // RemoveAll removes all cache items.
}

// CacheExpiryHandler identifies whether or not a cache item expired.
type CacheExpiryHandler interface {
	Expired(created time.Time, expiry time.Time) bool // Expired returns true if the response is expired.
}

// CacheStorage is a storage to manage cached HTTP responses.
type CacheStorage interface {
	Put(key string, data io.Reader) error         // Put stores a response with a timestamp.
	Fetch(key string) (data io.Reader, err error) // Fetch retrieves a response from the storage.
	Has(key string) bool                          // Has returns true if the key exists in the storage.
	Remove(key string) error                      // Remove deletes stored items by keys.
	Clear() error                                 // Clear deletes all stored items.
}

type cache struct {
	stg CacheStorage       // Data storage
	exp CacheExpiryHandler // Item expiry handler
}

// cacheExpByHeader checks the expiry by the page header
type cacheExpByHeader struct{}

// cacheExpByDuration checks the expiry by the time passed after caching the item
type cacheExpByDuration struct {
	duration time.Duration
}

// cacheExpByDate checks the expiry by comparing the cached timestamp to an expiry timestamp
type cacheExpByDate struct {
	expiry time.Time
}

// cacheExpNever checks the expiry by the page header
type cacheExpNever struct{}

// ------------------------------------------------------------------------

// NewCache returns a pointer to a newly created cache object.
func NewCache(cs CacheStorage, exp CacheExpiryHandler) (*cache, error) {
	if cs == nil {
		return nil, ErrCacheNoStorage
	}

	if exp == nil {
		return nil, ErrCacheSNoExpHandler
	}

	c := &cache{
		stg: cs,
		exp: exp,
	}

	return c, nil
}

// ------------------------------------------------------------------------

// Set writes a response to the cache.
func (c *cache) Set(resp *Response) error {
	url := resp.Request.Req.URL.String()
	key := c.keyFromURL(url)

	data, err := c.encodeResponse(resp)
	if err != nil {
		return err
	}

	return c.stg.Put(key, data)
}

// ------------------------------------------------------------------------

// Get retrieves a cached response.
func (c *cache) Get(url string) (*Response, error) {
	key := c.keyFromURL(url)

	data, err := c.stg.Fetch(key)
	if err != nil {
		return nil, err
	}

	resp, err := c.decodeData(data)
	if err != nil {
		return nil, err
	}

	if c.exp.Expired(resp.Created, resp.Expiry) {
		return nil, nil
	}

	return resp, nil
}

// ------------------------------------------------------------------------

// Remove removes a cache item by key.
func (c *cache) Remove(url string) error {
	return c.stg.Remove(c.keyFromURL(url))
}

// ------------------------------------------------------------------------

// RemoveAll removes all cache items.
func (c *cache) RemoveAll() error {
	return c.stg.Clear()
}

// ------------------------------------------------------------------------

func (c *cache) keyFromURL(url string) string {
	sum := sha1.Sum([]byte(url))
	return hex.EncodeToString(sum[:])
}

func (c *cache) encodeResponse(resp *Response) (io.Reader, error) {
	data := &bytes.Buffer{}
	err := gob.NewEncoder(data).Encode(resp)

	return data, err
}

func (c *cache) decodeData(data io.Reader) (*Response, error) {
	resp := &Response{}
	err := gob.NewDecoder(data).Decode(resp)

	return resp, err
}

// ------------------------------------------------------------------------

// NewCacheExpiryByHeader returns a pointer to a newly created expiration controller
// that is based on the response's cache expiry header.
func NewCacheExpiryByHeader() *cacheExpByHeader {
	return &cacheExpByHeader{}
}

// Expired implements the CacheExpiryHandler interface.
func (h *cacheExpByHeader) Expired(_ time.Time, expiry time.Time) bool {
	return time.Now().After(expiry)
}

// ------------------------------------------------------------------------

// NewCacheExpiryByDuration returns a pointer to a newly created expiration controller
// that is based on how long ago the item was cached.
func NewCacheExpiryByDuration(duration time.Duration) (*cacheExpByDuration, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("duration must be positive: %s was given", duration)
	}

	return &cacheExpByDuration{
		duration: duration,
	}, nil
}

// Expired implements the CacheExpiryHandler interface.
func (h *cacheExpByDuration) Expired(cachedAt time.Time, _ time.Time) bool {
	return time.Now().After(cachedAt.Add(h.duration))
}

// ------------------------------------------------------------------------

// NewCacheExpiryByDate returns a pointer to a newly created expiration controller that based on a fix date.
func NewCacheExpiryByDate(expiry time.Time) (*cacheExpByDate, error) {
	if expiry.IsZero() || expiry.Before(time.Now()) {
		return nil, fmt.Errorf("expiry date must be in the future: %s was given", expiry)
	}

	return &cacheExpByDate{
		expiry: expiry,
	}, nil
}

// Expired implements the CacheExpiryHandler interface.
func (h *cacheExpByDate) Expired(_ time.Time, _ time.Time) bool {
	return time.Now().After(h.expiry)
}

// ------------------------------------------------------------------------

// NewCacheExpiryByDate returns a pointer to a newly created expiration controller that never expires.
func NewCacheExpiryNever() *cacheExpNever {
	return &cacheExpNever{}
}

// Expired implements the CacheExpiryHandler interface.
func (h *cacheExpNever) Expired(_ time.Time, _ time.Time) bool {
	return false
}
