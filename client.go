package colly

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ------------------------------------------------------------------------

type Client struct {
	// DefConfig is the default client configuration settings if the URL doesn't match any filter criteria in ConfigList.
	DefConfig *clientConfig `json:"default_config" bson:"default_config,omitempty"`
	// ConfigList is a list of client configuration settings based on URL filter criteria.
	ConfigList []*clientConfig `json:"config_list" bson:"config_list,omitempty"`
	// Clt is the embedded HTTP client.
	Clt *http.Client `json:"http_client" bson:"http_client,omitempty"`
	// Cache attaches a cache service to keep a local copy of the responses.
	Cache `json:"cache" bson:"cache,omitempty"`
	// Proxy is a represents a web proxy service.
	Proxy `json:"proxy" bson:"proxy,omitempty"`
	// Tracer attaches a tracing service to enable capturing and reporting request performance for crawler tuning.
	Tracer `json:"tracer" bson:"tracer,omitempty"`

	lock *sync.RWMutex
}

// clientConfig is the internal representation of a specific client settings
type clientConfig struct {
	fc       *FilteredConfig
	waitChan chan bool
}

// ------------------------------------------------------------------------

// NewClient returns a pointer to a newly created client.
func NewClient(config *CollectorConfig) *Client {
	var configs []*clientConfig

	for i := range config.FilteredConfigs {
		configs = append(configs, &clientConfig{
			fc:       config.FilteredConfigs[i],
			waitChan: make(chan bool),
		})
	}

	return &Client{
		DefConfig: &clientConfig{
			fc: &FilteredConfig{
				Filter:      config.Filter,
				Delay:       config.Delay,
				RandomDelay: config.RandomDelay,
				MaxThreads:  config.MaxThreads,
			},
			waitChan: make(chan bool),
		},
		ConfigList: configs,
		Clt: &http.Client{
			Jar: config.CookieJar,
		},
		Cache:  config.Cache,
		Proxy:  config.Proxy,
		Tracer: config.Tracer,
		lock:   &sync.RWMutex{},
	}
}

// ------------------------------------------------------------------------
// Do sends checks the cache for a response or sends an HTTP request and returns an HTTP response,
// following policy (such as redirects, cookies, auth) as configured on the client.
// If the response was a success, it also tries to cache the response.
func (c *Client) Do(req *http.Request, bodySize int, checkHdrFunc checkHeadersFunc) (*Response, error) {
	useCache := req.Method == "GET" && req.Header.Get("Cache-Control") != "no-cache" && c.hasCache()

	// Try to serve the response from cache
	if useCache {
		if resp, err := c.Cache.Get(req.URL.String()); err == nil {
			return resp, nil
		}
	}

	resp, err := c.do(req, bodySize, checkHdrFunc)
	if err != nil || resp.Resp.StatusCode >= 500 || !useCache {
		return resp, err
	}

	return resp, c.Cache.Set(resp)
}

// ------------------------------------------------------------------------

// Sleep pauses the execution for the duration in the client config,
// or the default duration if the URL doesn't match any filter criteria.
func (c *Client) Sleep(URL *url.URL) {
	c.Match(URL).sleep()
}

// ------------------------------------------------------------------------

// Match returns the first client configuration settings where the URL matches the filter criteria.
// If there's no match, it returns the default client settings.
func (c *Client) Match(URL *url.URL) *clientConfig {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if len(c.ConfigList) == 0 {
		return c.DefConfig
	}

	for i := range c.ConfigList {
		if c.ConfigList[i].fc.Match(URL) {
			return c.ConfigList[i]
		}
	}

	return c.DefConfig
}

// ------------------------------------------------------------------------

// The sleep method pauses the execution for a random delay that is calculateed
// by combining the fix and a randomised delay of the client configuration settings.
func (cc *clientConfig) sleep() {
	delay := cc.fc.Delay

	if cc.fc.RandomDelay != 0 {
		delay += time.Duration(rand.Int63n(int64(cc.fc.RandomDelay)))
	}

	if delay <= 0 {
		return
	}

	time.Sleep(delay)
}

// ------------------------------------------------------------------------

func (c *Client) do(req *http.Request, bodySize int, checkHdrFunc checkHeadersFunc) (*Response, error) {
	defer c.Sleep(req.URL)

	if r != nil {
		r.waitChan <- true
		defer func(r *LimitRule) {
			randomDelay := time.Duration(0)
			if r.RandomDelay != 0 {
				randomDelay = time.Duration(rand.Int63n(int64(r.RandomDelay)))
			}
			time.Sleep(r.Delay + randomDelay)
			<-r.waitChan
		}(r)
	}

	res, err := h.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.Request != nil {
		*request = *res.Request
	}

	if !checkHeadersFunc(request, res.StatusCode, res.Header) {
		// closing res.Body (see defer above) without reading it aborts
		// the download
		return nil, ErrAbortedAfterHeaders
	}

	var bodyReader io.Reader = res.Body
	if bodySize > 0 {
		bodyReader = io.LimitReader(bodyReader, int64(bodySize))
	}
	contentEncoding := strings.ToLower(res.Header.Get("Content-Encoding"))
	if !res.Uncompressed && (strings.Contains(contentEncoding, "gzip") || (contentEncoding == "" && strings.Contains(strings.ToLower(res.Header.Get("Content-Type")), "gzip")) || strings.HasSuffix(strings.ToLower(request.URL.Path), ".xml.gz")) {
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			return nil, err
		}
		defer bodyReader.(*gzip.Reader).Close()
	}

	body, err := ioutil.ReadAll(bodyReader)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: res.StatusCode,
		Body:       body,
		Headers:    &res.Header,
	}, nil
}

// ------------------------------------------------------------------------

func (c *Client) hasCache() bool {
	return c.Cache != nil
}
