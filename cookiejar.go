package colly

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

// ------------------------------------------------------------------------

// CookieStorage manages a storage that saves, deletes and retrieves cookies.
type CookieStorage interface {
	Set(key string, entries io.Reader) error // Set sets the entries in binary format.
	Get(key string) (io.Reader, error)       // Get retrieves the entries in binary format.
	Remove(key string) error                 // Remove removes an entry by key.
	Clear() error                            // Clear deletes all stored items.
}

// cookieJar implements the http.CookieJar interface from the net/http package.
type cookieJar struct {
	psList cookiejar.PublicSuffixList
	lock   *sync.Mutex

	// storage saves the set of entries, keyed by their eTLD+1 and subkeyed by
	// their name/domain/path.
	storage CookieStorage

	// nextSeqNum is the next sequence number assigned to a new cookie
	// created SetCookies.
	nextSeqNum uint64
}

// entry is the internal representation of a cookie.
// This struct type is not used outside of this package per se, but the exported
// fields are those of RFC 6265.
type entry struct {
	Name       string    `json:"name" bson:"name,omitempty"`
	Value      string    `json:"value" bson:"value,omitempty"`
	Domain     string    `json:"domain" bson:"domain,omitempty"`
	Path       string    `json:"path" bson:"path,omitempty"`
	SameSite   string    `json:"same_site" bson:"same_site,omitempty"`
	Secure     bool      `json:"secure" bson:"secure,omitempty"`
	HttpOnly   bool      `json:"http_only" bson:"http_only,omitempty"`
	Persistent bool      `json:"persistent" bson:"persistent,omitempty"`
	HostOnly   bool      `json:"host_only" bson:"host_only,omitempty"`
	Expires    time.Time `json:"expires" bson:"expires,omitempty"`
	Creation   time.Time `json:"creation" bson:"creation,omitempty"`
	LastAccess time.Time `json:"last_access" bson:"last_access,omitempty"`

	// seqNum is a sequence number so that Cookies returns cookies in a
	// deterministic order, even for cookies that have equal Path length and
	// equal Creation time. This simplifies testing.
	seqNum uint64 `json:"seq_num" bson:"seq_num,omitempty"`
}

// entries is the internal representation of a submap.
type entries map[string]entry

// ------------------------------------------------------------------------

// These parameter values are specified in section 5.
// All computation is done with int32s, so that overflow behavior is identical
// regardless of whether int is 32-bit or 64-bit.
const (
	base        int32 = 36
	damp        int32 = 700
	initialBias int32 = 72
	initialN    int32 = 128
	skew        int32 = 38
	tmax        int32 = 26
	tmin        int32 = 1
)

// ------------------------------------------------------------------------

var (
	errIllegalDomain   = errors.New("cookiejar: illegal cookie domain attribute")
	errMalformedDomain = errors.New("cookiejar: malformed cookie domain attribute")
)

// endOfTime is the time when session (non-persistent) cookies expire.
// This instant is representable in most date/time formats (not just
// Go's time.Time) and should be far enough in the future.
var endOfTime = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

// ------------------------------------------------------------------------

// NewCookieJar returns a pointer to a newly created cookie jar.
// A nil *Options is equivalent to a zero Options.
// If no storage was given, an in-memory cookie jar will be returned.
func NewCookieJar(storage CookieStorage, o *cookiejar.Options) (http.CookieJar, error) {
	if storage == nil {
		return cookiejar.New(o)
	}

	jar := &cookieJar{
		storage: storage,
	}

	if o != nil {
		jar.psList = o.PublicSuffixList
	}

	return jar, nil
}

// ------------------------------------------------------------------------

// DecodeBinaryToEntries encodes the entry submap to bytes.
func DecodeBinaryToEntries(data io.Reader) (entries, error) {
	// Decode to a slice of cookies
	var e entries
	err := gob.NewDecoder(data).Decode(&e)

	return e, err
}

// ------------------------------------------------------------------------

// Cookies implements the Cookies method of the http.CookieJar interface.
// It returns an empty slice if the URL's scheme is not HTTP or HTTPS.
func (j *cookieJar) Cookies(u *url.URL) (cookies []*http.Cookie) {
	return j.cookies(u, time.Now())
}

