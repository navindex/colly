package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ------------------------------------------------------------------------

// Errors
var (
	ErrNotImplemented = errors.New("feature not implemented")
	ErrStorageEmpty   = errors.New("storage is empty")
	ErrStorageFull    = errors.New("storage is full")
	ErrStorageClosed  = errors.New("storage is closed")
	ErrBlankPath      = errors.New("no storage path was given")
	ErrBlankKey       = errors.New("no key was given")
	ErrBlankTableName = errors.New("no table name was given")
	ErrInvalidType    = errors.New("invalid storage type")
	ErrStorageLimit   = errors.New("unable to connect to the database, storage limit exceeded")
	ErrInvalidConn    = errors.New("invalid database connection")
	ErrMissingParams  = errors.New("storage parameters are missing")
	ErrNInvalidLength = errors.New("max queue length must be positive or zero for no limit")
	ErrMissingCmd     = func(cmd string) error { return fmt.Errorf("%s command is missing", cmd) }
)

// ------------------------------------------------------------------------

// CookiesToBytes encodes an array of cookies to bytes.
func CookiesToBytes(cookies []*http.Cookie) ([]byte, error) {
	// Extract the cookies
	c := []http.Cookie{}
	for i := range cookies {
		c = append(c, *cookies[i])
	}

	// Encode
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(c)

	return b.Bytes(), err
}

// ------------------------------------------------------------------------

// BytesToCookies retrieves a previously encoded array of cookies from bytes.
func BytesToCookies(data []byte) ([]*http.Cookie, error) {
	// Convert byte slice to io.Reader
	reader := bytes.NewReader(data)

	// Decode to a slice of cookies
	c := []http.Cookie{}
	err := gob.NewDecoder(reader).Decode(&c)
	if err != nil {
		return nil, err
	}

	// Create a slice of pointers
	cookies := []*http.Cookie{}
	for i := range c {
		cookies = append(cookies, &c[i])
	}

	return cookies, nil
}

// ------------------------------------------------------------------------

// Uint64ToBytes converts uint64 to bytes.
func Uint64ToBytes(i uint64) []byte {
	b := []byte{}
	binary.BigEndian.PutUint64(b, i)

	return b
}

// ------------------------------------------------------------------------

// BytesToUint64 converts bytes to uint64.
func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// ------------------------------------------------------------------------

// CurrentTimeToBytes converts the current timestamp to bytes.
func CurrentTimeToBytes() []byte {
	t := time.Now().Unix()
	if t < 0 {
		return nil
	}

	return Uint64ToBytes(uint64(t))
}
