package colly

import (
	"colly/filters"
	"colly/storage/mem"
	"errors"
	"strconv"
	"sync"
)

// ------------------------------------------------------------------------

// Filter represents a number of including/excluding filters.
type Filter struct {
	incl map[string]*filterItem
	excl map[string]*filterItem
	lock *sync.RWMutex
}

// FilterEngine privides the function to match the filter.
type FilterEngine interface {
	Match(any) bool // Match returns true if the filter is a match.
	// MatchError() error // MatchError returns the error that will be used for exclusive match.
}

// FilterMethod tells whether the filter supposed to include or exclude matches.
type FilterMethod bool

// FilterScope points out which part of the URL will be matched.
type FilterScope uint8

// filterItem represent an including/excluding URL filter
type filterItem struct {
	scope  FilterScope
	engine FilterEngine
	err    error
}

// ------------------------------------------------------------------------

const (
	FILTER_METHOD_INCLUDE FilterMethod = true
	FILTER_METHOD_EXCLUDE FilterMethod = false
)

const (
	DOMAIN_FILTER FilterScope = iota
	URL_FILTER
	DEPTH_FILTER
	REQUEST_FILTER
)

// ------------------------------------------------------------------------

var (
	ErrFilterItemReplaced     = errors.New("a filter item with the same label was overwritten") // ErrFilterItemReplaced when an existing filter item with the same label was overwritten.
	ErrFilterNoEngine         = errors.New("no filter engine was given")                        // ErrFilterNoEngine when an attampt was made to add a new filter item with no filter engine.
	ErrFilterURLDisallowed    = errors.New("URL is not allowed")                                // ErrFilterURLDisallowed is thrown for attempting to visit a URL that is not allowed.
	ErrFilterDomainDisallowed = errors.New("domain is not allowed")                             // ErrFilterDomainDisallowed is thrown for attempting to visit a domain that is not allowed.
	ErrFilterNoMatch          = errors.New("no matching filter")                                // ErrFilterNoMatch is thrown if no matching inclusive filter found.
	ErrFilterURLLength        = errors.New("URL is too long or too short")                      // ErrFilterURLLength is thrown when the URL length is outside of the limits.
	ErrFilterNoRevisit        = errors.New("the URL cannot be revisited")                       // ErrFilterNoRevisit is thrown when the number of revisits exhausted.
	ErrFilterNoRequest        = errors.New("request is missing, nothing to check")              // ErrFilterNoRequest is thrown when the request attribute of the Match function is nil.
	ErrFilterMaxDepth         = errors.New("maximum request depth limit reached")               // ErrFilterMaxDepth is thrown when the maximum request depth limit reached.
)

// ------------------------------------------------------------------------

// NewFilter returns a pointer to a newly created filter.
func NewFilter() *Filter {
	return &Filter{
		incl: map[string]*filterItem{},
		excl: map[string]*filterItem{},
		lock: &sync.RWMutex{},
	}
}

// ------------------------------------------------------------------------

// AddDomainGlob is a convenience method to add domain glob engine to the filter.
func (f *Filter) AddDomainGlob(method FilterMethod, globFilters []string, label ...string) error {
	engine, err := filters.NewGlobEngine(globFilters)
	if err != nil {
		return err
	}

	return f.AddEngine(method, DOMAIN_FILTER, engine, ErrFilterDomainDisallowed, label...)
}

// ------------------------------------------------------------------------

// AddURLGlob is a convenience method to add URL glob engine to the filter.
func (f *Filter) AddURLGlob(method FilterMethod, globFilters []string, label ...string) error {
	engine, err := filters.NewGlobEngine(globFilters)
	if err != nil {
		return err
	}

	return f.AddEngine(method, URL_FILTER, engine, ErrFilterURLDisallowed, label...)
}

// ------------------------------------------------------------------------

// AddDomainRegexp is a convenience method to add domain regexp engine to the filter.
func (f *Filter) AddDomainRegexp(method FilterMethod, regexpFilters []string, label ...string) error {
	engine, err := filters.NewRegexpEngine(regexpFilters)
	if err != nil {
		return err
	}

	return f.AddEngine(method, DOMAIN_FILTER, engine, ErrFilterDomainDisallowed, label...)
}

// ------------------------------------------------------------------------

// AddURLRegexp is a convenience method to add URL regexp engine to the filter.
func (f *Filter) AddURLRegexp(method FilterMethod, regexpFilters []string, label ...string) error {
	engine, err := filters.NewRegexpEngine(regexpFilters)
	if err != nil {
		return err
	}

	return f.AddEngine(method, URL_FILTER, engine, ErrFilterURLDisallowed, label...)
}

// ------------------------------------------------------------------------

// AddURLLength is a convenience method to add URL length engine to the filter.
func (f *Filter) AddURLLength(minLength uint, maxLength uint, label ...string) error {
	return f.AddEngine(FILTER_METHOD_EXCLUDE, URL_FILTER, filters.NewURLLengthEngine(minLength, maxLength), ErrFilterURLLength, label...)
}

// ------------------------------------------------------------------------

// AddRequestDepth is a convenience method to add request depth engine to the filter.
func (f *Filter) AddRequestDepth(maxDepth uint, label ...string) error {
	return f.AddEngine(FILTER_METHOD_EXCLUDE, URL_FILTER, filters.NewRequestDepthEngine(maxDepth), ErrFilterMaxDepth, label...)
}

