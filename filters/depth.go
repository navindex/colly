package filters

// ------------------------------------------------------------------------

// depthFilter represents a request depth filter
type depthFilter struct {
	limit uint
}

// ------------------------------------------------------------------------

// NewRequestDepthEngine returns a pointer to a newly created reqiest depth filter.
// This filter should be used with FILTER_METHOD_EXCLUDE method.
func NewRequestDepthEngine(maxDepth uint) *depthFilter {
	return &depthFilter{
		limit: maxDepth,
	}
}

// ------------------------------------------------------------------------

// Match reports whether the string str contains any match of the filter.
func (f *depthFilter) Match(d any) bool {
	depth, ok := d.(uint16)
	if !ok {
		return false
	}

	return uint(depth) > f.limit
}
