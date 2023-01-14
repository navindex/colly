package logger

// ------------------------------------------------------------------------

// Logger represents a logger that processes events.
type Logger interface {
	LogEvent(level Level, e *Event) // LogEvent logs an event.
	LogError(level Level, e error)  // LogError logs an error.
}

// A Level is a logging priority. Higher levels are more important.
type Level uint8

// Event represents an action inside a collector.
type Event struct {
	Type        string            // Type is the type of the event
	RequestID   uint32            // RequestID identifies the HTTP request of the Event
	CollectorID uint32            // CollectorID identifies the collector of the Event
	Values      map[string]string // Values contains the event's key-value pairs.
}

// ------------------------------------------------------------------------

// Logging levels
const (
	DEBUG_LEVEL Level = iota
	INFO_LEVEL
	WARN_LEVEL
	ERR_LEVEL
	FATAL_LEVEL
)

// ------------------------------------------------------------------------

var levelNames = []string{"DEBUG", "INFO", "WARN", "ERROR"}

// ------------------------------------------------------------------------

// NewEvent returns a pointer to a newly created event.
func NewEvent(eventType string, collectorID uint32, requestID uint32, args map[string]string) *Event {
	return &Event{
		CollectorID: collectorID,
		RequestID:   requestID,
		Type:        eventType,
		Values:      args,
	}
}
