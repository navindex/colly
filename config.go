package colly

import (
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2/filter"
	"github.com/gocolly/colly/v2/tracer"
)

// ------------------------------------------------------------------------

// Environment represents a collection of environment variables.
type Environment interface {
	Values() map[string]string // Values returns the key/value pairs stored in the environment structure.
}

type (
	ConfigSetter        func(c *CollectorConfig)             // ConfigSetter is a function to set a collector configuration option.
	EnvConfigSetter     func(c *CollectorConfig, val string) // EnvConfigSetter is a function to use an environment value to set a collector configuration option.
	ParseStatusCallback func(status int) bool                // ParseStatusCallback is a callback to enable or disable parsing the response, based on the status code.
	UserAgentCallback   func(args ...any) string             // UserAgentCallback is a callback function to return a user agent sting.
	HeaderCallback      func(args ...any) *http.Header       // HeaderCallback is a callback function to return a user agent sting.
)

// CollectorConfig is a collection of filters and instructions for requests in the collection.
type CollectorConfig struct {
	// Filter represents a number of URL filter criteria.
	// Each filter can be an including or excluding filter. Blank filters will be ignored.
	// Excluding filters will be evaluated before including filters.
	Filter *filter.Filter `json:"filter" bson:"filter,omitempty"`
	// MaxDepth limits the recursion depth of visited URLs.
	MaxDepth uint `json:"max_depth" bson:"max_depth,omitempty"`
	// MaxBodySize is the limit of the retrieved response body in bytes. 0 means unlimited.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	MaxBodySize uint `json:"max_body_size" bson:"max_body_size,omitempty"`
	// MaxRevisit, sets how many times the same URL can be visited.
	MaxRevisit uint `json:"max_revisit" bson:"max_revisit,omitempty"`
	// IgnoreRobotsTxt, if true, allows the Collector to ignore any restrictions set by the target
	// host's robots.txt file.  See http://www.robotstxt.org/ for more information.
	IgnoreRobotsTxt bool `json:"ignore_robots_txt" bson:"ignore_robots_txt,omitempty"`
	// Async turns on asynchronous network communication. Use Collector.Wait() to
	// be sure all requests have been finished.
	Async bool `json:"async" bson:"async,omitempty"`
	// DetectCharset enables character encoding detection for non-UTF8 response bodies
	// without explicit charset declaration. This feature uses https://github.com/saintfish/chardet.
	DetectCharset bool `json:"detect_charset" bson:"detect_charset,omitempty"`
	// TODO use this value, if false:
	// c.redirectHandler = func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	}
	FollowRedirects bool `json:"follow_redirects" bson:"follow_redirects,omitempty"`
	// ParseByStatus is a callback function to enable or disable parsing HTTP responses by status codes.
	// If blank, the collector will parse only successful HTTP responses.
	ParseStatusCallback `json:"parse_status_callback" bson:"parse_status_callback,omitempty"`
	// UserAgent is a allback function to create a user agent string.
	UserAgentCallback `json:"user_agent_callback" bson:"user_agent_callback,omitempty"`
	// CheckHead performs a HEAD request before every GET to pre-validate the response.
	CheckHead bool `json:"check_head" bson:"check_head,omitempty"`
	// HeaderCallback is a callback to create common headers for each request.
	HeaderCallback `json:"header_callback" bson:"header_callback,omitempty"`
	// Cache attaches a cache service to keep a local copy of the responses.
	Cache `json:"cache" bson:"cache,omitempty"`
	// TODO create CookieJar interface
	CookieJar `json:"cookie_jar" bson:"cookie_jar,omitempty"`
	// Tracer attaches a tracing service to enable capturing and reporting request performance for crawler tuning.
	Tracer `json:"tracer" bson:"tracer,omitempty"`
	// TODO create logger interface
	Logger `json:"logger" bson:"logger,omitempty"`
	// GroupRules are additional instructions by matching filter criteria.
	DomainRules []DomainRule `json:"domain_rules" bson:"domain_rules,omitempty"`
}

// DomainRules represent request processing instructions by matching domain filter criteria.
type DomainRule struct {
	// Filter represents a number of URL filter criteria.
	// Each filter can be an including or excluding filter. Blank filters will be ignored.
	// Excluding filters will be evaluated before including filters.
	Filter *filter.Filter `json:"filter" bson:"filter,omitempty"`
	// Delay is the duration to wait before creating a new request.
	Delay time.Duration `json:"delay" bson:"delay,omitempty"`
	// RandomDelay is the extra randomized duration to wait added to Delay before creating a new request.
	RandomDelay time.Duration `json:"random_delay" bson:"random_delay,omitempty"`
	// MaxThreads is the number of the maximum allowed concurrent requests of the matching domains.
	MaxThreads int `json:"max_threads" bson:"max_threads,omitempty"`
}

// ------------------------------------------------------------------------