// ------------------------------------------------------------------------

// SetCookies implements the SetCookies method of the http.CookieJar interface.
//
// It does nothing if the URL's scheme is not HTTP or HTTPS.
func (j *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.setCookies(u, cookies, time.Now())
}

// ------------------------------------------------------------------------

// cookies is like Cookies but takes the current time as a parameter.
func (j *cookieJar) cookies(u *url.URL, now time.Time) (cookies []*http.Cookie) {
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil
	}

	host, err := canonicalHost(u.Host)
	if err != nil {
		return nil
	}
	key := jarKey(host, j.psList)

	j.lock.Lock()
	defer j.lock.Unlock()

	b, err := j.storage.Get(key)
	if err != nil {
		return nil
	}
	submap, err := DecodeBinaryToEntries(b)
	if err != nil || submap == nil {
		return nil
	}

	https := u.Scheme == "https"
	path := u.Path
	if path == "" {
		path = "/"
	}

	modified := false
	var selected []entry
	for id, e := range submap {
		if e.Persistent && !e.Expires.After(now) {
			delete(submap, id)
			modified = true
			continue
		}

		if !e.shouldSend(https, host, path) {
			continue
		}

		e.LastAccess = now
		submap[id] = e
		selected = append(selected, e)
		modified = true
	}

	if modified {
		if len(submap) == 0 {
			j.storage.Remove(key)
		} else {
			if data, err := submap.BinaryEncode(); err == nil {
				j.storage.Set(key, data)
			}
		}
	}

	// sort according to RFC 6265 section 5.4 point 2: by longest
	// path and then by earliest creation time.
	sort.Slice(selected, func(i, j int) bool {
		s := selected
		if len(s[i].Path) != len(s[j].Path) {
			return len(s[i].Path) > len(s[j].Path)
		}
		if !s[i].Creation.Equal(s[j].Creation) {
			return s[i].Creation.Before(s[j].Creation)
		}
		return s[i].seqNum < s[j].seqNum
	})
	for _, e := range selected {
		cookies = append(cookies, &http.Cookie{Name: e.Name, Value: e.Value})
	}

	return cookies
}

// ------------------------------------------------------------------------

// setCookies is like SetCookies but takes the current time as parameter.
func (j *cookieJar) setCookies(u *url.URL, cookies []*http.Cookie, now time.Time) {
	if len(cookies) == 0 {
		return
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return
	}

	host, err := canonicalHost(u.Host)
	if err != nil {
		return
	}
	key := jarKey(host, j.psList)
	defPath := defaultPath(u.Path)

	j.lock.Lock()
	defer j.lock.Unlock()

	b, err := j.storage.Get(key)
	if err != nil {
		return
	}
	submap, err := DecodeBinaryToEntries(b)
	if err != nil || submap == nil {
		return
	}

	modified := false
	for _, cookie := range cookies {
		e, remove, err := j.newEntry(cookie, now, defPath, host)
		if err != nil {
			continue
		}
		id := e.id()

		if remove {
			if submap != nil {
				if _, ok := submap[id]; ok {
					delete(submap, id)
					modified = true
				}
			}
			continue
		}

		if submap == nil {
			submap = entries{}
		}

		if old, ok := submap[id]; ok {
			e.Creation = old.Creation
			e.seqNum = old.seqNum
		} else {
			e.Creation = now
			e.seqNum = j.nextSeqNum
			j.nextSeqNum++

		}

		e.LastAccess = now
		submap[id] = e
		modified = true
	}

	if modified {
		if len(submap) == 0 {
			j.storage.Remove(key)
		} else {
			if data, err := submap.BinaryEncode(); err == nil {
				j.storage.Set(key, data)
			}
		}
	}
}

// ------------------------------------------------------------------------

