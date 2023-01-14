package colly

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ------------------------------------------------------------------------

// Logger represents a logger that processes events.
type Logger interface {
	LogEvent(level LogLevel, e *LoggerEvent) // LogEvent logs an event.
	LogError(level LogLevel, e error)        // LogError logs an error.
}

// A LogLevel is a logging priority. Higher levels are more important.
type LogLevel uint8

// LoggerEvent represents an action inside a collector.
type LoggerEvent struct {
	Type        string            // Type is the type of the logger event.
	RequestID   uint32            // RequestID identifies the HTTP request of the logger event.
	CollectorID uint32            // CollectorID identifies the collector of the logger event.
	Values      map[string]string // Values contains the logger event's key-value pairs.
}

// stdLogger is the internal structure of an embedded standard logger.
type stdLogger struct {
	l       *log.Logger
	counter int32
	start   time.Time
}

// webLogger is a web based logger frontend.
type webLogger struct {
	req  map[uint32]webLoggerReqInfo
	resp []webLoggerReqInfo
	sync.Mutex
}

type webLoggerReqInfo struct {
	LogLevel
	ID             uint32
	CollectorID    uint32
	URL            string
	Started        time.Time
	Duration       time.Duration
	ResponseStatus string
}

// ------------------------------------------------------------------------

// Logging levels
const (
	LOG_DEBUG_LEVEL LogLevel = iota
	LOG_INFO_LEVEL
	LOG_WARN_LEVEL
	LOG_ERR_LEVEL
	LOG_FATAL_LEVEL
)

const webLoggerDefaultAddress = "127.0.0.1:7676"

const webLoggerPage = `<!DOCTYPE html>
<html>
<head>
	<title>Colly Debugger WebUI</title>
	<script src="https://code.jquery.com/jquery-latest.min.js" type="text/javascript"></script>
	<link rel="stylesheet" type="text/css" href="https://semantic-ui.com/dist/semantic.min.css">
</head>
<body>
<div class="ui inverted vertical masthead center aligned segment" id="menu">
	<div class="ui tiny secondary inverted menu">
		<a class="item" href="/"><b>Colly WebDebugger</b></a>
	</div>
</div>
<div class="ui grid container">
	<div class="row">
		<div class="eight wide column">
			<h1>Current Requests <span id="current_request_count"></span></h1>
			<div id="current_requests" class="ui small feed"></div>
		</div>
		<div class="eight wide column">
			<h1>Finished Requests <span id="request_log_count"></span></h1>
			<div id="request_log" class="ui small feed"></div>
		</div>
	</div>
</div>
<script>
	function curRequestTpl(url, started, collectorId) {
		return '<div class="event"><div class="content"><div class="summary">' + url + '</div><div class="meta">Collector #' + collectorId + ' - ' + started + "</div></div></div>";
	}
	function requestLogTpl(url, duration, collectorId) {
		return '<div class="event"><div class="content"><div class="summary">' + url + '</div><div class="meta">Collector #' + collectorId + ' - ' + (duration/1000000000) + "s</div></div></div>";
	}
	function fetchStatus() {
		$.getJSON("/status", function(data) {
			$("#current_requests").html("");
			$("#request_log").html("");
			$("#current_request_count").text('(' + Object.keys(data.CurrentRequests).length + ')');
			$("#request_log_count").text('(' + data.RequestLog.length + ')');
			for(var i in data.CurrentRequests) {
				var r = data.CurrentRequests[i];
				$("#current_requests").append(curRequestTpl(r.URL, r.Started, r.CollectorID));
			}
			for(var i in data.RequestLog.reverse()) {
				var r = data.RequestLog[i];
				$("#request_log").append(requestLogTpl(r.URL, r.Duration, r.CollectorID));
			}
			setTimeout(fetchStatus, 1000);
		});
	}
	$(document).ready(function() {
		fetchStatus();
	});
</script>
</body>
</html>
`

// ------------------------------------------------------------------------

var logLevelNames = []string{"DEBUG", "INFO", "WARN", "ERROR"}

// ------------------------------------------------------------------------

// NewLoggerEvent returns a pointer to a newly created event.
func NewLoggerEvent(eventType string, collectorID uint32, requestID uint32, args map[string]string) *LoggerEvent {
	return &LoggerEvent{
		CollectorID: collectorID,
		RequestID:   requestID,
		Type:        eventType,
		Values:      args,
	}
}

// ------------------------------------------------------------------------

// NewStdLogger returns a pointer to a newly created standard logger.
func NewStdLogger(dest io.Writer, prefix string, flag int) *stdLogger {
	if dest == nil {
		dest = os.Stderr
	}

	return &stdLogger{
		l:       log.New(dest, prefix, flag),
		counter: 0,
		start:   time.Now(),
	}
}

// LogEvent logs a logger event.
func (l *stdLogger) LogEvent(level LogLevel, e *LoggerEvent) {
	i := atomic.AddInt32(&l.counter, 1)
	l.l.Printf("%s: [%06d] %d [%6d - %s] %q (%s)\n", logLevelNames[level], i, e.CollectorID, e.RequestID, e.Type, e.Values, time.Since(l.start))
}

// LogError logs an error.
func (l *stdLogger) LogError(level LogLevel, e error) {
	i := atomic.AddInt32(&l.counter, 1)
	l.l.Printf("%s: [%06d]  %s (%s)\n", logLevelNames[level], i, e.Error(), time.Since(l.start))
}

// ------------------------------------------------------------------------

// NewWebLogger returns a pointer to a newly created web logger.
func NewWebLogger(address string) *webLogger {
	if ip := net.ParseIP(address); ip == nil {
		address = webLoggerDefaultAddress
	}

	w := &webLogger{
		req:  map[uint32]webLoggerReqInfo{},
		resp: []webLoggerReqInfo{},
	}

	http.HandleFunc("/", w.indexHandler)
	http.HandleFunc("/status", w.statusHandler)

	go http.ListenAndServe(address, nil)

	return w
}

// LogEvent logs an event.
func (w *webLogger) LogEvent(level LogLevel, e *LoggerEvent) {
	w.Lock()
	defer w.Unlock()

	switch e.Type {
	case "request":
		w.req[e.RequestID] = webLoggerReqInfo{
			CollectorID: e.CollectorID,
			ID:          e.RequestID,
			URL:         e.Values["url"],
			Started:     time.Now(),
		}
	case "response", "error":
		r := w.req[e.RequestID]
		r.LogLevel = level
		r.Duration = time.Since(r.Started)
		if status, ok := e.Values["status"]; ok {
			r.ResponseStatus = status
		}
		w.resp = append(w.resp, r)
		delete(w.req, e.RequestID)
	}
}

// LogError logs an error.
func (l *webLogger) LogError(level LogLevel, e error) {
	// Nothing to do
}

func (w *webLogger) indexHandler(wr http.ResponseWriter, r *http.Request) {
	wr.Write([]byte(webLoggerPage))
}

func (w *webLogger) statusHandler(wr http.ResponseWriter, r *http.Request) {
	w.Lock()
	jsonData, err := json.MarshalIndent(w, "", "  ")
	w.Unlock()
	if err == nil {
		wr.Write(jsonData)
	}
}
