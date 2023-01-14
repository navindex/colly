package logger

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"
)

// ------------------------------------------------------------------------

// webLogger is a web based logger frontend.
type webLogger struct {
	req  map[uint32]webReqInfo
	resp []webReqInfo
	sync.Mutex
}

type webReqInfo struct {
	Level
	ID             uint32
	CollectorID    uint32
	URL            string
	Started        time.Time
	Duration       time.Duration
	ResponseStatus string
}

// ------------------------------------------------------------------------

const webDefaultAddress = "127.0.0.1:7676"

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

// NewWebLogger returns a pointer to a newly created web logger.
func NewWebLogger(address string) *webLogger {
	if ip := net.ParseIP(address); ip == nil {
		address = webDefaultAddress
	}

	w := &webLogger{
		req:  map[uint32]webReqInfo{},
		resp: []webReqInfo{},
	}

	http.HandleFunc("/", w.indexHandler)
	http.HandleFunc("/status", w.statusHandler)

	go http.ListenAndServe(address, nil)

	return w
}

// ------------------------------------------------------------------------

// LogEvent logs an event.
func (w *webLogger) LogEvent(level Level, e *Event) {
	w.Lock()
	defer w.Unlock()

	switch e.Type {
	case "request":
		w.req[e.RequestID] = webReqInfo{
			CollectorID: e.CollectorID,
			ID:          e.RequestID,
			URL:         e.Values["url"],
			Started:     time.Now(),
		}
	case "response", "error":
		r := w.req[e.RequestID]
		r.Level = level
		r.Duration = time.Since(r.Started)
		if status, ok := e.Values["status"]; ok {
			r.ResponseStatus = status
		}
		w.resp = append(w.resp, r)
		delete(w.req, e.RequestID)
	}
}

// ------------------------------------------------------------------------

// LogError logs an error.
func (l *webLogger) LogError(level Level, e error) {
	// Nothing to do
}

// ------------------------------------------------------------------------

func (w *webLogger) indexHandler(wr http.ResponseWriter, r *http.Request) {
	wr.Write([]byte(webLoggerPage))
}

// ------------------------------------------------------------------------

func (w *webLogger) statusHandler(wr http.ResponseWriter, r *http.Request) {
	w.Lock()
	jsonData, err := json.MarshalIndent(w, "", "  ")
	w.Unlock()
	if err == nil {
		wr.Write(jsonData)
	}
}
