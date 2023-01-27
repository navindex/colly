package filters

// ------------------------------------------------------------------------

// urlLengthFilter represents an URL length filter
type urlLengthFilter struct {
	min uint
	max uint
}

// ------------------------------------------------------------------------

// NewURLLengthEngine returns a pointer to a newly created URL length filter.
// This filter should be used with FILTER_METHOD_EXCLUDE method.
func NewURLLengthEngine(minLength uint, maxLength uint) *urlLengthFilter {
	return &urlLengthFilter{
		min: minLength,
		max: maxLength,
	}
}

// ------------------------------------------------------------------------

// Match reports whether the string str contains any match of the filter.
func (f *urlLengthFilter) Match(u any) bool {
	str, ok := u.(string)
	if !ok {
		return false
	}

	len := len(str)

	return len < int(f.min) || len > int(f.max)
}
