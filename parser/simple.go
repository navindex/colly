package parser

import (
	"net/url"
)

// ------------------------------------------------------------------------

type simpleParser struct {
	parser func(string) (*url.URL, error)
}

// ------------------------------------------------------------------------

// NewSimpleParser returns a pointer to a newly created simple URL parser.
// NewSimpleParser implements domain.URLParser interface.
func NewSimpleParser() *simpleParser {
	return &simpleParser{
		parser: url.Parse,
	}
}

// ------------------------------------------------------------------------

// Parse parses a raw url into a URL structure.
func (p *simpleParser) Parse(rawURL string) (*url.URL, error) {
	return p.parser(rawURL)
}
