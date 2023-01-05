package queue

import (
	"errors"
	"sync"

	"github.com/gocolly/colly/v2/storage/mem"
)

// ------------------------------------------------------------------------

// Storage is a thread safe FIFO Storage to serve the worker pool.
type Storage interface {
	Push([]byte, <-chan struct{}) error   // Push inserts an item into the storage.
	Peek(<-chan struct{}) ([]byte, error) // Peek returns the oldest item from the storage without removing it.
	Pop(<-chan struct{}) ([]byte, error)  // Pop pops the oldest item from the storage or returns error if the storage is empty.
	Len(<-chan struct{}) (uint, error)    // Len returns the number of entries in the storage.
}

type pool struct {
	tasks      chan Work
	stg        Storage       // thread safe FIFO storage
	numWorkers uint          // the number of workers
	start      sync.Once     // ensure the pool can be started once
	stop       sync.Once     // ensure the pool can be stopped once
	abort      chan struct{} // close this channel to instruct the workers to stop working
}

// ------------------------------------------------------------------------

const maxMemStgLen uint = 100000

// ------------------------------------------------------------------------

var (
	ErrInvalidNumWorkers = errors.New("invalid number of workers")
)

// ------------------------------------------------------------------------

// NewPool returns a pointer to a newly created worker pool.
// A memory storage is used if no storage was given.
func NewPool(threads uint, stg Storage) *pool {
	if stg == nil {
		stg = mem.NewFIFOStorage(maxMemStgLen)
	}

	return &pool{
		tasks:      make(chan Work, maxMemStgLen),
		stg:        stg,
		numWorkers: threads,
		start:      sync.Once{},
		stop:       sync.Once{},
		abort:      make(chan struct{}),
	}
}

// ------------------------------------------------------------------------

// Start gets the worker pool ready to process the tasks.
func (p *pool) Start() {
	p.start.Do(func() {
		p.startWorkers()
	})
}

// ------------------------------------------------------------------------

// Stop instructs the worker pool to stop processing tasks.
func (p *pool) Stop() {
	p.start.Do(func() {
		close(p.abort)
	})
}

// ------------------------------------------------------------------------

// AddWorkBlocking adds a task to the worker pool.
// If the task limit reached, and all workers are occupied,
// this will hang until the task is consumed or Stop called.
func (p *pool) AddWorkBlocking(w Work) {
	select {
	case p.tasks <- w:
	case <-p.abort:
	}
}

// ------------------------------------------------------------------------

// AddWork adds a task to the worker pool.
func (p *pool) AddWork(w Work) {
	go p.AddWorkBlocking(w)
}

// ------------------------------------------------------------------------

// Size returns the number of items in the queue.
func (p *pool) Size() (uint, error) {
	return p.stg.Len()
}

// ------------------------------------------------------------------------

// IsEmpty returns true if the queue is empty or an error occurred.
func (p *pool) IsEmpty() bool {
	s, err := p.Size()

	return err != nil || s == 0
}

// ------------------------------------------------------------------------

func (p *pool) startWorkers() {
	for i := uint(0); i < p.numWorkers; i++ {
		go func(workerNum uint) {
			for {
				select {
				case <-p.abort:
					return
				case task, ok := <-p.tasks:
					if !ok {
						return
					}
					if err := task.Execute(); err != nil {
						task.OnFailure(err)
					}
				}
			}
		}(i)
	}
}
