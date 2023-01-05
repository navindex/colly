package colly

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ------------------------------------------------------------------------

// Parser is an URL parser.
type Parser interface {
	Parse(rawUrl string) (*url.URL, error)         // Parse parses a raw URL into a URL structure.
	ParseRef(rawUrl, ref string) (*url.URL, error) // ParseRef parses a raw url with a reference into a URL structure.
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

// Debugger represnts a debugging backends that processes events.
type Debugger interface {
	Event() // Event receives a new collector event.
}

// Event represents an action inside a collector.
type Event struct {
	Type        string            // Type is the type of the event
	RequestID   uint32            // RequestID identifies the HTTP request of the Event
	CollectorID uint32            // CollectorID identifies the collector of the Event
	Values      map[string]string // Values contains the event's key-value pairs.
}

// Callback functions
type (
	RequestCallback         func(*Request)         // RequestCallback is a type alias for OnRequest callback functions.
	ResponseHeadersCallback func(*Response)        // ResponseHeadersCallback is a type alias for OnResponseHeaders callback functions.
	ResponseCallback        func(*Response)        // ResponseCallback is a type alias for OnResponse callback functions.
	HTMLCallback            func(*HTMLElement)     // HTMLCallback is a type alias for OnHTML callback functions.
	XMLCallback             func(*XMLElement)      // XMLCallback is a type alias for OnXML callback functions.
	ErrorCallback           func(*Response, error) // ErrorCallback is a type alias for OnError callback functions.
	ScrapedCallback         func(*Response)        // ScrapedCallback is a type alias for OnScraped callback functions.
)

// MaxVisitReachedError is the error type for already visited URLs.
// It's returned synchronously by Visit when the URL passed to Visit is already visited.
// When already visited URL is encountered after following
// redirects, this error appears in OnError callback, and if Async
// mode is not enabled, is also returned by Visit.
type MaxVisitReachedError struct {
	// Destination is the URL that was attempted to be visited.
	// It might not match the URL passed to Visit if redirect was followed.
	Destination *url.URL
	Visits      uint
}

// ------------------------------------------------------------------------

// Errors
var (
	ErrForbiddenDomain     = errors.New("forbidden domain")                         // ErrForbiddenDomain is thrown when visiting a domain that is not allowed.
	ErrMissingURL          = errors.New("missing URL")                              // ErrMissingURL is thrown when the URL is missing.
	ErrMaxDepth            = errors.New("max depth limit reached")                  // ErrMaxDepth is thrown for exceeding max depth.
	ErrForbiddenURL        = errors.New("forbidden URL")                            // ErrForbiddenURL is thrown for visiting a URL that is not allowed.
	ErrNoMatchingFilter    = errors.New("no filter match")                          // ErrNoMatchingFilter is thrown when visiting a URL that is not allowed by filters.
	ErrRobotsTxtBlocked    = errors.New("URL blocked by robots.txt")                // ErrRobotsTxtBlocked is thrown for robots.txt errors.
	ErrNoCookieJar         = errors.New("cookie jar not available")                 // ErrNoCookieJar is thrown for missing cookie jar.
	ErrNoFilterDefined     = errors.New("no filter defined in domain rule")         // ErrNoPattern is thrown for DomainRule without valid filters.
	ErrEmptyProxyURL       = errors.New("proxy URL list is empty")                  // ErrEmptyProxyURL is thrown for empty Proxy URL list.
	ErrAbortedAfterHeaders = errors.New("aborted after receiving response headers") // ErrAbortedAfterHeaders is returned when OnResponseHeaders aborts the transfer.
	ErrQueueFull           = errors.New("maximum queue size reached")               // ErrQueueFull is returned when the queue is full.
)

// ------------------------------------------------------------------------

// Error implements error interface.
func (e *MaxVisitReachedError) Error() string {
	return fmt.Sprintf("%q already visited %d times", e.Destination, e.Visits)
}

// ------------------------------------------------------------------------

// StrToUInt converts a string to an unsigned integer.
func StrToUInt(str string) (uint, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("StrToUInt: %w", err)
	}
	if i < 0 {
		return 0, fmt.Errorf("StrToUInt: parsing %q: value must be positive or zero", str)
	}

	return uint(i), nil
}

// ------------------------------------------------------------------------

// StrToBool converts a string to boolean.
func StrToBool(str string) (val bool, err error) {
	switch strings.TrimSpace(strings.ToLower(str)) {
	case "1", "yes", "true", "y":
		val = true
	case "0", "no", "false", "n":
		val = false
	default:
		err = fmt.Errorf("StrToBool: unable to convert %q to boolean", str)
		val = false
	}

	return val, err
}

// ------------------------------------------------------------------------

// IsFalsy returns true if the string represents the boolean value true.
func IsTruthy(str string) bool {
	val, err := StrToBool(str)

	return err == nil && val
}

// ------------------------------------------------------------------------

// IsFalsy returns TRUE if the string represents the boolean value false.
func IsFalsy(str string) bool {
	val, err := StrToBool(str)

	return err == nil && !val
}
