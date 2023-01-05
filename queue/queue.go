package queue

import (
	"errors"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/domain"
	"github.com/gocolly/colly/v2/storage/mem"
)

// ------------------------------------------------------------------------

// WorkerPool is a collection of methods to control a worker pool.
type WorkerPool interface {
	Start()              // Start gets the worker pool ready to process the tasks. It should be called only once.
	Stop()               // Stop instructs the worker pool to stop processing tasks. It should be called only once.
	AddWork(Work)        // AddWork adds a task to the worker pool.
	Size() (uint, error) // Size returns the number of tasks in the worker pool.
	IsEmpty() bool       // IsEmpty returns true if the queue is empty.
}

// queue implements a domain.Queue interface
type queue struct {
	stg       domain.QueueStorage // thread safe queue storage
	processor domain.JobHandler   // responsible for processing the jobs
	threads   uint                // the number of processing threads
	abort     bool                // if true, instructs the queue to stop the job processing
	lock      *sync.Mutex         // guards wake and running ???
	wakeChan  chan struct{}
}

// ------------------------------------------------------------------------

const maxLength uint = 100000

// ------------------------------------------------------------------------

var (
	ErrAlreadyStarted = errors.New("the queue is already being processed")
)

// ------------------------------------------------------------------------

// New returns a pointer to a newly created request queue.
// A memory storage is used if no storage was given.
// A WHATWGP parser is used if no parser was given.
func New(threads uint, stg domain.QueueStorage) (*queue, error) {
	if stg == nil {
		var err error
		if stg, err = mem.NewQueueStorage(maxLength); err != nil {
			return nil, err
		}
	}

	return &queue{
		threads: threads,
		stg:     stg,
		lock:    &sync.Mutex{},
		abort:   false,
	}, nil
}

// ------------------------------------------------------------------------

// Size returns the number of items in the queue.
func (q *queue) Size() (uint, error) {
	return q.stg.Len()
}

// ------------------------------------------------------------------------

// IsEmpty returns true if the queue is empty or an error occurred.
func (q *queue) IsEmpty() bool {
	s, err := q.Size()

	return err != nil || s == 0
}

// ------------------------------------------------------------------------

// AddItem adds a new item to the queue.
func (q *queue) AddItem(item domain.Job) error {
	// Convert the item to bytes
	bytes, err := item.Encode()
	if err != nil {
		return err
	}

	// Add it to the queue
	if err := q.stg.Push(bytes); err != nil {
		return err
	}

	// Check if the processor is running
	q.lock.Lock()
	processing := q.wakeChan != nil
	q.lock.Unlock()

	// ???
	if !processing {
		q.wakeChan <- struct{}{}
	}

	return nil
}

// ------------------------------------------------------------------------

// Run starts consumer threads and calls the Collector to perform requests.
// Run blocks while the queue has active requests.
// The underlying Storage must not be used directly while Run blocks.
func (q *queue) Start(c *colly.Collector) error {
	if err := q.prepareProcess(); err != nil {
		return err
	}

	requestChan := make(chan *colly.Request)
	defer close(requestChan)

	completeChan, errChan := make(chan struct{}), make(chan error, 1)

	for i := 0; i < q.threads; i++ {
		go independentRunner(requestChan, completeChan)
	}

	go q.loop(c, requestChan, completeChan, errChan)

	return <-errc
}

// ------------------------------------------------------------------------

// Stop will stop the running queue processor.
func (q *queue) Stop() {
	q.lock.Lock()
	q.abort = true
	q.lock.Unlock()
}

// ------------------------------------------------------------------------

func (q *queue) prepareProcess() error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.wakeChan != nil && q.abort == false {
		return ErrAlreadyStarted
	}

	q.wakeChan = make(chan struct{})
	q.abort = false

	return nil
}

// ------------------------------------------------------------------------

func (q *queue) loop(c *colly.Collector, requestc chan<- *colly.Request, complete <-chan struct{}, errc chan<- error) {
	var active int

	for {
		size, err := q.storage.Len()
		if err != nil {
			errc <- err
			break
		}

		if size == 0 && active == 0 || !q.running {
			// Terminate when
			//   1. No active requests
			//   2. Emtpy queue
			errc <- nil
			break
		}

		sent := requestc
		var req *colly.Request
		if size > 0 {
			req, err = q.loadRequest(c)
			if err != nil {
				// ignore an error returned by GetRequest() or
				// UnmarshalRequest()
				continue
			}
		} else {
			sent = nil
		}

	Sent:
		for {
			select {
			case sent <- req:
				active++
				break Sent
			case <-q.wake:
				if sent == nil {
					break Sent
				}
			case <-complete:
				active--
				if sent == nil && active == 0 {
					break Sent
				}
			}
		}
	}
}
