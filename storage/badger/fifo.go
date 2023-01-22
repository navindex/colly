package badger

import (
	"bytes"
	"colly/storage"
	"encoding/binary"
	"io"
	"time"

	"github.com/dgraph-io/badger/v3"
)

// ------------------------------------------------------------------------

type stgFIFO struct {
	s *stgBase
}

// ------------------------------------------------------------------------

// NewFIFOStorage returns a pointer to a newly created BadgerDB FIFO storage.
func NewFIFOStorage(path string, keepData bool) (*stgFIFO, error) {
	cfg := config{
		prefix:      []byte{byte(TYPE_FIFO), 0},
		clearOnOpen: !keepData,
	}

	s, err := NewBaseStorage(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &stgFIFO{
		s: s,
	}, nil
}

// ------------------------------------------------------------------------

// Close closes the BadgerDB FIFO storage.
func (s *stgFIFO) Close() error {
	return s.s.Close()
}

// ------------------------------------------------------------------------

// Clear removes all entries from the BadgerDB FIFO storage.
func (s *stgFIFO) Clear() error {
	return s.s.Clear()
}

// ------------------------------------------------------------------------

// Len returns the number of request queues in the BadgerDB FIFO storage.
func (s *stgFIFO) Len() (uint, error) {
	return s.s.Len()
}

// ------------------------------------------------------------------------

// Push inserts an item into the BadgerDB FIFO storage.
func (s *stgFIFO) Push(item io.Reader) error {
	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}

	return s.s.Set(encodeTime(time.Now()), data)
}

// ------------------------------------------------------------------------

// Pop pops the oldest item from the FIFO storage or returns error if the storage is empty.
func (s *stgFIFO) Pop() (io.Reader, error) {
	return s.headValue(true)
}

// ------------------------------------------------------------------------

// Peek returns the oldest item from the queue without removing it.
func (s *stgFIFO) Peek() (io.Reader, error) {
	return s.headValue(false)
}

// ------------------------------------------------------------------------

func (s *stgFIFO) headKey() ([]byte, error) {
	var key []byte

	opt := badger.DefaultIteratorOptions
	err := s.s.db.dbh.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(s.s.config.prefix); it.Next() {
			iKey := it.Item().Key()
			if bytes.Compare(key, iKey) == -1 {
				copy(key, iKey)
			}
		}
		return nil
	})

	return key, err
}

// ------------------------------------------------------------------------

func (s *stgFIFO) headValue(remove bool) (io.Reader, error) {
	// Find the head key
	key, err := s.headKey()
	if err != nil {
		return nil, err
	}
	if len(key) == 0 {
		return nil, storage.ErrStorageEmpty
	}

	// Get the value
	var data []byte
	err = s.s.db.dbh.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		data, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return nil, err
	}

	// Delete the head key
	if remove {
		err = s.s.db.dbh.Update(func(txn *badger.Txn) error {
			return txn.Delete(key)
		})
	}

	return bytes.NewReader(data), err
}

// ------------------------------------------------------------------------

// encodeTime converts the time to bytes
func encodeTime(t time.Time) []byte {
	b := []byte{}
	binary.BigEndian.PutUint64(b, uint64(t.Unix()))

	return b
}
