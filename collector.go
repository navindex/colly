package colly

import (
	"bytes"
	"colly/storage"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xmlquery"
	"github.com/temoto/robotstxt"
)

// ------------------------------------------------------------------------

// EventCallbacks is an ordered list of callback functions, grouped by events.
type EventCallbacks interface {
	Add(event uint8, arg string, fn any, index ...int) // Add inserts ar appends a new callback function.
	Remove(event uint8, arg string, index ...int)      // Remove removes some or all of the event functions.
	Get(event uint8) map[string][]any                  // Get retrieves all callback functions attached to an event, mapped to the arguments.
	GetArg(event uint8, arg string) []any              // GetArg retrieves all callback functions attached to an event with an argument.
	Count(event uint8, arg ...string) int              // Count returns the number of items attached to an event or argument.
	IsEmpty(event uint8, arg ...string) bool           //IsEmpty returns true if no callback attached to the event or argument; otherwise returns false.
}

// Callback functions
type (
	RequestCallback         func(*Request)         // RequestCallback is a type alias for OnRequest callback functions.
	ResponseHeadersCallback func(*Response)        // ResponseHeadersCallback is a type alias for OnResponseHeaders callback functions.
	ResponseCallback        func(*Response)        // ResponseCallback is a type alias for OnResponse callback functions.
	ErrorCallback           func(*Response, error) // ErrorCallback is a type alias for OnError callback functions.
	HTMLCallback            func(*HTMLElement)     // HTMLCallback is a type alias for OnHTML callback functions.
	XMLCallback             func(*XMLElement)      // XMLCallback is a type alias for OnXML callback functions.
	ScrapedCallback         func(*Response)        // ScrapedCallback is a type alias for OnScraped callback functions.
)

// Collector represents the individual settings of a collector.
type Collector struct {
	ID        uint32           `json:"id" bson:"id,omitempty"`               // ID is the unique identifier of a collector.
	Config    *CollectorConfig `json:"config" bson:"config,omitempty"`       // Config represents the collector's configuration settings.
	Callbacks EventCallbacks   `json:"callbacks" bson:"callbacks,omitempty"` // Callbacks contains the callback functions for the events.
	Ctx       *context.Context `json:"context" bson:"context,omitempty"`     // Context is the context that will be used for HTTP requests.

	sysCallbacks EventCallbacks // system callback functions will be called before other callbacks

	store         storage.BaseStorage
	robotsMap     map[string]*robotstxt.RobotsData
	requestCount  uint32
	responseCount uint32
	backend       *httpBackend
	wg            *sync.WaitGroup
	lock          *sync.RWMutex
}

// ------------------------------------------------------------------------

// Collector events that can be caught by an event handler
const (
	ON_REQUEST uint8 = iota
	ON_RESPONSE_HDR
	ON_RESPONSE
	ON_ERROR
	ON_HTML
	ON_XML
	ON_SCRAPED
)

// Empty event argument.
const NO_ARG string = ""

// ------------------------------------------------------------------------

// NewCollector returns a pointer to a newly created Collector instance.
func NewCollector(config *CollectorConfig, callbacks EventCallbacks) *Collector {
	if config == nil {
		config = NewConfig()
	}

	if callbacks == nil {
		callbacks = NewEventList()
	}

	return &Collector{
		Config:       config,
		Callbacks:    callbacks,
		sysCallbacks: NewEventList(),
	}
}

// ------------------------------------------------------------------------

// OnRequest is convenience method to register a function
// that will be executed before every request made by the Collector.
// The position identifies the execution order.
func (c *Collector) OnRequest(fn RequestCallback, position ...int) {
	c.Callbacks.Add(ON_REQUEST, NO_ARG, fn, position...)
}

// OnRequestDetach removes a number of registered request callback functions.
// If no position was given, all request callback functions will be removed.
func (c *Collector) OnRequestDetach(position ...int) {
	c.Callbacks.Remove(ON_REQUEST, NO_ARG, position...)
}

