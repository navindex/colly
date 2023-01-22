package badger

import (
	"colly/storage"
	"sync"

	"github.com/dgraph-io/badger/v3"
)

// ------------------------------------------------------------------------

// dbconn encapsulates the SQLite3 database handle
type dbconn struct {
	path     string
	dbh      *badger.DB // Database handle
	useCount uint16
}

// / stgBase is a generic BadgerDB storage
type stgBase struct {
	db     *dbconn
	config *config
	closed bool
}

// Storage config
type config struct {
	prefix      []byte
	clearOnOpen bool
}

type dataType byte

// ------------------------------------------------------------------------

const (
	TYPE_VISIT dataType = iota
	TYPE_COOKIE
	TYPE_FIFO
	TYPE_CACHE
)

// ------------------------------------------------------------------------

// Database list indexed by path
var connections = map[string]*dbconn{}

// Maximum number of storages connected to the same database.
// 0 value means no limit.
var maxUseCount uint16 = 100

var connLock = &sync.Mutex{}

// ------------------------------------------------------------------------

// connect attaches a storage to a database
func connect(path string) (*dbconn, error) {
	if path == "" {
		return nil, storage.ErrBlankPath
	}

	opt := badger.DefaultOptions(path)

	connLock.Lock()
	defer connLock.Unlock()

	conn, present := connections[path]
	if !present {
		dbh, err := badger.Open(opt)
		if err != nil {
			return nil, err
		}

		conn = &dbconn{
			path:     path,
			dbh:      dbh,
			useCount: 0,
		}
		connections[path] = conn
	}

	if maxUseCount > 0 && conn.useCount >= maxUseCount {
		return nil, storage.ErrStorageLimit
	}
	conn.useCount++

	return conn, nil
}

// ------------------------------------------------------------------------

// disconnect detaches a storage from the database
// and closes the database if no more storages are connected
func (dbc *dbconn) disconnect() {
	connLock.Lock()
	defer connLock.Unlock()

	dbc.useCount--

	// Remove dbc if this was the last connected storage
	if dbc.useCount <= 0 {
		dbc.dbh.Close()
		delete(connections, dbc.path)
	}
}

// ------------------------------------------------------------------------

// NewBadgerStorage returns a pointer to a newly created BadgerDB storage.
func NewBaseStorage(path string, config *config) (*stgBase, error) {
	if config == nil || len(config.prefix) == 0 {
		return nil, storage.ErrMissingParams
	}

	db, err := connect(path)
	if err != nil {
		return nil, err
	}

	s := &stgBase{
		db:     db,
		config: config,
		closed: false,
	}

	// Clear the data if required
	if s.config.clearOnOpen {
		if err := s.DropPrefix(s.config.prefix); err != nil {
			s.db.disconnect()

			return nil, err
		}
	}

	return s, nil
}

// ------------------------------------------------------------------------

// Close closes the BadgerDB storage.
func (s *stgBase) Close() error {
	s.db.disconnect()
	s.db = nil
	s.closed = true

	return nil
}

// ------------------------------------------------------------------------

// Clear removes all entries from the SQLite3 storage.
func (s *stgBase) Clear() error {
	return s.DropPrefix(s.config.prefix)
}

// ------------------------------------------------------------------------

// DropPrefix drops all the keys with the provided prefix.
func (s *stgBase) DropPrefix(prefix []byte) error {
	return s.db.dbh.DropPrefix(append(s.config.prefix, prefix...))
}

// ------------------------------------------------------------------------

// Set adds a key-value pair to the storage.
func (s *stgBase) Set(key, value []byte) error {
	if len(key) == 0 {
		return storage.ErrBlankKey
	}

	prefixedKey := append(s.config.prefix, key...)

	return s.db.dbh.Update(func(txn *badger.Txn) error {
		return txn.Set(prefixedKey, value)
	})
}

// ------------------------------------------------------------------------

// Set adds a key with a boolean value to the storage.
func (s *stgBase) SetBool(key []byte, value bool) error {
	byteVal := []byte{0}
	if value {
		byteVal = []byte{1}
	}

	return s.Set(key, byteVal)
}

// ------------------------------------------------------------------------

// Get looks for key and returns the corresponding value.
// If key is not found, nil will be returned.
func (s *stgBase) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, storage.ErrBlankKey
	}

	var (
		value       []byte
		prefixedKey = append(s.config.prefix, key...)
	)

	err := s.db.dbh.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefixedKey)
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(value)

		return err
	})

	if err == badger.ErrKeyNotFound {
		value = nil
		err = nil
	}

	return value, err
}

// ------------------------------------------------------------------------

// GetBool looks for key and returns the corresponding boolean value.
// If key is not found, false will be returned.
func (s *stgBase) GetBool(key []byte) (bool, error) {
	value, err := s.Get(key)
	if err != nil {
		return false, err
	}

	return len(value) > 0 && value[0] == 1, nil
}

// ------------------------------------------------------------------------

// Len returns the number of entries in the BadgerDB storage.
func (s *stgBase) Len() (uint, error) {
	var count uint

	opt := badger.DefaultIteratorOptions

	if err := s.db.dbh.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)
		defer it.Close()

		for it.Rewind(); it.ValidForPrefix(s.config.prefix); it.Next() {
			count++
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return count, nil
}
