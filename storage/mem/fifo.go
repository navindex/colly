// In-memory FIFO (First In First Out) storage.
package mem

import (
	"bytes"
	"colly/storage"
	"io"
	"sync"
)

// ------------------------------------------------------------------------

// stgMultiFIFO is an in-memory multi-thread FIFO storage
type stgMultiFIFO struct {
	threads  map[uint32]*stgFIFO
	capacity uint
	lock     *sync.RWMutex
}

// stgFIFO is a FIFO storage
type stgFIFO struct {
	head  *dataNode
	tail  *dataNode
	count uint
	lock  *sync.Mutex
}

// dataNode is an item in the FIFO storage
type dataNode struct {
	data []byte
	next *dataNode
}

// ------------------------------------------------------------------------

// NewFIFOStorage returns a pointer to a newly created in-memory FIFO storage.
func NewFIFOStorage(capacity uint) *stgMultiFIFO {
	return &stgMultiFIFO{
		threads:  map[uint32]*stgFIFO{},
		capacity: capacity,
		lock:     &sync.RWMutex{},
	}
}

// ------------------------------------------------------------------------

// Close method is required to implement the Queue interface.
func (s *stgMultiFIFO) Close() error {
	return s.Clear()
}

// ------------------------------------------------------------------------

// Clear removes all entries from a number of threads of the in-memory FIFO storage,
// or removes all entries from all threads if no ID was given.
func (s *stgMultiFIFO) Clear(ids ...uint32) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if len(ids) == 0 {
		s.threads = map[uint32]*stgFIFO{}

		return nil
	}

	for _, id := range ids {
		delete(s.threads, id)
	}

	return nil
}

// ------------------------------------------------------------------------

// Capacity returns the maximum number of items that can be stored in the FIFO storage.
func (s *stgMultiFIFO) Capacity() uint {
	return s.capacity
}

// ------------------------------------------------------------------------

// Len returns the number of items in the FIFO storage.
func (s *stgMultiFIFO) Len(id uint32) (uint, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if t, present := s.threads[id]; present {
		return t.len(), nil
	}

	return 0, nil
}

// ------------------------------------------------------------------------

// Push appends a value at the end/tail of the queue.
// Note: this function does mutate the queue.
func (s *stgMultiFIFO) Push(id uint32, item io.Reader) error {
	data, err := io.ReadAll(item)
	if err != nil {
		return err
	}

	s.addThread(id)

	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.threads[id].push(data, s.capacity)
}

// ------------------------------------------------------------------------

// Pop removes and returns the oldest value in the queue.
// Note: this function does mutate the queue.
func (s *stgMultiFIFO) Pop(id uint32) (io.Reader, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.hasThread(id) {
		return nil, storage.ErrStorageEmpty
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.threads[id].pop()
}

// ------------------------------------------------------------------------

// Peek returns the oldest value in the queue without removing it.
// Note: this function does NOT mutate the queue.
func (s *stgMultiFIFO) Peek(id uint32) (io.Reader, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if !s.hasThread(id) {
		return nil, storage.ErrStorageEmpty
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.threads[id].peek()
}

// ------------------------------------------------------------------------

// The addThread method adds a new thread if it doesn't exist.
func (s *stgMultiFIFO) addThread(id uint32) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !s.hasThread(id) {
		s.threads[id] = &stgFIFO{
			head:  nil,
			tail:  nil,
			count: 0,
			lock:  &sync.Mutex{},
		}
	}
}

// The hasThread method returns true if a thread with the ID exists.
func (s *stgMultiFIFO) hasThread(id uint32) bool {
	_, present := s.threads[id]

	return present
}

// ------------------------------------------------------------------------

// The len method returns the number of items in the FIFO thread.
func (s *stgFIFO) len() uint {
	s.lock.Lock()
	defer s.lock.Unlock()

	return s.count
}

// The push method appends a value at the end/tail of the queue.
// Note: this function does mutate the queue.
func (s *stgFIFO) push(data []byte, capacity uint) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.count >= capacity {
		return storage.ErrStorageFull
	}

	node := &dataNode{
		data: data,
	}

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

// The pop method removes and returns the oldest value in the thread.
// Note: this function does mutate the queue.
func (s *stgFIFO) pop() (io.Reader, error) {
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

	return bytes.NewReader(node.data), nil
}

// The peek method returns the oldest value in the thread without removing it.
// Note: this function does NOT mutate the queue.
func (s *stgFIFO) peek() (io.Reader, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.head == nil {
		return nil, storage.ErrStorageEmpty
	}

	return bytes.NewReader(s.head.data), nil
}
