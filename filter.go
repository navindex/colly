package colly

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gobwas/glob"
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
	Match(string) bool
}

// FilterMethod tells whether the filter supposed to include or exclude matches.
type FilterMethod bool

// FilterScope points out which part of the URL will be matched.
type FilterScope uint8

// FilterOperator identifies the the logical operator for combined filter engines.
type FilterOperator uint8

// VisitStorage is a Storage to save and retreive visiting information.
type VisitStorage interface {
	AddVisit(key string) error           // AddVisit stores an URL that is visited.
	PastVisits(key string) (uint, error) // PastVisits returns how many times the URL was visited before.
	Remove(key string) error             // Remove removes an entry by URL.
	Clear() error                        // Clear deletes all stored items.
}

// filterItem represent an including/excluding URL filter
type filterItem struct {
	scope  FilterScope
	engine FilterEngine
}

// globFilter represents a number of glob expression filters
type globFilter struct {
	globs []glob.Glob
}

// regexpFilter represents a number of regular expression filters
type regexpFilter struct {
	re []*regexp.Regexp
}

// lengthFilter represents an URL length filter
type lengthFilter struct {
	limit uint
}

// visitedFilter represents a filter that checks that the URL was visited before
type visitFilter struct {
	maxRevisits uint
	stg         VisitStorage
}

// multiFilter combines multiple filter with AND or OR operator
type multiFilter struct {
	items []FilterEngine
	op    FilterOperator
}

// ------------------------------------------------------------------------

const (
	FILTER_METHOD_INCLUDE FilterMethod = true
	FILTER_METHOD_EXCLUDE FilterMethod = false
)

const (
	FILTER_OPERATOR_AND FilterOperator = iota
	FILTER_OPERATOR_OR
)

const (
	DOMAIN_FILTER FilterScope = iota
	URL_FILTER
)

// ------------------------------------------------------------------------

