package colly

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	ErrNoHTTPRequest       = errors.New("HTTP Request reference is nil")            // ErrNoHTTPRequest is thrown when the HTTP request pointer is set to nil.
	ErrNoCollector         = errors.New("Collector reference is nil")               // ErrNoCollector is thrown when the Collector pointer is set to nil.
	ErrMaxDepth            = errors.New("max depth limit reached")                  // ErrMaxDepth is thrown for exceeding max depth.
	ErrRobotsTxtBlocked    = errors.New("URL blocked by robots.txt")                // ErrRobotsTxtBlocked is thrown for robots.txt errors.
	ErrNoCookieJar         = errors.New("cookie jar not available")                 // ErrNoCookieJar is thrown for missing cookie jar.
	ErrNoFilterDefined     = errors.New("no filter defined")                        // ErrNoFilterDefined is thrown when no valid filter was provided.
	ErrEmptyProxyURL       = errors.New("proxy URL list is empty")                  // ErrEmptyProxyURL is thrown for empty Proxy URL list.
	ErrAbortedAfterHeaders = errors.New("aborted after receiving response headers") // ErrAbortedAfterHeaders is returned when OnResponseHeaders aborts the transfer.
	ErrQueueFull           = errors.New("maximum queue size reached")               // ErrQueueFull is returned when the queue is full.
	ErrCacheNoStorage      = errors.New("missing cache storage")                    // ErrCacheNoStorage is thrown when an attempt was made to create a Cache without a storage.
	ErrCacheNoPath         = errors.New("file cache path is blank")                 // ErrCacheNoPath is thrown when an attempt was made to create a file cache with a blank path.
	ErrCacheNoExpHandler   = errors.New("missing cache expiry handler")             // ErrCacheNoExpHandler is thrown when an attempt was made to create a Cache without an expiry handler.
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

// ------------------------------------------------------------------------

// InSlice checks if a slice contains a given value.
func InSlice[E comparable](needle E, haystack []E) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, e := range haystack {
		if needle == e {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------------------

// ContainsAny reports whether any of substr is within s.
func ContainsAny(s string, substr ...string) bool {
	if len(substr) == 0 {
		return false
	}

	for _, sub := range substr {
		if strings.Contains(s, sub) {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------------------

func RandomString(len uint) string {
	buf := make([]byte, int(len))

	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", buf[:])
}

// ------------------------------------------------------------------------

// newFormReader returns a form data reader
func NewFormReader(data map[string]string) io.Reader {
	form := url.Values{}

	for k, v := range data {
		form.Add(k, v)
	}

	return strings.NewReader(form.Encode())
}

// ------------------------------------------------------------------------

// newFormReader returns a new multipart data reader
func NewMultipartReader(boundary string, data map[string][]byte) io.Reader {
	dashBoundary := "--" + boundary

	buffer := bytes.NewBuffer([]byte{})

	buffer.WriteString(fmt.Sprintf("Content-type: multipart/form-data; boundary=%s\n\n", boundary))

	for contentType, content := range data {
		buffer.WriteString(dashBoundary + "\n" +
			fmt.Sprintf("Content-Disposition: form-data; name=%s\n", contentType) +
			fmt.Sprintf("Content-Length: %d \n\n", len(content)))
		buffer.Write(content)
		buffer.WriteString("\n")
	}

	buffer.WriteString(dashBoundary + "--\n\n")

	return buffer
}

// ------------------------------------------------------------------------

// MergeHeaders merges multiple HTTP headers.
func MergeHeaders(headers ...http.Header) http.Header {
	hdr := http.Header{}
	if len(headers) >= 1 {
		return headers[0]
	}
	if len(headers) <= 1 {
		return hdr
	}

	for i := range headers[1:] {
		for k, v := range headers[i] {
			for _, value := range v {
				hdr.Add(k, value)
			}
		}

	}

	return hdr
}

// ------------------------------------------------------------------------

// IsXML returns true if the path extention indicates an XML file.
func IsXML(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".xml") || strings.HasSuffix(strings.ToLower(path), ".xml.gz")
}