// newEntry creates an entry from a http.Cookie c. now is the current time and
// is compared to c.Expires to determine deletion of c. defPath and host are the
// default-path and the canonical host name of the URL c was received from.
//
// remove records whether the jar should delete this cookie, as it has already
// expired with respect to now. In this case, e may be incomplete, but it will
// be valid to call e.id (which depends on e's Name, Domain and Path).
//
// A malformed c.Domain will result in an error.
func (j *cookieJar) newEntry(c *http.Cookie, now time.Time, defPath, host string) (e entry, remove bool, err error) {
	e.Name = c.Name

	if c.Path == "" || c.Path[0] != '/' {
		e.Path = defPath
	} else {
		e.Path = c.Path
	}

	e.Domain, e.HostOnly, err = j.domainAndType(host, c.Domain)
	if err != nil {
		return e, false, err
	}

	// MaxAge takes precedence over Expires.
	if c.MaxAge < 0 {
		return e, true, nil
	} else if c.MaxAge > 0 {
		e.Expires = now.Add(time.Duration(c.MaxAge) * time.Second)
		e.Persistent = true
	} else {
		if c.Expires.IsZero() {
			e.Expires = endOfTime
			e.Persistent = false
		} else {
			if !c.Expires.After(now) {
				return e, true, nil
			}
			e.Expires = c.Expires
			e.Persistent = true
		}
	}

	e.Value = c.Value
	e.Secure = c.Secure
	e.HttpOnly = c.HttpOnly

	switch c.SameSite {
	case http.SameSiteDefaultMode:
		e.SameSite = "SameSite"
	case http.SameSiteStrictMode:
		e.SameSite = "SameSite=Strict"
	case http.SameSiteLaxMode:
		e.SameSite = "SameSite=Lax"
	}

	return e, false, nil
}

// ------------------------------------------------------------------------

// domainAndType determines the cookie's domain and hostOnly attribute.
func (j *cookieJar) domainAndType(host, domain string) (string, bool, error) {
	if domain == "" {
		// No domain attribute in the SetCookie header indicates a
		// host cookie.
		return host, true, nil
	}

	if isIP(host) {
		// RFC 6265 is not super clear here, a sensible interpretation
		// is that cookies with an IP address in the domain-attribute
		// are allowed.

		// RFC 6265 section 5.2.3 mandates to strip an optional leading
		// dot in the domain-attribute before processing the cookie.
		//
		// Most browsers don't do that for IP addresses, only curl
		// version 7.54) and and IE (version 11) do not reject a
		//     Set-Cookie: a=1; domain=.127.0.0.1
		// This leading dot is optional and serves only as hint for
		// humans to indicate that a cookie with "domain=.bbc.co.uk"
		// would be sent to every subdomain of bbc.co.uk.
		// It just doesn't make sense on IP addresses.
		// The other processing and validation steps in RFC 6265 just
		// collaps to:
		if host != domain {
			return "", false, errIllegalDomain
		}

		// According to RFC 6265 such cookies should be treated as
		// domain cookies.
		// As there are no subdomains of an IP address the treatment
		// according to RFC 6265 would be exactly the same as that of
		// a host-only cookie. Contemporary browsers (and curl) do
		// allows such cookies but treat them as host-only cookies.
		// So do we as it just doesn't make sense to label them as
		// domain cookies when there is no domain; the whole notion of
		// domain cookies requires a domain name to be well defined.
		return host, true, nil
	}

	// From here on: If the cookie is valid, it is a domain cookie (with
	// the one exception of a public suffix below).
	// See RFC 6265 section 5.2.3.
	if domain[0] == '.' {
		domain = domain[1:]
	}

	if len(domain) == 0 || domain[0] == '.' {
		// Received either "Domain=." or "Domain=..some.thing",
		// both are illegal.
		return "", false, errMalformedDomain
	}

	domain, isASCII := toLower(domain)
	if !isASCII {
		// Received non-ASCII domain, e.g. "perché.com" instead of "xn--perch-fsa.com"
		return "", false, errMalformedDomain
	}

	if domain[len(domain)-1] == '.' {
		// We received stuff like "Domain=www.example.com.".
		// Browsers do handle such stuff (actually differently) but
		// RFC 6265 seems to be clear here (e.g. section 4.1.2.3) in
		// requiring a reject.  4.1.2.3 is not normative, but
		// "Domain Matching" (5.1.3) and "Canonicalized Host Names"
		// (5.1.2) are.
		return "", false, errMalformedDomain
	}

	// See RFC 6265 section 5.3 #5.
	if j.psList != nil {
		if ps := j.psList.PublicSuffix(domain); ps != "" && !hasDotSuffix(domain, ps) {
			if host == domain {
				// This is the one exception in which a cookie
				// with a domain attribute is a host cookie.
				return host, true, nil
			}
			return "", false, errIllegalDomain
		}
	}

	// The domain must domain-match host: www.mycompany.com cannot
	// set cookies for .ourcompetitors.com.
	if host != domain && !hasDotSuffix(host, domain) {
		return "", false, errIllegalDomain
	}

	return domain, false, nil
}

