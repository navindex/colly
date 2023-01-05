package colly

import (
	"bytes"
	"context"
	"encoding/gob"
	"net/http"
	"strings"
)

// ------------------------------------------------------------------------

// Request is an extended HTTP request made by a Collector.
type Request struct {
	ID     uint64           `json:"id" bson:"id,omitempty"`                     // ID is the unique identifier of the request.
	Depth  uint16           `json:"depth" bson:"depth,omitempty"`               // Depth is the number of the parents of the request.
	Req    *http.Request    `json:"http_request" bson:"http_request,omitempty"` // Req is the embedded HTTP request.
	Ctx    *context.Context `json:"context" bson:"context,omitempty"`           // Ctx carries values between request and response.
	Proxy  Proxy            `json:"proxy" bson:"proxy,omitempty"`               // Proxy is the proxy service that handles the request.
	Parser Parser           `json:"proxy" bson:"proxy,omitempty"`               // Parser is the URL parser service.
	Tracer Tracer           `json:"tracer" bson:"tracer,omitempty"`             // Tracer is a request tracing service.

	collector *Collector
	abort     bool
	// baseURL   *url.URL
}

// type requestHandler struct{}

// ------------------------------------------------------------------------

// NewRequest returns a pointer to a newly created request.
func NewRequest(method string, rawURL string, parser Parser) (*Request, error) {
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		return nil, err
	}

	if parser != nil {
		if URL, err := parser.Parse(rawURL); err == nil {
			req.URL = URL
		}
	}

	ctx := context.Background()

	return &Request{
		Req:    req,
		Ctx:    &ctx,
		Parser: parser,
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

// func (rp *requestHandler) Start() {

// }

// ------------------------------------------------------------------------

// func (rp *requestHandler) Stop() {

// }

// ------------------------------------------------------------------------

// HasVisited checks if the provided URL has been visited.
func (r *Request) HasVisited(URL string) (bool, error) {
	return r.collector.HasVisited(URL)
}

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
func (r *Request) Post(URL string, requestData map[string]string) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, createFormReader(requestData), r.Ctx, nil, true)
}

// ------------------------------------------------------------------------

// PostRaw starts a collector job by creating a POST request with raw binary data.
// PostRaw preserves the Context of the previous request.
// It also calls the previously provided callbacks.
func (r *Request) PostRaw(URL string, requestData []byte) error {
	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, bytes.NewReader(requestData), r.Ctx, nil, true)
}

// ------------------------------------------------------------------------

// PostMultipart starts a collector job by creating a Multipart POST request
// with raw binary data.
// It also calls the previously provided callbacks.
func (r *Request) PostMultipart(URL string, requestData map[string][]byte) error {
	boundary := randomBoundary()

	hdr := http.Header{}
	hdr.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	hdr.Set("User-Agent", r.collector.UserAgent)

	return r.collector.scrape(r.AbsoluteURL(URL), "POST", r.Depth+1, createMultipartReader(boundary, requestData), r.Ctx, hdr, true)
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
// 		body, err = ioutil.ReadAll(r.Body)
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
