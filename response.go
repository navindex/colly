package colly

import (
	"bytes"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

// ------------------------------------------------------------------------

// Response is an encapsulated HTTP response, created by a Collector.
type Response struct {
	Resp          *http.Response `json:"response" bson:"response,omitempty"`       // Response is the embedded HTTP response.
	Request       *Request       `json:"request" bson:"request,omitempty"`         // Request is the embedded Request.
	ExtStatusCode int            `json:"status_code" bson:"status_code,omitempty"` // StatusCode is the extended response status code.
	Body          []byte         `json:"body" bson:"body,omitempty"`               // Body is the content of the Response

}

// ------------------------------------------------------------------------

// FileName returns the sanitized file name parsed from "Content-Disposition"
// header or from URL.
func (r *Response) FileName() string {
	_, params, err := mime.ParseMediaType(r.Resp.Header.Get("Content-Disposition"))

	fName, ok := params["filename"]
	if err != nil || !ok {
		url := r.Request.Req.URL
		fName = strings.TrimPrefix(url.Path, "/")

		if url.RawQuery != "" {
			fName = fName + "_" + url.RawQuery
		}
	}

	return SanitizeFileName(fName)
}

// ------------------------------------------------------------------------

func (r *Response) fixCharset(detectCharset bool, defaultEncoding string) error {
	if len(r.Body) == 0 {
		return nil
	}

	// Use default encoding if exists
	if defaultEncoding != "" {
		body, err := encodeBytes(r.Body, "text/plain; charset="+defaultEncoding)
		if err != nil {
			return err
		}

		r.Body = body

		return nil
	}

	contentType := strings.ToLower(r.Resp.Header.Get("Content-Type"))

	// Exit if content is not textual data
	if noTextualData(contentType) {
		return nil
	}

	// Detect character set if missing
	if !strings.Contains(contentType, "charset") {
		if !detectCharset {
			return nil
		}

		r, err := chardet.NewTextDetector().DetectBest(r.Body)
		if err != nil {
			return err
		}

		contentType = "text/plain; charset=" + r.Charset
	}

	// Nothing more to do if the character set is UTF-8
	if strings.Contains(contentType, "utf-8") || strings.Contains(contentType, "utf8") {
		return nil
	}

	// Convert to the newly set character set
	body, err := encodeBytes(r.Body, contentType)
	if err != nil {
		return err
	}
	r.Body = body

	return nil
}

// ------------------------------------------------------------------------

func encodeBytes(b []byte, contentType string) ([]byte, error) {
	r, err := charset.NewReader(bytes.NewReader(b), contentType)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

// ------------------------------------------------------------------------

func noTextualData(contentType string) bool {
	return strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/")
}
