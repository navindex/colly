package filter

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

// ------------------------------------------------------------------------

// globFilter represents a number of glob expression filters
type globFilter struct {
	globs []glob.Glob
}

// ------------------------------------------------------------------------

// NewGlobFilter returns a pointer to a newly created glob pattern filter.
func NewGlobFilter(filters []string) (*globFilter, error) {
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

// ------------------------------------------------------------------------

// Match reports whether the string str contains any match of the filter.
func (f *globFilter) Match(str string) bool {
	for _, glb := range f.globs {
		if glb.Match(str) {
			return true
		}
	}

	return false
}