var EnvMap = map[string]EnvConfigSetter{
	"ALLOWED_DOMAINS":    func(c *CollectorConfig, val string) { c.SetAllowedDomains(strings.Split(val, ",")) },
	"DISALLOWED_DOMAINS": func(c *CollectorConfig, val string) { c.SetDisallowedDomains(strings.Split(val, ",")) },
	"USER_AGENT":         func(c *CollectorConfig, val string) { c.UserAgentCallback = func(_ ...any) string { return val } },
	"DETECT_CHARSET": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.Logger.Errorf("DETECT_CHARSET error: %v", err)
		} else {
			c.DetectCharset = b
		}
	},
	"IGNORE_ROBOTSTXT": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.Logger.Errorf("IGNORE_ROBOTSTXT error: %v", err)
		} else {
			c.IgnoreRobotsTxt = b
		}
	},
	"FOLLOW_REDIRECTS": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.Logger.Errorf("FOLLOW_REDIRECTS error: %v", err)
		} else {
			c.FollowRedirects = b
		}
	},
	"CACHE_DIR": func(c *CollectorConfig, val string) {
		// FIXME Create filesystem Cache and set the directory
		// c.CacheDir = val
	},
	"DISABLE_COOKIES": func(c *CollectorConfig, _ string) {
		// TODO Create CookieJar interface first
		// FIXME c.CookieJar == nil
		// c.backend.Client.Jar = nil
	},
	"MAX_BODY_SIZE": func(c *CollectorConfig, val string) {
		if n, err := StrToUInt(val); err != nil {
			c.Logger.Errorf("MAX_BODY_SIZE error: %v", err)
		} else {
			c.MaxBodySize = n
		}
	},
	"MAX_DEPTH": func(c *CollectorConfig, val string) {
		if n, err := StrToUInt(val); err != nil {
			c.Logger.Errorf("MAX_DEPTH error: %v", err)
		} else {
			c.MaxDepth = n
		}
	},
	"MAX_REVISIT": func(c *CollectorConfig, val string) {
		if n, err := StrToUInt(val); err != nil {
			c.Logger.Errorf("MAX_REVISIT error: %v", err)
		} else {
			c.MaxRevisit = n
		}
	},
	"PARSE_HTTP_ERROR_RESPONSE": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.Logger.Errorf("PARSE_HTTP_ERROR_RESPONSE error: %v", err)
		} else {
			fn := parseSuccessResponse
			if b {
				fn = parseErrorResponse
			}
			c.ParseStatusCallback = fn
		}
	},
	"TRACE_HTTP": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.Logger.Errorf("FOLLOW_REDIRECTS error: %v", err)
		} else {
			var t Tracer
			if b {
				t = tracer.NewSimpleTracer()
			}
			c.Tracer = t
		}
	},
}

var (
	parseSuccessResponse = func(code int) bool { return code < 300 }
	parseErrorResponse   = func(code int) bool { return code >= 400 }
	parseAllResponse     = func(code int) bool { return code < 300 || code >= 400 }
)

// ------------------------------------------------------------------------

// NewConfig returns a pointer to a newly created collector configuration settings.
// Same default values are set.
func NewConfig() *CollectorConfig {
	return &CollectorConfig{
		ParseStatusCallback: parseSuccessResponse,
		FollowRedirects:     true,
		// FIXME Cache: ...,
	}
}

// ------------------------------------------------------------------------

// ProcessEnv processes the environment variables by setting the relevant values in CollectorConfig.
func (c *CollectorConfig) ProcessEnv(env Environment, envMap map[string]EnvConfigSetter) {
	if envMap == nil {
		envMap = EnvMap
	}

	for k, v := range env.Values() {
		fn, present := envMap[k]
		if !present {
			c.Logger.Errorf("ProcessEnv: unknown environment variable: %s", k)
			continue
		}

		fn(c, v)
	}
}

// ------------------------------------------------------------------------

// SetAllowedDomains is a convenience method to set the allowed domains.
func (c *CollectorConfig) SetAllowedDomains(domains []string) error {
	f, err := filter.NewGlobFilter(domains)
	if err != nil {
		return err
	}

	if c.Filter == nil {
		c.Filter = filter.New()
	} else {
		c.Filter.Remove(filter.INCLUDE, filter.DOMAIN_FILTER)
	}

	c.Filter.Append(filter.INCLUDE, filter.DOMAIN_FILTER, f)

	return nil
}

// ------------------------------------------------------------------------

// SetDisallowedDomains is a convenience method to set the disallowed domains.
func (c *CollectorConfig) SetDisallowedDomains(domains []string) error {
	f, err := filter.NewGlobFilter(domains)
	if err != nil {
		return err
	}

	if c.Filter == nil {
		c.Filter = filter.New()
	} else {
		c.Filter.Remove(filter.EXCLUDE, filter.DOMAIN_FILTER)
	}

	c.Filter.Append(filter.EXCLUDE, filter.DOMAIN_FILTER, f)

	return nil
}

// ------------------------------------------------------------------------

// ParseSuccessResponse is a convenience method to enable parsing only the HTTP success responses.
func (c *CollectorConfig) ParseSuccessResponses() {
	c.ParseStatusCallback = parseSuccessResponse
}

// ------------------------------------------------------------------------

// ParseErrorResponse is a convenience method to enable parsing only the HTTP error responses.
func (c *CollectorConfig) ParseErrorResponses() {
	c.ParseStatusCallback = parseErrorResponse
}

// ------------------------------------------------------------------------

// ParseAllResponse is a convenience method to enable parsing HTTP success and error responses.
func (c *CollectorConfig) ParseAllResponses() {
	c.ParseStatusCallback = parseAllResponse
}
