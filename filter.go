package colly

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

// ------------------------------------------------------------------------

// Filter represents a number of including/excluding filters.
type Filter struct {
	incl []*filterItem
	excl []*filterItem
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
	maxVisits uint
	stg       VisitStorage
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
	ErrFilterNoStorage    = errors.New("no storage was given")
	ErrFilterNoEngine     = errors.New("no filter engine was given")
	ErrFilterZeroMaxVisit = errors.New("maximum number of visits is zero, must be positive")
)

// ------------------------------------------------------------------------

// NewFilter returns a pointer to a newly created filter.
func NewFilter() *Filter {
	return &Filter{
		incl: make([]*filterItem, 0),
		excl: make([]*filterItem, 0),
	}
}

// Append appends a new filter to the filter list.
func (f *Filter) Append(method FilterMethod, scope FilterScope, engine FilterEngine) {
	if method == FILTER_METHOD_INCLUDE {
		f.incl = append(f.incl, &filterItem{
			scope:  scope,
			engine: engine,
		})

		return
	}

	f.excl = append(f.excl, &filterItem{
		scope:  scope,
		engine: engine,
	})
}

// Remove removes all filters with a specific method and scope.
func (f *Filter) Remove(method FilterMethod, scope FilterScope) {
	var items []*filterItem

	if method == FILTER_METHOD_INCLUDE {
		items = f.incl
	} else {
		items = f.excl
	}

	newItems := []*filterItem{}
	for _, item := range items {
		if item.scope != scope {
			newItems = append(newItems, item)
		}
	}

	if method == FILTER_METHOD_INCLUDE {
		f.incl = newItems
	} else {
		f.excl = newItems
	}
}

// Match reports whether the URL contains any match of the filter.
// Excluding filters will be evaluated before including filters.
func (f *Filter) Match(URL *url.URL) bool {
	segments := map[FilterScope]string{}

	// Check the exclusions first
	for _, item := range f.excl {
		if _, present := segments[item.scope]; !present {
			segments[item.scope] = item.segment(URL)
		}

		if item.engine.Match(segments[item.scope]) {
			return false
		}
	}

	for _, item := range f.incl {
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
// whether or not the URL was visited. It must be used with REQUEST_ID_FILTER.
func NewVisitedFilterItem(storage VisitStorage, maxVisits uint) (*visitFilter, error) {
	if storage == nil {
		return nil, ErrFilterNoStorage
	}

	if maxVisits == 0 {
		return nil, ErrFilterZeroMaxVisit
	}

	return &visitFilter{
		maxVisits: maxVisits,
		stg:       storage,
	}, nil
}

// Match reports whether the string str contains any match of the filter.
func (f *visitFilter) Match(str string) bool {
	visited, err := f.stg.PastVisits(str)

	return err != nil || visited < f.maxVisits
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