func (c *Collector) handleOnRequest(r *Request) {
	if c.HasLogger() {
		c.logEvent(LOG_INFO_LEVEL, "request", r.ID, map[string]string{
			"url": r.Req.URL.String(),
		})
	}

	for _, fn := range c.Callbacks.GetArg(ON_REQUEST, NO_ARG) {
		if callback, ok := fn.(RequestCallback); ok {
			callback(r)
		}
	}
}

// ------------------------------------------------------------------------

// OnResponseHeaders is convenience method to register a function
// that will be executed after every response when headers and status
// are already received, but body is not yet read.
// The position identifies the execution order.
// Like in OnRequest, you can call Request.Abort to abort the transfer. This might be
// useful if, for example, you're following all hyperlinks, but want to avoid
// downloading files.
// Be aware that using this will prevent HTTP/1.1 connection reuse, as
// the only way to abort a download is to immediately close the connection.
// HTTP/2 doesn't suffer from this problem, as it's possible to close
// specific stream inside the connection.
func (c *Collector) OnResponseHeaders(fn ResponseHeadersCallback, position ...int) {
	c.Callbacks.Add(ON_RESPONSE_HDR, NO_ARG, fn, position...)
}

// OnResponseHeadersDetach removes a number of registered response header callback functions.
// If no position was given, all response header callback functions will be removed.
func (c *Collector) OnResponseHeadersDetach(position ...int) {
	c.Callbacks.Remove(ON_RESPONSE_HDR, NO_ARG, position...)
}

func (c *Collector) handleOnResponseHeaders(resp *Response) {
	if c.HasLogger() {
		level := LOG_INFO_LEVEL
		if resp.Resp.StatusCode >= 300 {
			level = LOG_WARN_LEVEL
		}
		c.logEvent(level, "response_hdr", resp.Request.ID, map[string]string{
			"url":         resp.Request.Req.URL.String(),
			"status_code": strconv.Itoa(resp.Resp.StatusCode),
			"status_msg":  resp.Resp.Status,
		})
	}

	for _, fn := range c.Callbacks.GetArg(ON_RESPONSE_HDR, NO_ARG) {
		if callback, ok := fn.(ResponseHeadersCallback); ok {
			callback(resp)
		}
	}
}

// ------------------------------------------------------------------------

// OnResponse is convenience method to register a function that will be executed
// after every response. The position identifies the execution order.
func (c *Collector) OnResponse(fn ResponseCallback, position ...int) {
	c.Callbacks.Add(ON_RESPONSE, NO_ARG, fn, position...)
}

// OnResponseDetach removes a number of registered response callback functions.
// If no position was given, all response callback functions will be removed.
func (c *Collector) OnResponseDetach(position ...int) {
	c.Callbacks.Remove(ON_RESPONSE, NO_ARG, position...)
}

func (c *Collector) handleOnResponse(resp *Response) {
	if !c.Config.ParseStatusCallback(resp.Resp.StatusCode) {
		return
	}

	if c.HasLogger() {
		c.logEvent(LOG_INFO_LEVEL, "response", resp.Request.ID, map[string]string{
			"url":         resp.Request.Req.URL.String(),
			"status_code": strconv.Itoa(resp.Resp.StatusCode),
			"status_msg":  resp.Resp.Status,
		})
	}

	for _, fn := range c.Callbacks.GetArg(ON_RESPONSE, NO_ARG) {
		if callback, ok := fn.(ResponseCallback); ok {
			callback(resp)
		}
	}
}

// ------------------------------------------------------------------------

// OnError is convenience method to register a function that will be executed
// after an error occurs during the HTTP request.
// The position identifies the execution order.
func (c *Collector) OnError(fn ErrorCallback, position ...int) {
	c.Callbacks.Add(ON_ERROR, NO_ARG, fn, position...)
}

// OnErrorDetach removes a number of registered error response callback functions.
// If no position was given, all error response callback functions will be removed.
func (c *Collector) OnErrorDetach(position ...int) {
	c.Callbacks.Remove(ON_ERROR, NO_ARG, position...)
}

