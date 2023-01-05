package colly

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

// ------------------------------------------------------------------------

// Response is an encapsulated HTTP response, created by a Collector.
type Response struct {
	Response      *http.Response `json:"response" bson:"response,omitempty"`       // Response is the embedded HTTP response.
	Request       *Request       `json:"request" bson:"request,omitempty"`         // Request is the embedded Request.
	ExtStatusCode int            `json:"status_code" bson:"status_code,omitempty"` // StatusCode is the extended response status code.
	Body          []byte         `json:"body" bson:"body,omitempty"`               // Body is the content of the Response

}

// ------------------------------------------------------------------------

func (r *Response) fixCharset(detectCharset bool, defaultEncoding string) error {
	if len(r.Body) == 0 {
		return nil
	}

	if defaultEncoding != "" {
		tmpBody, err := encodeBytes(r.Body, "text/plain; charset="+defaultEncoding)
		if err != nil {
			return err
		}
		r.Body = tmpBody
		return nil
	}

	contentType := strings.ToLower(r.Response.Header.Get("Content-Type"))

	if strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/") {
		// These MIME types should not have textual data.

		return nil
	}

	if !strings.Contains(contentType, "charset") {
		if !detectCharset {
			return nil
		}

		d := chardet.NewTextDetector()
		r, err := d.DetectBest(r.Body)
		if err != nil {
			return err
		}

		contentType = "text/plain; charset=" + r.Charset
	}

	if strings.Contains(contentType, "utf-8") || strings.Contains(contentType, "utf8") {
		return nil
	}

	tmpBody, err := encodeBytes(r.Body, contentType)
	if err != nil {
		return err
	}

	r.Body = tmpBody

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
