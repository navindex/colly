package colly

import (
	"colly/filters"
	"colly/storage/filesys"
	"colly/storage/mem"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ------------------------------------------------------------------------

type (
	ConfigSetter        func(c *CollectorConfig)             // ConfigSetter is a function to set a collector configuration option.
	EnvConfigSetter     func(c *CollectorConfig, val string) // EnvConfigSetter is a function to use an environment value to set a collector configuration option.
	ParseStatusCallback func(status int) bool                // ParseStatusCallback is a callback to enable or disable parsing the response, based on the status code.
	UserAgentCallback   func() string                        // UserAgentCallback is a callback function to return a user agent string.
	HeaderCallback      func() http.Header                   // HeaderCallback is a callback function to return a list of HTTP headers.
)

// CollectorConfig is a list of collection settings.
type CollectorConfig struct {
	// Filter represents a number of URL filter criteria.
	// Each filter can be an including or excluding filter. Blank filters will be ignored.
	// Excluding filters will be evaluated before including filters.
	*Filter `json:"filter" bson:"filter,omitempty"`
	// MaxDepth limits the recursion depth of visited URLs.
	MaxDepth uint `json:"max_depth" bson:"max_depth,omitempty"`
	// MaxBodySize is the limit of the retrieved response body in bytes. 0 means unlimited.
	// The default value for MaxBodySize is 10MB (10 * 1024 * 1024 bytes).
	MaxBodySize uint `json:"max_body_size" bson:"max_body_size,omitempty"`
	// IgnoreRobotsTxt, if true, allows the Collector to ignore any restrictions set by the target
	// host's robots.txt file.  See http://www.robotstxt.org/ for more information.
	IgnoreRobotsTxt bool `json:"ignore_robots_txt" bson:"ignore_robots_txt,omitempty"`
	// DetectCharset enables character encoding detection for non-UTF8 response bodies
	// without explicit charset declaration. This feature uses https://github.com/saintfish/chardet.
	DetectCharset bool `json:"detect_charset" bson:"detect_charset,omitempty"`
	// TODO use this value, if false:
	// c.redirectHandler = func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	}
	FollowRedirects bool `json:"follow_redirects" bson:"follow_redirects,omitempty"`
	// CheckHead performs a HEAD request before every GET to pre-validate the response.
	CheckHead bool `json:"check_head" bson:"check_head,omitempty"`
	// Async turns on asynchronous network communication. Use Collector.Wait() to
	// be sure all requests have been finished.
	Async bool `json:"async" bson:"async,omitempty"`
	// Delay is the duration to wait before creating a new request.
	// This value is used only if none of filtered configurations is a match.
	Delay time.Duration `json:"delay" bson:"delay,omitempty"`
	// RandomDelay is a randomized duration to be added to Delay before creating a new request.
	// This value is used only if none of filtered configurations is a match.
	RandomDelay time.Duration `json:"random_delay" bson:"random_delay,omitempty"`
	// MaxThreads is the default number of the maximum allowed concurrent requests of the matching domains.
	// This value is used only if none of filtered configurations is a match.
	MaxThreads uint `json:"max_threads" bson:"max_threads,omitempty"`
	// ParseByStatus is a callback function to enable or disable parsing HTTP responses by status codes.
	// If blank, the collector will parse only successful HTTP responses.
	ParseStatusCallback `json:"parse_status_callback" bson:"parse_status_callback,omitempty"`
	// UserAgent is a allback function to create a user agent string.
	UserAgentCallback `json:"user_agent_callback" bson:"user_agent_callback,omitempty"`
	// HeaderCallback is a callback to create common headers for each request.
	HeaderCallback `json:"header_callback" bson:"header_callback,omitempty"`
	// Cache attaches a cache service to keep a local copy of the responses.
	Cache `json:"cache" bson:"cache,omitempty"`
	// TODO create CookieJar interface
	CookieJar http.CookieJar `json:"cookie_jar" bson:"cookie_jar,omitempty"`
	// Parser represents an URL parser service.
	Parser `json:"parser" bson:"parser,omitempty"`
	// Proxy is a represents a web proxy service.
	Proxy `json:"proxy" bson:"proxy,omitempty"`
	// Tracer attaches a tracing service to enable capturing and reporting request performance for crawler tuning.
	Tracer `json:"tracer" bson:"tracer,omitempty"`
	// Logger logs the collector events.
	Logger `json:"logger" bson:"logger,omitempty"`
	// FilteredConfigs is a list of configuration settings that based on URL filter criteria.
	FilteredConfigs []*FilteredConfig `json:"filtered_configs" bson:"filtered_configs,omitempty"`
}

// FilteredConfig represents configuration settings that based on URL filter criteria.
// These settings overwrite similar settings in CollectorConfig if the URL mathces the filter.
type FilteredConfig struct {
	// Filter represents a number of URL filter criteria.
	// Each filter can be an including or excluding filter. Blank filters will be ignored.
	// Excluding filters will be evaluated before including filters.
	*Filter `json:"filter" bson:"filter,omitempty"`
	// Delay is the duration to wait before creating a new request.
	Delay time.Duration `json:"delay" bson:"delay,omitempty"`
	// RandomDelay is the extra randomized duration to wait added to Delay before creating a new request.
	RandomDelay time.Duration `json:"random_delay" bson:"random_delay,omitempty"`
	// MaxThreads is the number of the maximum allowed concurrent requests of the matching domains.
	MaxThreads uint `json:"max_threads" bson:"max_threads,omitempty"`
}

// ------------------------------------------------------------------------

var EnvMap = map[string]EnvConfigSetter{
	"ALLOWED_DOMAINS":    func(c *CollectorConfig, val string) { c.SetAllowedDomains(strings.Split(val, ",")) },
	"DISALLOWED_DOMAINS": func(c *CollectorConfig, val string) { c.SetDisallowedDomains(strings.Split(val, ",")) },
	"USER_AGENT":         func(c *CollectorConfig, val string) { c.UserAgentCallback = func() string { return val } },
	"DETECT_CHARSET": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("DETECT_CHARSET error: %v", err))
		} else {
			c.DetectCharset = b
		}
	},
	"IGNORE_ROBOTSTXT": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("IGNORE_ROBOTSTXT error: %v", err))
		} else {
			c.IgnoreRobotsTxt = b
		}
	},
	"FOLLOW_REDIRECTS": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("FOLLOW_REDIRECTS error: %v", err))
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
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("MAX_BODY_SIZE error: %v", err))
		} else {
			c.MaxBodySize = n
		}
	},
	"MAX_DEPTH": func(c *CollectorConfig, val string) {
		if n, err := StrToUInt(val); err != nil {
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("MAX_DEPTH error: %v", err))
		} else {
			c.MaxDepth = n
		}
	},
	"MAX_REVISIT": func(c *CollectorConfig, val string) {
		if n, err := StrToUInt(val); err != nil {
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("MAX_REVISIT error: %v", err))
		} else {
			c.SetMaxRevisits(n)
		}
	},
	"PARSE_HTTP_ERROR_RESPONSE": func(c *CollectorConfig, val string) {
		if b, err := StrToBool(val); err != nil {
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("PARSE_HTTP_ERROR_RESPONSE error: %v", err))
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
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("FOLLOW_REDIRECTS error: %v", err))
		} else {
			var t Tracer
			if b {
				t = NewSimpleTracer()
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
	jar, _ := NewCookieJar(nil, nil)
	cache, _ := NewCache(mem.NewCacheStorage(), NewCacheExpiryByHeader())

	return &CollectorConfig{
		MaxDepth:            0,
		MaxBodySize:         10 * 1024 * 1024,
		IgnoreRobotsTxt:     true,
		MaxThreads:          1,
		UserAgentCallback:   func() string { return "colly v3" },
		Cache:               cache,
		ParseStatusCallback: parseSuccessResponse,
		FollowRedirects:     true,
		CookieJar:           jar,
		Parser:              NewWHATWGParser(),
	}
}

// ------------------------------------------------------------------------

// NewFilteredConfig returns a pointer to a newly created configuration settings that matches the filter.
func NewFilteredConfig(filter *Filter, delay time.Duration, randomDelay time.Duration, maxThreads uint) (*FilteredConfig, error) {
	if filter == nil {
		return nil, ErrNoFilterDefined
	}

	return &FilteredConfig{
		Filter:      filter,
		Delay:       delay,
		RandomDelay: randomDelay,
		MaxThreads:  maxThreads,
	}, nil
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
			c.logError(LOG_WARN_LEVEL, fmt.Errorf("ProcessEnv: unknown environment variable: %s", k))
			continue
		}

		fn(c, v)
	}
}