func (c *Collector) handleOnError(resp *Response, err error, ctx *Context) error {
	if resp == nil {
		response = &Response{
			Request: request,
			Ctx:     ctx,
		}
	}

	if err == nil && resp != nil && resp.Resp != nil && c.Config.ParseStatusCallback(resp.Resp.StatusCode) {
		return nil
	}

	if err == nil && (c.ParseHTTPErrorResponse || resp.StatusCode < 203) {
		return nil
	}
	if err == nil && resp.Resp.StatusCode >= 203 {
		err = errors.New(http.StatusText(resp.Resp.StatusCode))
	}
	if c.HasLogger() {
		c.logEvent(LOG_WARN_LEVEL, "error", resp.Request.ID, map[string]string{
			"url":         resp.Request.Req.URL.String(),
			"status_code": strconv.Itoa(resp.Resp.StatusCode),
			"status_msg":  resp.Resp.Status,
		})
	}
	if resp.Request == nil {
		resp.Request = request
	}
	if response.Ctx == nil {
		response.Ctx = request.Ctx
	}

	for _, fn := range c.Callbacks.GetArg(ON_ERROR, NO_ARG) {
		if callback, ok := fn.(ErrorCallback); ok {
			callback(resp, err)
		}
	}

	return err
}

// ------------------------------------------------------------------------

// OnHTML is convenience method to register a function that will be executed
// on every HTML element matched by the GoQuery Selector parameter.
// GoQuery Selector is a selector used by https://github.com/PuerkitoBio/goquery
func (c *Collector) OnHTML(goquerySelector string, fn HTMLCallback, position ...int) {
	c.Callbacks.Add(ON_HTML, goquerySelector, fn, position...)
}

// OnHTMLDetach removes a number of registered HTML callback functions.
// If no position was given, all functions will be removed for the given GoQuery Selector.
func (c *Collector) OnHTMLDetach(goquerySelector string, position ...int) {
	c.Callbacks.Remove(ON_HTML, goquerySelector, position...)
}

func (c *Collector) handleOnHTML(resp *Response) error {
	if c.Callbacks.IsEmpty(ON_HTML) || !strings.Contains(strings.ToLower(resp.Resp.Header.Get("Content-Type")), "html") {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(resp.Body))
	if err != nil {
		return err
	}

	if href, found := doc.Find("base[href]").Attr("href"); found {
		baseURL, err := c.Config.Parser.ParseRef(resp.Request.Req.URL.String(), href)
		if err == nil {
			resp.Request.baseURL = baseURL
		}

	}
	for selector, fnList := range c.Callbacks.Get(ON_HTML) {
		i := 0
		doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
			for _, n := range s.Nodes {
				e := NewHTMLElementFromSelectionNode(resp, s, n, i)
				i++
				if c.HasLogger() {
					c.logEvent(LOG_INFO_LEVEL, "html", resp.Request.ID, map[string]string{
						"selector": selector,
						"url":      resp.Request.Req.URL.String(),
					})
				}

				for _, fn := range fnList {
					if callback, ok := fn.(HTMLCallback); ok {
						callback(e)
					}
				}
			}
		})
	}
	return nil
}

// ------------------------------------------------------------------------

// OnXML is convenience method to register a function that will be executed
// on every XML element matched by the Xpath uery parameter.
// xpath Query is used by https://github.com/antchfx/xmlquery
func (c *Collector) OnXML(xpathQuery string, fn XMLCallback, position ...int) {
	c.Callbacks.Add(ON_XML, xpathQuery, fn, position...)
}

// OnXMLDetach removes a number of registered XML callback functions.
// If no position was given, all functions will be removed for the given Xpath Query.
func (c *Collector) OnXMLDetach(xpathQuery string, position ...int) {
	c.Callbacks.Remove(ON_XML, xpathQuery, position...)
}