var (
	ErrFilterNoStorage    = errors.New("no storage was given")                              // ErrFilterNoStorage is thrown when no storage attribute was given.
	ErrFilterItemReplaced = errors.New("a filter item with the same label was overwritten") // ErrFilterItemReplaced when an existing filter item with the same label was overwritten.
	ErrFilterNoEngine     = errors.New("no filter engine was given")                        // ErrFilterNoEngine when an attampt was made to add a new filter item with no filter engine.
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

// Add adds a new filter item to the filter.
func (f *Filter) Add(method FilterMethod, scope FilterScope, engine FilterEngine, label ...string) error {
	f.lock.Lock()
	defer f.lock.Unlock()

	key, err := f.setKey(method, label)

	if method == FILTER_METHOD_INCLUDE {
		f.incl[key] = &filterItem{
			scope:  scope,
			engine: engine,
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

// Has returns true if a filter item exists by method and label.
func (f *Filter) Has(method FilterMethod, label string) bool {
	if method == FILTER_METHOD_INCLUDE {
		_, present := f.incl[label]

		return present
	}

	_, present := f.excl[label]

	return present
}

// ------------------------------------------------------------------------

// Remove removes a filter with a specific method and label.
func (f *Filter) Remove(method FilterMethod, label string) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if method == FILTER_METHOD_INCLUDE {
		delete(f.incl, label)

		return
	}

	delete(f.excl, label)
}

// ------------------------------------------------------------------------

// RemoveAll removes all filters with a specific method and scope.
func (f *Filter) RemoveAll(method FilterMethod, scope FilterScope) {
	f.lock.Lock()
	defer f.lock.Unlock()

	if method == FILTER_METHOD_INCLUDE {
		for key, item := range f.incl {
			if item.scope == scope {
				delete(f.incl, key)
			}
		}

		return
	}

	for key, item := range f.excl {
		if item.scope == scope {
			delete(f.excl, key)
		}
	}
}

// ------------------------------------------------------------------------

// Match reports whether the URL contains any match of the filter.
// Excluding filters will be evaluated before including filters.
// The optional tags will only check filters with matching tag.
func (f *Filter) Match(URL *url.URL, tags ...string) bool {
	segments := map[FilterScope]string{}
	checkTag := len(tags) > 0

	f.lock.RLock()
	defer f.lock.RUnlock()

	// Check the exclusions first
	for key, item := range f.excl {
		if checkTag && !InSlice(key, tags) {
			continue
		}

		if _, present := segments[item.scope]; !present {
			segments[item.scope] = item.segment(URL)
		}

		if item.engine.Match(segments[item.scope]) {
			return false
		}
	}

	for key, item := range f.incl {
		if checkTag && !InSlice(key, tags) {
			continue
		}

		if _, present := segments[item.scope]; !present {
			segments[item.scope] = item.segment(URL)
		}

		if item.engine.Match(segments[item.scope]) {
			return true
		}

	}

	return false
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

	if f.Has(method, key) {
		return key, ErrFilterItemReplaced
	}

	return key, nil
}

// ------------------------------------------------------------------------

// NewGlobFilterItem returns a pointer to a newly created glob pattern filter.
func NewGlobFilterItem(filters []string) (*globFilter, error) {
	f := &globFilter{
		globs: []glob.Glob{},
	}

	errList := []string{}

	// Compile and add the filters
	for _, fltr := range filters {
		if len(fltr) == 0 {
			continue
		}

		glb, err := glob.Compile(fltr)
		if err != nil {
			errList = append(errList, fltr)
			continue
		}

		f.globs = append(f.globs, glb)
	}

	if len(errList) > 0 {
		return f, fmt.Errorf("unable to compile the following filters: %v", "`"+strings.Join(errList, "`, `")+"`")
	}

	return f, nil
}

// Match reports whether the string str contains any match of the filter.
func (f *globFilter) Match(str string) bool {
	for _, glb := range f.globs {
		if glb.Match(str) {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------------------

// NewRegexpFilterItem returns a pointer to a newly created regular expression filter.
func NewRegexpFilterItem(filters []string) (*regexpFilter, error) {
	f := &regexpFilter{
		re: []*regexp.Regexp{},
	}

	errList := []string{}

	// Compile and add the filters
	for _, fltr := range filters {
		if len(fltr) == 0 {
			continue
		}

		re, err := regexp.Compile(fltr)
		if err != nil {
			errList = append(errList, fltr)
			continue
		}

		f.re = append(f.re, re)
	}

	if len(errList) > 0 {
		return f, fmt.Errorf("unable to compile the following filters: %v", "`"+strings.Join(errList, "`, `")+"`")
	}

	return f, nil
}

// Match reports whether the string str contains any match of the filter.
func (f *regexpFilter) Match(str string) bool {
	for _, re := range f.re {
		if re.MatchString(str) {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------------------

// NewLengthFilterItem returns a pointer to a newly created URL Length filter.
func NewLengthFilterItem(maxLength uint) *lengthFilter {
	return &lengthFilter{
		limit: maxLength,
	}
}

// Match reports whether the string str contains any match of the filter.
func (f *lengthFilter) Match(str string) bool {
	return len(str) <= int(f.limit)
}

// ------------------------------------------------------------------------

// NewVisitedFilterItem returns a pointer to a newly created filter that check
// whether or not the URL is eligible for a new visit.
func NewVisitedFilterItem(storage VisitStorage, maxRevisits uint) (*visitFilter, error) {
	if storage == nil {
		return nil, ErrFilterNoStorage
	}

	return &visitFilter{
		maxRevisits: maxRevisits,
		stg:         storage,
	}, nil
}

// Match returns true if the .
func (f *visitFilter) Match(str string) bool {
	visited, err := f.stg.PastVisits(str)

	return err != nil || visited <= f.maxRevisits
}

// ------------------------------------------------------------------------

// NewCombinedFilterItems returns a pointer to a newly created combined filter.
func NewCombinedFilterItems(op FilterOperator, filters ...FilterEngine) (*multiFilter, error) {
	if len(filters) == 0 {
		return nil, ErrFilterNoEngine
	}

	return &multiFilter{
		items: filters,
		op:    op,
	}, nil
}

// Match reports whether the string str contains match of any filter or all filters,
// depending on the logical operator.
func (f *multiFilter) Match(str string) bool {
	switch f.op {
	case FILTER_OPERATOR_AND:
		for _, filter := range f.items {
			if !filter.Match(str) {
				return false
			}
		}

		return true
	case FILTER_OPERATOR_OR:
		for _, filter := range f.items {
			if filter.Match(str) {
				return true
			}
		}

		return false
	}

	return false
}

// ------------------------------------------------------------------------

func (i *filterItem) segment(URL *url.URL) string {
	switch i.scope {
	case DOMAIN_FILTER:
		return URL.Hostname()
	case URL_FILTER:
		return URL.String()
	}

	return ""
}
