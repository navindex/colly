package parser

import (
	"net/url"

	whatwg "github.com/nlnwa/whatwg-url/url"
)

// ------------------------------------------------------------------------

type whatwgParser struct {
	parser whatwg.Parser
}

// ------------------------------------------------------------------------

// NewWHATWGParser returns a pointer to a newly created WHATWG URL parser.
// NewWHATWGParser implements domain.URLParser interface.
func NewWHATWGParser() *whatwgParser {
	return &whatwgParser{
		parser: whatwg.NewParser(whatwg.WithPercentEncodeSinglePercentSign()),
	}
}

// ------------------------------------------------------------------------

// Parse parses a raw url into a URL structure.
func (p *whatwgParser) Parse(rawURL string) (*url.URL, error) {
	wurl, err := p.parser.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	return url.Parse(wurl.Href(false))
}

// ------------------------------------------------------------------------

// ParseRef parses a raw url with a reference into a URL structure.
func (p *whatwgParser) ParseRef(rawURL string, ref string) (*url.URL, error) {
	wurl, err := p.parser.ParseRef(rawURL, ref)
	if err != nil {
		return nil, err
	}

	return url.Parse(wurl.Href(false))
}