// ------------------------------------------------------------------------

// BinaryEncode encodes the entry submap to bytes.
func (e entries) BinaryEncode() (io.Reader, error) {
	data := &bytes.Buffer{}
	err := gob.NewEncoder(data).Encode(e)

	return data, err
}

// ------------------------------------------------------------------------

// id returns the domain;path;name triple of e as an id.
func (e *entry) id() string {
	return fmt.Sprintf("%s;%s;%s", e.Domain, e.Path, e.Name)
}

// ------------------------------------------------------------------------

// shouldSend determines whether e's cookie qualifies to be included in a
// request to host/path. It is the caller's responsibility to check if the
// cookie is expired.
func (e *entry) shouldSend(https bool, host, path string) bool {
	return e.domainMatch(host) && e.pathMatch(path) && (https || !e.Secure)
}

// ------------------------------------------------------------------------

// domainMatch checks whether e's Domain allows sending e back to host.
// It differs from "domain-match" of RFC 6265 section 5.1.3 because we treat
// a cookie with an IP address in the Domain always as a host cookie.
func (e *entry) domainMatch(host string) bool {
	if e.Domain == host {
		return true
	}
	return !e.HostOnly && hasDotSuffix(host, e.Domain)
}

// ------------------------------------------------------------------------

// pathMatch implements "path-match" according to RFC 6265 section 5.1.4.
func (e *entry) pathMatch(requestPath string) bool {
	if requestPath == e.Path {
		return true
	}
	if strings.HasPrefix(requestPath, e.Path) {
		if e.Path[len(e.Path)-1] == '/' {
			return true // The "/any/" matches "/any/path" case.
		} else if requestPath[len(e.Path)] == '/' {
			return true // The "/any" matches "/any/path" case.
		}
	}
	return false
}

// ------------------------------------------------------------------------

// hasDotSuffix reports whether s ends in "."+suffix.
func hasDotSuffix(s, suffix string) bool {
	return len(s) > len(suffix) && s[len(s)-len(suffix)-1] == '.' && s[len(s)-len(suffix):] == suffix
}

// canonicalHost strips port from host if present and returns the canonicalized
// host name.
func canonicalHost(host string) (string, error) {
	var err error
	if hasPort(host) {
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			return "", err
		}
	}
	// Strip trailing dot from fully qualified domain names.
	host = strings.TrimSuffix(host, ".")
	encoded, err := toASCII(host)
	if err != nil {
		return "", err
	}
	// We know this is ascii, no need to check.
	lower, _ := toLower(encoded)
	return lower, nil
}

// hasPort reports whether host contains a port number. host may be a host
// name, an IPv4 or an IPv6 address.
func hasPort(host string) bool {
	colons := strings.Count(host, ":")
	if colons == 0 {
		return false
	}
	if colons == 1 {
		return true
	}
	return host[0] == '[' && strings.Contains(host, "]:")
}

// jarKey returns the key to use for a jar.
func jarKey(host string, psl cookiejar.PublicSuffixList) string {
	if isIP(host) {
		return host
	}

	var i int
	if psl == nil {
		i = strings.LastIndex(host, ".")
		if i <= 0 {
			return host
		}
	} else {
		suffix := psl.PublicSuffix(host)
		if suffix == host {
			return host
		}
		i = len(host) - len(suffix)
		if i <= 0 || host[i-1] != '.' {
			// The provided public suffix list psl is broken.
			// Storing cookies under host is a safe stopgap.
			return host
		}
		// Only len(suffix) is used to determine the jar key from
		// here on, so it is okay if psl.PublicSuffix("www.buggy.psl")
		// returns "com" as the jar key is generated from host.
	}
	prevDot := strings.LastIndex(host[:i-1], ".")
	return host[prevDot+1:]
}

