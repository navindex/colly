package storage

import (
	"errors"
	"fmt"
)

// ------------------------------------------------------------------------

// BaseStorage is an interface which handles Collector's internal data.
type BaseStorage interface {
	Clear() error // Clear removes all entries from the storage.
	Close() error // Close closes the storage ensures writes all pending updates.
}

// ------------------------------------------------------------------------

// Errors
var (
	ErrNotImplemented   = errors.New("feature not implemented")
	ErrStorageEmpty     = errors.New("storage is empty")
	ErrStorageFull      = errors.New("storage is full")
	ErrStorageClosed    = errors.New("storage is closed")
	ErrBlankPath        = errors.New("no storage path was given")
	ErrBlankKey         = errors.New("no key was given")
	ErrInvalidKey       = errors.New("invalid key")
	ErrBlankTableName   = errors.New("no table name was given")
	ErrInvalidType      = errors.New("invalid storage type")
	ErrStorageLimit     = errors.New("unable to connect to the database, storage limit exceeded")
	ErrInvalidConn      = errors.New("invalid database connection")
	ErrMissingParams    = errors.New("storage parameters are missing")
	ErrMissingStatement = errors.New("statement is missing")
	ErrInvalidLength    = errors.New("max queue length must be positive or zero for no limit")
	ErrMissingCmd       = func(cmd string) error { return fmt.Errorf("%s command is missing", cmd) }
)
