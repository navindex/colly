package colly

import (
	"colly/storage/mem"
	"io"
)

// ------------------------------------------------------------------------

// Queue is a thread safe FIFO storage.
type Queue interface {
	Clear(...uint32) error         // Clear removes all entries from a number dispatch queues or the whole queue if no ID was given.
	Close() error                  // Close closes the queue.
	Len(uint32) (uint, error)      // Len returns the number of items in a dispatch queue.
	Push(uint32, io.Reader) error  // Push appends a value at the end/tail of a dispatch queue.
	Pop(uint32) (io.Reader, error) // Pop removes and returns the oldest value in a dispatch queue.
	Capacity() uint                // Capacity returns the maximum capcity of a dispatch queue.
}

// Job represents a queue item.
type Job interface {
	Encode() (io.Reader, error) // Encode converts the job to bytes.
}

// JobDecoder is a function to decode bytes to a job.
type JobDecoder func(io.Reader) (any, error)

// JobQueue manages adds and removes elements in the job queue.
type JobQueue interface {
	Push(Job) error    // Push appends a job at the end/tail of the queue.
	Pop() (Job, error) // Pop removes and returns the oldest job in the queue.
}

type jobQueue struct {
	id      uint32
	stg     Queue
	decoder JobDecoder
}

// ------------------------------------------------------------------------

const defJobQueueCapacity uint = 100000

// ------------------------------------------------------------------------

// NewJobQueue returns a pointer to a newly created job queue.
func NewJobQueue(id uint32, decoder JobDecoder, storage Queue) (*jobQueue, error) {
	if decoder == nil {
		return nil, ErrNoJobDecoder
	}

	if storage == nil {
		storage = mem.NewFIFOStorage(defJobQueueCapacity)
	}

	return &jobQueue{
		id:      id,
		stg:     storage,
		decoder: decoder,
	}, nil
}

// ------------------------------------------------------------------------

// Clone creates a copy of the job queue with a new ID.
// The new job queue uses the same storage and the same decoder.
func (q *jobQueue) Clone(id uint32) *jobQueue {
	return &jobQueue{
		id:      id,
		stg:     q.stg,
		decoder: q.decoder,
	}
}

// ------------------------------------------------------------------------

// Storage returns the storage behind the job queue.
// The same storage can serve multiple job queues.
func (q *jobQueue) Storage() Queue {
	return q.stg
}

// ------------------------------------------------------------------------

// Len returns the number of items in the queue.
func (q *jobQueue) Len() (uint, error) {
	return q.stg.Len(q.id)
}

// ------------------------------------------------------------------------

// IsEmpty returns true if the queue is empty.
func (q *jobQueue) IsEmpty() bool {
	len, err := q.stg.Len(q.id)
	return err != nil && len == 0
}

// ------------------------------------------------------------------------

// Push appends a job at the end/tail of the queue.
func (q *jobQueue) Push(job Job) error {
	rdr, err := job.Encode()
	if err != nil {
		return err
	}

	return q.stg.Push(q.id, rdr)
}

// ------------------------------------------------------------------------

// Pop removes and returns the oldest job in the queue.
func (q *jobQueue) Pop() (any, error) {
	rdr, err := q.stg.Pop(q.id)
	if err != nil {
		return nil, err
	}

	return q.decoder(rdr)
}