func (c *Collector) handleOnXML(resp *Response) error {
	if c.Callbacks.IsEmpty(ON_XML) {
		return nil
	}

	contentType := strings.ToLower(resp.Resp.Header.Get("Content-Type"))
	isXMLFile := isXML(resp.Request.Req.URL.Path)
	if !strings.Contains(contentType, "html") && (!strings.Contains(contentType, "xml") && !isXMLFile) {
		return nil
	}

	if strings.Contains(contentType, "html") {
		doc, err := htmlquery.Parse(bytes.NewBuffer(resp.Body))
		if err != nil {
			return err
		}
		if e := htmlquery.FindOne(doc, "//base"); e != nil {
			for _, a := range e.Attr {
				if a.Key == "href" {
					baseURL, err := c.Config.Parser.Parse(a.Val)
					if err == nil {
						resp.Request.baseURL = baseURL
					}
					break
				}
			}
		}

		for query, fnList := range c.Callbacks.Get(ON_XML) {
			for _, n := range htmlquery.Find(doc, query) {
				e := NewXMLElementFromHTMLNode(resp, n)

				if c.HasLogger() {
					c.logEvent(LOG_INFO_LEVEL, "xml", resp.Request.ID, map[string]string{
						"selector": query,
						"url":      resp.Request.Req.URL.String(),
					})
				}

				for _, fn := range fnList {
					if callback, ok := fn.(XMLCallback); ok {
						callback(e)
					}
				}
			}
		}
	} else if strings.Contains(contentType, "xml") || isXMLFile {
		doc, err := xmlquery.Parse(bytes.NewBuffer(resp.Body))
		if err != nil {
			return err
		}

		for query, fnList := range c.Callbacks.Get(ON_XML) {
			xmlquery.FindEach(doc, query, func(i int, n *xmlquery.Node) {
				e := NewXMLElementFromXMLNode(resp, n)

				if c.HasLogger() {
					c.logEvent(LOG_INFO_LEVEL, "xml", resp.Request.ID, map[string]string{
						"selector": query,
						"url":      resp.Request.Req.URL.String(),
					})
				}

				for _, fn := range fnList {
					if callback, ok := fn.(XMLCallback); ok {
						callback(e)
					}
				}
			})
		}
	}
	return nil
}

// ------------------------------------------------------------------------

// OnScraped is convenience method to register a function that will be executed
// as a final part of the scraping. The position identifies the execution order.
func (c *Collector) OnScraped(fn ScrapedCallback, position ...int) {
	c.Callbacks.Add(ON_SCRAPED, NO_ARG, fn, position...)
}

// OnScrapedDetach removes a number of registered scraped callback functions.
// If no position was given, all scraped callback functions will be removed.
func (c *Collector) OnScrapedDetach(position ...int) {
	c.Callbacks.Remove(ON_SCRAPED, NO_ARG, position...)
}

func (c *Collector) handleOnScraped(resp *Response) {
	if c.HasLogger() {
		c.logEvent(LOG_INFO_LEVEL, "scraped", resp.Request.ID, map[string]string{
			"url": resp.Request.Req.URL.String(),
		})
	}

	for _, fn := range c.Callbacks.GetArg(ON_SCRAPED, NO_ARG) {
		if callback, ok := fn.(ScrapedCallback); ok {
			callback(resp)
		}
	}
}

// ------------------------------------------------------------------------

// ------------------------------------------------------------------------

// ------------------------------------------------------------------------

// ------------------------------------------------------------------------

func (c *Collector) HasLogger() bool {
	return c.Config.hasLogger()
}

// ------------------------------------------------------------------------

func (c *Collector) logEvent(level LogLevel, eventType string, requestID uint32, args map[string]string) {
	if c.Config.hasLogger() {
		c.Config.Logger.LogEvent(level, NewLoggerEvent(eventType, c.ID, requestID, args))
	}
}

// ------------------------------------------------------------------------

func isXML(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".xml") || strings.HasSuffix(strings.ToLower(path), ".xml.gz")
}
