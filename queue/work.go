package queue

// ------------------------------------------------------------------------

// Work represents a task in the worker pool.
type Work interface {
	Execute() error  // Execute performs the work.
	OnFailure(error) // OnFailure handles any execution error.
}

// ------------------------------------------------------------------------