// ------------------------------------------------------------------------

// AddRevisit is a convenience method to add URL revisit engine to the filter.
func (f *Filter) AddRevisit(maxRevisits uint, storage filters.VisitStorage, label ...string) error {
	if storage == nil {
		storage = mem.NewVisitStorage()
	}

	engine, err := filters.NewRevisitEngine(storage, maxRevisits)
	if err != nil {
		return err
	}

	return f.AddEngine(FILTER_METHOD_EXCLUDE, URL_FILTER, engine, ErrFilterNoRevisit, label...)
}

// ------------------------------------------------------------------------

// Add adds a new filter item to the filter.
func (f *Filter) AddEngine(method FilterMethod, scope FilterScope, engine FilterEngine, err error, label ...string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	key, err := f.setKey(method, label)

	if method == FILTER_METHOD_INCLUDE {
		f.incl[key] = &filterItem{
			scope:  scope,
			engine: engine,
			err:    err,
		}

		return err
	}

	f.excl[key] = &filterItem{
		scope:  scope,
		engine: engine,
	}

	return err
}

// ------------------------------------------------------------------------

// Has returns true if a filter item exists by label and optional method.
func (f *Filter) Has(label string, method ...FilterMethod) bool {
	hasMethod := len(method) > 0
	isInclude := hasMethod && method[0] == FILTER_METHOD_INCLUDE

	if !hasMethod || isInclude {
		if _, present := f.incl[label]; present || isInclude {
			return present
		}
	}

	_, present := f.excl[label]

	return present
}

// ------------------------------------------------------------------------

// Remove removes a filter with a specific label and optional method.
func (f *Filter) Remove(label string, method ...FilterMethod) {
	hasMethod := len(method) > 0
	isInclude := hasMethod && method[0] == FILTER_METHOD_INCLUDE

	f.lock.Lock()
	defer f.lock.Unlock()

	if !hasMethod || isInclude {
		if delete(f.incl, label); isInclude {
			return
		}
	}

	delete(f.excl, label)
}

// ------------------------------------------------------------------------

// RemoveByScope removes all filters with a specific scope and optional method.
func (f *Filter) RemoveByScope(scope FilterScope, method ...FilterMethod) {
	hasMethod := len(method) > 0
	isInclude := hasMethod && method[0] == FILTER_METHOD_INCLUDE

	f.lock.Lock()
	defer f.lock.Unlock()

	if !hasMethod || isInclude {
		for key, item := range f.incl {
			if item.scope == scope {
				delete(f.incl, key)
			}
		}
		if isInclude {
			return
		}
	}

	for key, item := range f.excl {
		if item.scope == scope {
			delete(f.excl, key)
		}
	}
}

// ------------------------------------------------------------------------

// Match returns error if the Request matches any exclusive fiter or
// inclusive filters exist and the Request doesn't match any of them.
// Excluding filters will be evaluated before including filters.
// The optional tags will only check filters with matching tag.
func (f *Filter) Match(req *Request, tags ...string) error {
	if req == nil {
		return ErrFilterNoRequest
	}

	segments := map[FilterScope]any{}
	checkTag := len(tags) > 0

	f.lock.RLock()
	defer f.lock.RUnlock()

	// Check the exclusions first
	for key, item := range f.excl {
		if checkTag && !InSlice(key, tags) {
			continue
		}

		if _, present := segments[item.scope]; !present {
			segments[item.scope] = item.segment(req)
		}

		if item.engine.Match(segments[item.scope]) {
			return item.err
		}
	}

	// If no inclusive filter, everything is allowed
	if len(f.incl) == 0 {
		return nil
	}

	// Check for any matching inclusive filter
	for key, item := range f.incl {
		if checkTag && !InSlice(key, tags) {
			continue
		}

		if _, present := segments[item.scope]; !present {
			segments[item.scope] = item.segment(req)
		}

		if item.engine.Match(segments[item.scope]) {
			return nil
		}

	}

	return ErrFilterNoMatch
}

// ------------------------------------------------------------------------

// Count returns the number of filter items attached to this filter.
func (f *Filter) Count() int {
	return len(f.incl) + len(f.excl)
}

// ------------------------------------------------------------------------

// IsEmpty returns true if no filter items attached to this filter.
func (f *Filter) IsEmpty() bool {
	return len(f.incl) == 0 && len(f.excl) == 0
}

// ------------------------------------------------------------------------

func (f *Filter) setKey(method FilterMethod, label []string) (string, error) {
	var (
		key  string
		list *map[string]*filterItem
	)

	if method == FILTER_METHOD_INCLUDE {
		list = &f.incl
	} else {
		list = &f.excl
	}

	if len(label) > 0 {
		key = label[0]
	} else {
		key = strconv.Itoa(1 + len(*list))
	}

	if f.Has(key, method) {
		return key, ErrFilterItemReplaced
	}

	return key, nil
}

// ------------------------------------------------------------------------

func (i *filterItem) segment(req *Request) any {
	switch i.scope {
	case DOMAIN_FILTER:
		if req.Req == nil || req.Req.URL == nil {
			return nil
		}
		return req.Req.URL.Hostname()
	case URL_FILTER:
		if req.Req == nil || req.Req.URL == nil {
			return nil
		}
		return req.Req.URL.String()
	case DEPTH_FILTER:
		return req.Depth
	default:
		return req
	}
}
