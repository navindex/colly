package filter

import (
	"fmt"
	"regexp"
	"strings"
)

// ------------------------------------------------------------------------

// regexpFilter represents a number of regular expression filters
type regexpFilter struct {
	re []*regexp.Regexp
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

// ------------------------------------------------------------------------

// Match reports whether the string str contains any match of the filter.
func (f *regexpFilter) Match(str string) bool {
	for _, re := range f.re {
		if re.MatchString(str) {
			return true
		}
	}

	return false
}
