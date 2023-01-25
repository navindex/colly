package colly

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

// ------------------------------------------------------------------------

// Request is an extended HTTP request made by a Collector.
type Request struct {
	ID     uint32           `json:"id" bson:"id,omitempty"`                     // ID is the unique identifier of the request.
	Depth  uint16           `json:"depth" bson:"depth,omitempty"`               // Depth is the number of the parents of the request.
	Req    *http.Request    `json:"http_request" bson:"http_request,omitempty"` // Req is the embedded HTTP request.
	Ctx    *context.Context `json:"context" bson:"context,omitempty"`           // Ctx carries values between request and response.
	Parser Parser           `json:"parser" bson:"parser,omitempty"`             // Parser is the URL parser service.
	Tracer Tracer           `json:"tracer" bson:"tracer,omitempty"`             // Tracer is a request tracing service.

	// CharEncode is the character encoding of the response body.
	// Leave it blank to allow automatic character encoding of the response body.
	// It is empty by default and it can be set in OnRequest callback.
	CharEncoding string `json:"char_encoding" bson:"char_encoding,omitempty"`

	collector *Collector
	abort     bool
	baseURL   *url.URL
}

// type requestHandler struct{}

// ------------------------------------------------------------------------

// NewRequest returns a pointer to a newly created request.
func NewRequest(method string, rawURL string, parser Parser, tracer Tracer, body io.ReadCloser) (*Request, error) {
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, err
	}

	if parser == nil {
		parser = NewWHATWGParser()
	}

	URL, err := parser.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	req.URL = URL
	ctx := context.Background()

	return &Request{
		Req:    req,
		Ctx:    &ctx,
		Parser: parser,
		Tracer: tracer,
	}, nil
}

// ------------------------------------------------------------------------

// NewRequestFromBytes extracts the binary data into a newly created request.
func NewRequestFromBytes(b []byte) (*Request, error) {
	// Convert byte slice to io.Reader
	reader := bytes.NewReader(b)

	// Decode into a new request
	r := &Request{}
	err := gob.NewDecoder(reader).Decode(r)
	if err != nil {
		return nil, err
	}

	return r, err
}

// ------------------------------------------------------------------------

// Clone creates a new request with the context of the original request.
func (r *Request) Clone(method string, rawURL string, body io.ReadCloser) (*Request, error) {
	if r.Req == nil {
		return nil, ErrNoHTTPRequest
	}

	if r.collector == nil {
		return nil, ErrNoCollector
	}

	req, err := http.NewRequestWithContext(r.Req.Context(), method, rawURL, body)
	if err != nil {
		return nil, err
	}

	URL, err := r.Parser.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	req.URL = URL
	if h := r.Req.Header.Get("Host"); h != "" {
		req.Header.Set("Host", h)
	}

	return &Request{
		ID:        atomic.AddUint32(&r.collector.requestCount, 1),
		Req:       req,
		Ctx:       r.Ctx,
		Parser:    r.Parser,
		Tracer:    r.Tracer,
		collector: r.collector,
	}, nil
}

// ------------------------------------------------------------------------

// Abort prevents to start further requests.
func (r *Request) Abort() {
	r.abort = true
}

// ------------------------------------------------------------------------

// func (rp *requestHandler) Start() {

// }

// ------------------------------------------------------------------------

// func (rp *requestHandler) Stop() {

// }

// ------------------------------------------------------------------------

// HasVisited checks if the provided URL has been visited.
// func (r *Request) HasVisited(URL string) (bool, error) {
// 	return r.collector.HasVisited(URL)
// }

// ------------------------------------------------------------------------

// Visit continues Collector's collecting job by creating a request and
// preserves the Context of the previous request.
// It also calls the previously provided callbacks.
func (r *Request) Visit(URL string) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "GET", r.Depth+1, nil, r.Ctx, nil, true)
}

// ------------------------------------------------------------------------

// Post continues a collector job by creating a POST request and
// preserves the context of the previous request.
// It also calls the previously provided callbacks.
func (r *Request) Post(URL string, reqData map[string]string) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, NewFormReader(reqData), r.Ctx, nil, true)
}

// ------------------------------------------------------------------------

// PostRaw starts a collector job by creating a POST request with raw binary data.
// PostRaw preserves the Context of the previous request.
// It also calls the previously provided callbacks.
func (r *Request) PostRaw(URL string, reqData []byte) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, bytes.NewReader(reqData), r.Ctx, nil, true)
}

// ------------------------------------------------------------------------

// PostMultipart starts a collector job by creating a Multipart POST request
// with raw binary data.
// It also calls the previously provided callbacks.
func (r *Request) PostMultipart(URL string, reqData map[string][]byte) error {
	boundary := RandomString(30)

	hdr := http.Header{}
	hdr.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	hdr.Set("User-Agent", r.collector.Config.UserAgentCallback())

	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, NewMultipartReader(boundary, reqData), r.Ctx, hdr, true)
}

// ------------------------------------------------------------------------

// Retry submits HTTP request again with the same parameters.
func (r *Request) Retry() error {
	r.Req.Header.Del("Cookie")
	return r.collector.scrape(r.Req.URL.String(), r.Req.Method, r.Depth, r.Req.Body, r.Ctx, r.Req.Header, false)
}

// ------------------------------------------------------------------------

// Do submits the request.
func (r *Request) Do() error {
	return r.collector.scrape(r.Req.URL.String(), r.Req.Method, r.Depth, r.Req.Body, r.Ctx, r.Req.Header, !r.collector.AllowURLRevisit)
}

// ------------------------------------------------------------------------

// ToBytes converts the request to bytes.
func (r *Request) ToBytes() ([]byte, error) {
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(r)

	return b.Bytes(), err
}

// ------------------------------------------------------------------------

// Marshal serializes the Request
// func (r *Request) Marshal() ([]byte, error) {
// 	ctx := make(map[string]any)
// 	if r.Ctx != nil {
// 		r.Ctx.ForEach(func(k string, v any) any {
// 			ctx[k] = v
// 			return nil
// 		})
// 	}
// 	var err error
// 	var body []byte
// 	if r.Body != nil {
// 		body, err = io.ReadAll(r.Body)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	sr := &serializableRequest{
// 		URL:    r.URL.String(),
// 		Host:   r.Host,
// 		Method: r.Method,
// 		Depth:  r.Depth,
// 		Body:   body,
// 		ID:     r.ID,
// 		Ctx:    ctx,
// 	}
// 	if r.Headers != nil {
// 		sr.Headers = *r.Headers
// 	}
// 	return json.Marshal(sr)
// }

// ------------------------------------------------------------------------

// AbsoluteURL returns the resolved absolute URL of an URL chunk.
// It returns empty string if the URL chunk is a fragment or could not be parsed.
func (r *Request) AbsoluteURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "#") {
		return ""
	}

	absURL, err := r.Parser.ParseRef(r.Req.URL.String(), rawURL)
	if err != nil {
		return ""
	}

	return absURL.String()
}

// ------------------------------------------------------------------------

// WithTrace returns the embedded HTTP Request with HTTP Trace added to its context.
func WithTrace(req *http.Request, t Tracer) *http.Request {
	return req.WithContext(t.WithContext(req.Context()))
}
