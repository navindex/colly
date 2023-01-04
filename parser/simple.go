package parser

import (
	"net/url"
)

// ------------------------------------------------------------------------

type simpleParser struct{}

// ------------------------------------------------------------------------

// NewSimpleParser returns a pointer to a newly created simple URL parser.
// NewSimpleParser implements domain.URLParser interface.
func NewSimpleParser() *simpleParser {
	return &simpleParser{}
}

// ------------------------------------------------------------------------

// Parse parses a raw url into a URL structure.
func (p *simpleParser) Parse(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// ------------------------------------------------------------------------

// ParseRef parses a raw url with a reference into a URL structure.
func (p *simpleParser) ParseRef(rawURL string, ref string) (*url.URL, error) {
	u, err := p.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return u.Parse(ref)
}
