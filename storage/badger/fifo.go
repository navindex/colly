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

// Capacity returns the maximum number of items that can be stored in the FIFO storage.
func (s *stgFIFO) Capacity() uint {
	return 1000000000
}

// ------------------------------------------------------------------------

// Clear removes all entries from the BadgerDB FIFO storage.
func (s *stgFIFO) Clear(ids ...uint32) error {
	if len(ids) == 0 {
		return s.s.Clear()
	}

	for _, id := range ids {
		s.s.DropPrefix(s.prefixedID(id))
	}

	return nil
}

// ------------------------------------------------------------------------

// Len returns the number of request queues in the BadgerDB FIFO storage.
func (s *stgFIFO) Len(id uint32) (uint, error) {
	return s.s.Len(encodeID(id))
}

// ------------------------------------------------------------------------

// Push inserts an item into the BadgerDB FIFO storage.
func (s *stgFIFO) Push(id uint32, item io.Reader) error {
	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}

	key := append(encodeID(id), encodeTime(time.Now())...)

	return s.s.Set(key, data)
}

// ------------------------------------------------------------------------

// Pop pops the oldest item from the FIFO storage or returns error if the storage is empty.
func (s *stgFIFO) Pop(id uint32) (io.Reader, error) {
	return s.headValue(encodeID(id), true)
}

// ------------------------------------------------------------------------

// Pop pops maxmum n of the oldest item from the FIFO storage
// or returns error if the storage is empty.
// func (s *stgFIFO) MultiPop(n uint) ([]io.Reader, error) {
// 	if n < 1 {
// 		return nil, storage.ErrInvalidNumber
// 	}

// 	items := []io.Reader{}
// 	for i := uint(0); i < n; i++ {
// 		item, err := s.headValue(true)
// 		if err != nil {
// 			return items, err
// 		}
// 		items = append(items, item)
// 	}

// 	return items, nil
// }

// ------------------------------------------------------------------------

// Peek returns the oldest item from the queue without removing it.
func (s *stgFIFO) Peek(id uint32) (io.Reader, error) {
	return s.headValue(encodeID(id), false)
}

// ------------------------------------------------------------------------

func (s *stgFIFO) headKey(prefix []byte) ([]byte, error) {
	var key []byte

	p := append(s.s.config.prefix, prefix...)
	opt := badger.DefaultIteratorOptions
	err := s.s.db.dbh.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.ValidForPrefix(p); it.Next() {
			iKey := it.Item().Key()
			if bytes.Compare(key, iKey) == -1 {
				copy(key, iKey)
			}
		}
		return nil
	})

	return key, err
}

func (s *stgFIFO) headValue(prefix []byte, remove bool) (io.Reader, error) {
	if prefix == nil {
		return nil, storage.ErrBlankKey
	}

	// Find the head key
	key, err := s.headKey(prefix)
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

func (s *stgFIFO) prefixedID(id uint32) []byte {
	return append(s.s.config.prefix, encodeID(id)...)
}

// ------------------------------------------------------------------------

// encodeTime converts the time to 8 bytes
func encodeTime(t time.Time) []byte {
	b := []byte{}
	binary.BigEndian.PutUint64(b, uint64(t.Unix()))

	return b
}

// encodeID converts the thread ID to 4 bytes
func encodeID(id uint32) []byte {
	b := []byte{}
	binary.BigEndian.PutUint32(b, id)

	return b
}
