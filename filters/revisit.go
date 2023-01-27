package filters

import "errors"

// ------------------------------------------------------------------------

// VisitStorage is a Storage to save and retreive visiting information.
type VisitStorage interface {
	AddVisit(key string) error           // AddVisit stores an URL that is visited.
	PastVisits(key string) (uint, error) // PastVisits returns how many times the URL was visited before.
	Remove(key string) error             // Remove removes an entry by URL.
	Clear() error                        // Clear deletes all stored items.
}

// revisitFilter represents a filter that checks how many times the URL was visited
type revisitFilter struct {
	maxRevisits uint
	stg         VisitStorage
}

// ------------------------------------------------------------------------

// ErrFilterNoStorage is thrown when no storage attribute was given.
var ErrFilterNoStorage = errors.New("invalid or missing storage")

// ------------------------------------------------------------------------

// NewRevisitEngine returns a pointer to a newly created filter that check
// whether or not the URL is eligible for a new visit.
// This filter should be used with FILTER_METHOD_EXCLUDE method.
func NewRevisitEngine(storage VisitStorage, maxRevisits uint) (*revisitFilter, error) {
	if storage == nil {
		return nil, ErrFilterNoStorage
	}

	return &revisitFilter{
		maxRevisits: maxRevisits,
		stg:         storage,
	}, nil
}

// ------------------------------------------------------------------------

// Match returns false if the URL can be revisited.
func (f *revisitFilter) Match(u any) bool {
	str, ok := u.(string)
	if !ok {
		return false
	}

	visited, err := f.stg.PastVisits(str)

	return err == nil || visited > f.maxRevisits
}
