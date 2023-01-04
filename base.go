package colly

import (
	"context"
	"net/url"
	"time"
)

// ------------------------------------------------------------------------

// Parser is an URL parser.
type Parser interface {
	Parse(rawUrl string) (*url.URL, error)         // Parse parses a raw URL into a URL structure.
	ParseRef(rawUrl, ref string) (*url.URL, error) // ParseRef parses a raw url with a reference into a URL structure.
}

// URLFilter represents an URL filter service.
type URLFilter interface {
	Match(*url.URL) bool // Match reports whether the URL contains any match of the filter.
}

type RuleEnforcer interface{}

// Tracer provides a contract to manage an http trace.
type Tracer interface {
	WithContext(ctx context.Context) context.Context // WithContext returns a new context based on the provided parent context.
}

// Proxy represents a proxy service.
type Proxy interface{}

type Cache interface {
	// CanBeCached(request *http.Request) bool
	Save(key string, value []byte) error
}

// ------------------------------------------------------------------------

// CollectorConfig is a collection of filters and instructions for requests in the collection.
type CollectorConfig struct {
	// MaxDepth limits the recursion depth of visited URLs.
	MaxDepth int `json:"max_depth" bson:"max_depth,omitempty"`
	// MaxBodySize is the limit of the retrieved response body in bytes. 0 means unlimited.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	MaxBodySize int `json:"max_body_size" bson:"max_body_size,omitempty"`
	// URLFilter represents a number of URL filter criteria.
	// Each filter can be an including or excluding filter. Blank filters will be ignored.
	// Excluding filters will be evaluated before including filters.
	URLFilter `json:"url_filter" bson:"url_filter,omitempty"`
	// AllowRevisit, if true, allows multiple downloads of the same URL.
	AllowRevisit bool `json:"revisit" bson:"revisit,omitempty"`
	// IgnoreRobotsTxt, if true, allows the Collector to ignore any restrictions set by the target
	// host's robots.txt file.  See http://www.robotstxt.org/ for more information.
	IgnoreRobotsTxt bool `json:"ignore_robots_txt" bson:"ignore_robots_txt,omitempty"`
	// Async turns on asynchronous network communication. Use Collector.Wait() to
	// be sure all requests have been finished.
	Async bool `json:"async" bson:"async,omitempty"`
	// ParsedStatuses allows parsing HTTP responses by status codes.
	// If blank, the collector will parse only successful HTTP responses.
	ParsedStatuses []int `json:"parsed_statuses" bson:"parsed_statuses,omitempty"`
	// CheckHead performs a HEAD request before every GET to pre-validate the response.
	CheckHead bool `json:"check_head" bson:"check_head,omitempty"`
	// Tracer attaches a tracing service to enable capturing and reporting request performance for crawler tuning.
	Tracer `json:"tracer" bson:"tracer,omitempty"`
	// GroupRules are additional instructions by matching filter criteria.
	DomainRules []DomainRules `json:"domain_rules" bson:"domain_rules,omitempty"`
}

// DomainRules represent request processing instructions by matching domain filter criteria.
type DomainRules struct {
	// URLFilter represents a number of URL filter criteria.
	// Each filter can be an including or excluding filter. Blank filters will be ignored.
	// Excluding filters will be evaluated before including filters.
	URLFilter `json:"url_filter" bson:"url_filter,omitempty"`
	// Delay is the duration to wait before creating a new request.
	Delay time.Duration `json:"delay" bson:"delay,omitempty"`
	// RandomDelay is the extra randomized duration to wait added to Delay before creating a new request.
	RandomDelay time.Duration `json:"random_delay" bson:"random_delay,omitempty"`
	// MaxThreads is the number of the maximum allowed concurrent requests of the matching domains.
	MaxThreads int `json:"max_threads" bson:"max_threads,omitempty"`
}
