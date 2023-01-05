package colly

import (
	"context"
	"sync"

	"github.com/gocolly/colly/storage"
	"github.com/temoto/robotstxt"
)

// ------------------------------------------------------------------------

// Collector represents the individual settings of a collector.
type Collector struct {
	// ID is the unique identifier of a collector
	ID uint32 `json:"id" bson:"id,omitempty"`
	// Limits is a collection of all request limits.
	Config *CollectorConfig `json:"config" bson:"config,omitempty"`
	// Context is the context that will be used for HTTP requests. You can set this
	// to support clean cancellation of scraping.
	Ctx *context.Context `json:"context" bson:"context,omitempty"`
	// Debugger processes collector events.
	Debugger `json:"debugger" bson:"debugger,omitempty"`

	store                    storage.Storage
	robotsMap                map[string]*robotstxt.RobotsData
	htmlCallbacks            []*htmlCallbackContainer
	xmlCallbacks             []*xmlCallbackContainer
	requestCallbacks         []RequestCallback
	responseCallbacks        []ResponseCallback
	responseHeadersCallbacks []ResponseHeadersCallback
	errorCallbacks           []ErrorCallback
	scrapedCallbacks         []ScrapedCallback
	requestCount             uint32
	responseCount            uint32
	backend                  *httpBackend
	wg                       *sync.WaitGroup
	lock                     *sync.RWMutex
}

// ------------------------------------------------------------------------

// NewCollector creates a new Collector instance with default configuration.
func NewCollector(config *CollectorConfig) *Collector {
	if config == nil {
		config = NewConfig()
	}

	return &Collector{
		Config: config,
	}
}

// ------------------------------------------------------------------------
