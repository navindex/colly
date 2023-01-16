package badger

import (
	"colly/storage"
	"net/url"

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
func (s *stgFIFO) Push(item []byte) error {
	return s.s.Set(storage.CurrentTimeToBytes(), item)

}

// ------------------------------------------------------------------------

// Pop pops the oldest item from the FIFO storage or returns error if the storage is empty.
func (s *stgFIFO) Pop(u *url.URL) ([]byte, error) {
	return s.headValue(true)
}

// ------------------------------------------------------------------------

// Peek returns the oldest item from the queue without removing it.
func (s *stgFIFO) Peek() ([]byte, error) {
	return s.headValue(false)
}

// ------------------------------------------------------------------------

func (s *stgFIFO) headKey() ([]byte, error) {
	var key []byte

	opt := badger.DefaultIteratorOptions

	err := s.s.db.dbh.View(func(txn *badger.Txn) error {
		epoch := uint64(0)
		it := txn.NewIterator(opt)
		defer it.Close()

		for it.Rewind(); it.ValidForPrefix(s.s.config.prefix); it.Next() {
			iKey := it.Item().Key()
			iEpoch := storage.BytesToUint64(iKey)

			if iEpoch < epoch || epoch == 0 {
				key, epoch = iKey, iEpoch
			}
		}

		return nil
	})

	return key, err
}

// ------------------------------------------------------------------------

func (s *stgFIFO) headValue(remove bool) ([]byte, error) {
	var (
		key   []byte
		value []byte
		err   error
	)

	s.s.lock.Lock()
	defer s.s.lock.Unlock()

	// Find the head key
	if key, err = s.headKey(); err != nil {
		return nil, err
	}

	if len(key) == 0 {
		return nil, storage.ErrStorageEmpty
	}

	// Get the value
	err = s.s.db.dbh.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(value)

		return err
	})

	// Delete the head key
	if err == nil && remove {
		err = s.s.db.dbh.Update(func(txn *badger.Txn) error {
			return txn.Delete(key)
		})
	}

	return value, err
}
