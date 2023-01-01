package filter

import (
	"fmt"
	"net/url"
	"regexp"
)

// ------------------------------------------------------------------------

type regexpList []*regexp.Regexp

// regexpFilter represent an enabling/disabling regular expression filter on multiple parts of the URL.
type regexpFilter struct {
	allow bool
	hosts regexpList
	urls  regexpList
}

// ------------------------------------------------------------------------

// NewRegexpFilter returns a pointer to a newly created regular expression filter.
func NewRegexpFilter(allow bool, domainFilters, urlFilters []string) (*regexpFilter, error) {
	f := &regexpFilter{
		allow: allow,
		hosts: regexpList{},
		urls:  regexpList{},
	}

	errList := []string{}

	// Compile and add the domain filters
	if failed := f.hosts.add(domainFilters); len(failed) > 0 {
		errList = append(failed, failed...)
	}

	// Compile and add the URL filters
	if failed := f.urls.add(urlFilters); len(failed) > 0 {
		errList = append(failed, failed...)
	}

	if len(errList) > 0 {
		return f, fmt.Errorf("unable to compile the following filters: %v", errList)
	}

	return f, nil
}

// ------------------------------------------------------------------------

// Included returns true when the URL matches the filter criteria.
func (f *regexpFilter) Included(URL *url.URL) bool {
	return anyRegexpMatch(URL.Hostname(), f.hosts, f.allow) && anyRegexpMatch(URL.String(), f.urls, f.allow)
}

// ------------------------------------------------------------------------

// Excluded returns true when the URL does not match the filter criteria.
func (f *regexpFilter) Excluded(URL *url.URL) bool {
	return !f.Included(URL)
}

// ------------------------------------------------------------------------

func (l regexpList) add(filters []string) []string {
	errList := []string{}

	for _, filter := range filters {
		re, err := regexp.Compile(filter)
		if err != nil {
			errList = append(errList, filter)
			continue
		}
		l = append(l, re)
	}

	return errList
}

// ------------------------------------------------------------------------

func anyRegexpMatch(str string, r []*regexp.Regexp, allow bool) bool {
	if len(r) == 0 {
		return allow
	}

	for _, re := range r {
		if re.MatchString(str) {
			return true
		}
	}

	return false
}
