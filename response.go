package colly

import (
	"bytes"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

// ------------------------------------------------------------------------

// Response is an encapsulated HTTP response, created by a Collector.
type Response struct {
	Request       *Request       `json:"request" bson:"request,omitempty"`         // Request is the embedded Request.
	Resp          *http.Response `json:"response" bson:"response,omitempty"`       // Response is the embedded HTTP response.
	ExtStatusCode uint           `json:"status_code" bson:"status_code,omitempty"` // ExtStatusCode is the extended response status code.
	Body          []byte         `json:"body" bson:"body,omitempty"`               // Body is the content of the response.
	Created       time.Time      `json:"created" bson:"created,omitempty"`         // Received is the date and time when the response was created.
	Expiry        time.Time      `json:"expiry" bson:"expiry,omitempty"`           // Expiry is the response expiry date and time.
}

// ------------------------------------------------------------------------

// NewResponse returns a pointer to a newly created response.
func NewResponse(req *Request, resp *http.Response, detectCharset bool) (*Response, error) {
	r := &Response{
		Request: req,
		Resp:    resp,
	}

	if err := r.setBody(detectCharset); err != nil {
		return nil, err
	}

	r.setExtStatusCode()
	r.setCreated()
	r.setExpiry()

	return r, nil
}

// ------------------------------------------------------------------------

func (r *Response) setBody(detectCharset bool) error {
	data, err := io.ReadAll(r.Resp.Body)
	if err == nil {
		return err
	}

	r.Body = data
	if len(r.Body) == 0 {
		return nil
	}

	contentType := strings.ToLower(r.Resp.Header.Get("Content-Type"))

	// Exit if content is not textual data
	if noTextualData(contentType) {
		return nil
	}

	// Use default encoding if exists
	if enc := r.Request.CharEncoding; enc != "" {
		return r.encodeBody("text/plain; charset=" + enc)
	}

	// Exit if no charset with no detect or charset is utf8
	hasCharset := strings.Contains(contentType, "charset")
	if (!hasCharset && !detectCharset) ||
		(hasCharset && ContainsAny(contentType, "utf-8", "utf8")) {
		return nil
	}

	// Detect character set if missing
	res, err := chardet.NewTextDetector().DetectBest(r.Body)
	if err != nil {
		return err
	}
	contentType = "text/plain; charset=" + res.Charset

	// Convert to the newly set character set
	return r.encodeBody(contentType)
}

// ------------------------------------------------------------------------

func (r *Response) setCreated() {
	r.Created = time.Now()

	if ageHdr := r.Resp.Header.Get("Age"); ageHdr != "" {
		if sec, err := strconv.Atoi(ageHdr); err == nil {
			r.Created = r.Created.Add(-time.Duration(sec) * time.Second)
		}
	}
}

// ------------------------------------------------------------------------

func (r *Response) setExpiry() {
	if cc := r.Resp.Header.Get("Cache-Control"); cc != "" {
		if ContainsAny(cc, "no-cache", "no-store") {
			r.Expiry = r.Created

			return
		}

		if sec := findHeaderTokenValue(cc, "max-age"); sec != nil {
			r.Expiry = r.Created.Add(time.Second * time.Duration(*sec))

			return
		}

		if sec := findHeaderTokenValue(cc, "s-maxage"); sec != nil {
			r.Expiry = r.Created.Add(time.Second * time.Duration(*sec))

			return
		}
	}

	if exp := parseHeaderDate(r.Resp.Header.Get("Expires")); exp != nil {
		r.Expiry = *exp

		return
	}

	r.Expiry = time.Unix(1<<63-1, 0)

}

// ------------------------------------------------------------------------

// FIXME
func (r *Response) setExtStatusCode() {
	r.ExtStatusCode = uint(r.Resp.StatusCode)
}

// ------------------------------------------------------------------------

// CacheKey returns a cache key parsed from "Content-Disposition" header or from URL.
func (r *Response) cacheKey() string {
	_, params, err := mime.ParseMediaType(r.Resp.Header.Get("Content-Disposition"))

	key, ok := params["filename"]
	if err != nil || !ok {
		url := r.Request.Req.URL
		key = strings.TrimPrefix(url.Path, "/")

		if url.RawQuery != "" {
			key = key + "_" + url.RawQuery
		}
	}

	return key
}

// ------------------------------------------------------------------------

func (r *Response) encodeBody(contentType string) error {
	rdr, err := charset.NewReader(bytes.NewReader(r.Body), contentType)
	if err == nil {
		r.Body, err = io.ReadAll(rdr)
	}
	return err
}

// ------------------------------------------------------------------------

func noTextualData(contentType string) bool {
	return strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/")
}

// ------------------------------------------------------------------------

func parseHeaderDate(hdr string) *time.Time {
	if hdr == "" {
		return nil
	}

	var (
		t   time.Time
		err error
	)

	if t, err = mail.ParseDate(hdr); err != nil {
		if t, err = time.Parse(time.RFC850, hdr); err != nil {
			if t, err = time.Parse(time.ANSIC, hdr); err != nil {
				return nil
			}
		}
	}

	return &t
}

// ------------------------------------------------------------------------

func findHeaderTokenValue(hdr string, token string) *int {
	token = token + "="

	for _, s := range strings.Split(hdr, ",") {
		s := strings.TrimSpace(s)
		if strings.HasPrefix(s, token) {
			if sec, err := strconv.Atoi(s[len(token):]); err == nil {
				return &sec
			}

			return nil
		}
	}

	return nil
}
