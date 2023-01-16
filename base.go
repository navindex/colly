package colly

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kennygrant/sanitize"
)

// ------------------------------------------------------------------------

type RuleEnforcer interface{}

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
	ErrCacheNoStorage      = errors.New("missing cache storage")                    // ErrCacheNoStorage is thrown when an attempt was made to create a Cache without a storage.
	ErrCacheSNoExpHandler  = errors.New("missing cache expiry handler")             // ErrCacheSNoExpHandler is thrown when an attempt was made to create a Cache without an expiry handler.
)

// ------------------------------------------------------------------------

// Error implements error interface.
func (e *MaxVisitReachedError) Error() string {
	return fmt.Sprintf("%q already visited %d times", e.Destination, e.Visits)
}

// ------------------------------------------------------------------------

// SanitizeFileName replaces dangerous characters in a string
// so the return value can be used as a safe file name.
func SanitizeFileName(fileName string) string {
	ext := sanitize.BaseName(filepath.Ext(fileName))
	name := sanitize.BaseName(fileName[:len(fileName)-len(ext)])

	if ext == "" {
		ext = ".unknown"
	}

	return strings.Replace(name+ext, "-", "_", -1)
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
