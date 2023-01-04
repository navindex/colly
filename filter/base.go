package filter

import (
	"net/url"
)

// ------------------------------------------------------------------------

// Filter represents a number of including/excluding filters.
type Filter struct {
	incl []*filterItem
	excl []*filterItem
}

// Engine privides the function to match the filter.
type Engine interface {
	Match(string) bool
}

// Method tells whether the filter supposed to include or exclude matches.
type Method bool

// Scope points out which part of the URL will be matched.
type Scope uint8

// filterItem represent an including/excluding URL filter
type filterItem struct {
	scope  Scope
	engine Engine
}

// ------------------------------------------------------------------------

const (
	INCLUDE Method = true
	EXCLUDE Method = false
)

const (
	DOMAIN_FILTER Scope = iota
	URL_FILTER
)

// ------------------------------------------------------------------------

// New returns a pointer to a newly created filter.
func New() *Filter {
	return &Filter{
		incl: make([]*filterItem, 0),
		excl: make([]*filterItem, 0),
	}
}

// ------------------------------------------------------------------------

// Append appends a new filter to the filter list.
func (f *Filter) Append(method Method, scope Scope, engine Engine) {
	if method == INCLUDE {
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

// ------------------------------------------------------------------------

// Match reports whether the URL contains any match of the filter.
func (f *Filter) Match(URL *url.URL) bool {
	segments := map[Scope]string{}

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

func (i *filterItem) segment(URL *url.URL) string {
	switch i.scope {
	case DOMAIN_FILTER:
		return URL.Hostname()
	case URL_FILTER:
		return URL.String()
	}

	return ""
}