// ------------------------------------------------------------------------

// SetAllowedDomains is a convenience method to set the allowed domains.
func (c *CollectorConfig) SetAllowedDomains(domains []string) error {
	if c.Filter == nil {
		c.Filter = NewFilter()
	} else {
		c.Filter.RemoveByScope(DOMAIN_FILTER, FILTER_METHOD_INCLUDE)
	}

	return c.Filter.AddDomainGlob(FILTER_METHOD_INCLUDE, domains, "allowed_domains")
}

// SetDisallowedDomains is a convenience method to set the disallowed domains.
func (c *CollectorConfig) SetDisallowedDomains(domains []string) error {
	if c.Filter == nil {
		c.Filter = NewFilter()
	} else {
		c.Filter.RemoveByScope(DOMAIN_FILTER, FILTER_METHOD_EXCLUDE)
	}

	return c.Filter.AddDomainGlob(FILTER_METHOD_EXCLUDE, domains, "disallowed_domains")
}

// SetUserAgent sets the user agent used by the Collector.
func (c *CollectorConfig) SetUserAgent(ua string) {
	c.UserAgentCallback = func() string {
		return ua
	}
}

// SetCustomHeaders sets the custom headers used by the Collector.
func (c *CollectorConfig) SetCustomHeaders(headers map[string]string) {
	customHdr := http.Header{}
	for header, value := range headers {
		customHdr.Add(header, value)
	}

	c.HeaderCallback = func() http.Header {
		return customHdr
	}
}