// isIP reports whether host is an IP address.
func isIP(host string) bool {
	return net.ParseIP(host) != nil
}

// defaultPath returns the directory part of an URL's path according to
// RFC 6265 section 5.1.4.
func defaultPath(path string) string {
	if len(path) == 0 || path[0] != '/' {
		return "/" // Path is empty or malformed.
	}

	i := strings.LastIndex(path, "/") // Path starts with "/", so i != -1.
	if i == 0 {
		return "/" // Path has the form "/abc".
	}
	return path[:i] // Path is either of form "/abc/xyz" or "/abc/xyz/".
}

// encode encodes a string as specified in section 6.3 and prepends prefix to
// the result.
//
// The "while h < length(input)" line in the specification becomes "for
// remaining != 0" in the Go code, because len(s) in Go is in bytes, not runes.
func encode(prefix, s string) (string, error) {
	output := make([]byte, len(prefix), len(prefix)+1+2*len(s))
	copy(output, prefix)
	delta, n, bias := int32(0), initialN, initialBias
	b, remaining := int32(0), int32(0)
	for _, r := range s {
		if r < utf8.RuneSelf {
			b++
			output = append(output, byte(r))
		} else {
			remaining++
		}
	}
	h := b
	if b > 0 {
		output = append(output, '-')
	}
	for remaining != 0 {
		m := int32(0x7fffffff)
		for _, r := range s {
			if m > r && r >= n {
				m = r
			}
		}
		delta += (m - n) * (h + 1)
		if delta < 0 {
			return "", fmt.Errorf("cookiejar: invalid label %q", s)
		}
		n = m
		for _, r := range s {
			if r < n {
				delta++
				if delta < 0 {
					return "", fmt.Errorf("cookiejar: invalid label %q", s)
				}
				continue
			}
			if r > n {
				continue
			}
			q := delta
			for k := base; ; k += base {
				t := k - bias
				if t < tmin {
					t = tmin
				} else if t > tmax {
					t = tmax
				}
				if q < t {
					break
				}
				output = append(output, encodeDigit(t+(q-t)%(base-t)))
				q = (q - t) / (base - t)
			}
			output = append(output, encodeDigit(q))
			bias = adapt(delta, h+1, h == b)
			delta = 0
			h++
			remaining--
		}
		delta++
		n++
	}
	return string(output), nil
}

func encodeDigit(digit int32) byte {
	switch {
	case 0 <= digit && digit < 26:
		return byte(digit + 'a')
	case 26 <= digit && digit < 36:
		return byte(digit + ('0' - 26))
	}
	panic("cookiejar: internal error in punycode encoding")
}

// adapt is the bias adaptation function specified in section 6.1.
func adapt(delta, numPoints int32, firstTime bool) int32 {
	if firstTime {
		delta /= damp
	} else {
		delta /= 2
	}
	delta += delta / numPoints
	k := int32(0)
	for delta > ((base-tmin)*tmax)/2 {
		delta /= base - tmin
		k += base
	}
	return k + (base-tmin+1)*delta/(delta+skew)
}

// toASCII converts a domain or domain label to its ASCII form. For example,
// toASCII("bücher.example.com") is "xn--bcher-kva.example.com", and
// toASCII("golang") is "golang".
func toASCII(s string) (string, error) {
	// acePrefix is the ASCII Compatible Encoding prefix.
	const acePrefix = "xn--"

	if isASCII(s) {
		return s, nil
	}
	labels := strings.Split(s, ".")
	for i, label := range labels {
		if !isASCII(label) {
			a, err := encode(acePrefix, label)
			if err != nil {
				return "", err
			}
			labels[i] = a
		}
	}
	return strings.Join(labels, "."), nil
}

// toLower returns the lowercase version of s if s is ASCII and printable.
func toLower(s string) (lower string, ok bool) {
	if !isPrint(s) {
		return "", false
	}

	return strings.ToLower(s), true
}

// isPrint returns whether s is ASCII and printable according to
// https://tools.ietf.org/html/rfc20#section-4.2.
func isPrint(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < ' ' || s[i] > '~' {
			return false
		}
	}
	return true
}

// isASCII returns whether s is ASCII.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
