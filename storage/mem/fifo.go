// In-memory FIFO (First In First Out) storage.
package mem

import (
	"sync"

	"github.com/gocolly/colly/v2/storage"
)

// ------------------------------------------------------------------------

// in-memory FIFO storage
type stgFIFO struct {
	head     *dataNode
	tail     *dataNode
	count    uint
	maxCount uint
	lock     *sync.Mutex
	closed   bool
}

// Data node
type dataNode struct {
	data []byte
	next *dataNode
}

// ------------------------------------------------------------------------

// NewFIFOStorage returns a pointer to a newly created in-memory FIFO storage.
func NewFIFOStorage(maxLength uint) *stgFIFO {
	return &stgFIFO{
		maxCount: maxLength,
		lock:     &sync.Mutex{},
		closed:   false,
	}
}

// ------------------------------------------------------------------------

// Close closes the in-memory FIFO storage.
func (s *stgFIFO) Close() error {
	if s.closed {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.tail = nil
	s.head = nil
	s.count = 0
	s.closed = true

	return nil
}

// ------------------------------------------------------------------------

// Clear removes all entries from the in-memory ueue storage.
func (s *stgFIFO) Clear() error {
	if s.closed {
		return storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.tail = nil
	s.head = nil
	s.count = 0

	return nil
}

// ------------------------------------------------------------------------

// Len returns the number of items in the FIFO storage.
func (s *stgFIFO) Len() (uint, error) {
	if s.closed {
		return 0, storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return s.count, nil
}

// ------------------------------------------------------------------------

// Push appends a value at the end/tail of the queue.
// Note: this function does mutate the queue.
func (s *stgFIFO) Push(item []byte) error {
	if s.closed {
		return storage.ErrStorageClosed
	}

	if s.count >= s.maxCount {
		return storage.ErrStorageFull
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	node := &dataNode{data: item}

	if s.tail == nil {
		s.tail = node
		s.head = node
	} else {
		s.tail.next = node
		s.tail = node
	}

	s.count++

	return nil
}

// ------------------------------------------------------------------------

// Pop removes and returns the oldest value in the queue.
// Note: this function does mutate the queue.
func (s *stgFIFO) Pop() ([]byte, error) {
	if s.closed {
		return nil, storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.head == nil {
		return nil, storage.ErrStorageEmpty
	}

	node := s.head
	s.head = node.next

	if s.head == nil {
		s.tail = nil
	}

	s.count--

	return node.data, nil
}

// ------------------------------------------------------------------------

// Peek returns the oldest value in the queue without removing it.
// Note: this function does NOT mutate the queue.
func (s *stgFIFO) Peek() ([]byte, error) {
	if s.closed {
		return nil, storage.ErrStorageClosed
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	node := s.head
	if node == nil {
		return nil, storage.ErrStorageEmpty
	}

	return node.data, nil
}