// SetTracer sets the request tracer.
// If no attribute given, it will use a simple tracer.
func (c *CollectorConfig) SetTracer(tracer ...Tracer) {
	if len(tracer) > 0 {
		c.Tracer = tracer[0]

		return
	}

	c.Tracer = NewSimpleTracer()
}

// SetLogger sets the logger.
// If no attribute given, it will use a standard logger.
func (c *CollectorConfig) SetLogger(logger ...Logger) {
	if len(logger) > 0 {
		c.Logger = logger[0]

		return
	}

	c.Logger = NewStdLogger(os.Stderr, "", log.LstdFlags)
}

// SetCache sets the request cache.
// If no storage attribute given, it will use an in-memory cache.
func (c *CollectorConfig) SetCache(storage CacheStorage, expHandler CacheExpiryHandler) error {
	if storage == nil {
		return ErrCacheNoStorage
	}

	if expHandler == nil {
		return ErrCacheNoExpHandler
	}

	cache, err := NewCache(storage, expHandler)
	if err != nil {
		return err
	}
	c.Cache = cache

	return nil
}

// SetCache sets the request cache.
// If no expiry handler given, it will use the response cache headers.
func (c *CollectorConfig) SetFileCache(path string, expHandler CacheExpiryHandler) error {
	if path == "" {
		return ErrCacheNoPath
	}

	storage, err := filesys.NewCacheStorage(path)
	if err != nil {
		return err
	}

	if expHandler == nil {
		expHandler = NewCacheExpiryByHeader()
	}

	cache, err := NewCache(storage, expHandler)
	if err != nil {
		return err
	}
	c.Cache = cache

	return nil
}

// SetMaxRevisits sets how many times the same URL can be visited.
// The storage attribute, if not nil, will be used to store the number of visits.
// If no storage is given, the visits will be used in the memory.
func (c *CollectorConfig) SetMaxRevisits(maxRevisits uint, storage ...filters.VisitStorage) error {
	const label = "revisit"
	var stg filters.VisitStorage

	if len(storage) > 0 {
		stg = storage[0]
	}

	if c.Filter == nil {
		c.Filter = NewFilter()
	}

	return c.Filter.AddRevisit(maxRevisits, stg, "revisit")
}

// ------------------------------------------------------------------------

// ParseSuccessResponse is a convenience method to enable parsing only the HTTP success responses.
func (c *CollectorConfig) ParseSuccessResponses() {
	c.ParseStatusCallback = parseSuccessResponse
}

// ParseErrorResponse is a convenience method to enable parsing only the HTTP error responses.
func (c *CollectorConfig) ParseErrorResponses() {
	c.ParseStatusCallback = parseErrorResponse
}

// ParseAllResponse is a convenience method to enable parsing HTTP success and error responses.
func (c *CollectorConfig) ParseAllResponses() {
	c.ParseStatusCallback = parseAllResponse
}

// ------------------------------------------------------------------------

func (c *CollectorConfig) hasLogger() bool {
	return c.Logger != nil
}

func (c *CollectorConfig) logError(level LogLevel, err error) {
	if c.hasLogger() {
		c.Logger.LogError(level, err)
	}
}
