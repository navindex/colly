package logger

import (
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"
)

// ------------------------------------------------------------------------

// stdLogger is the internal structure of an embedded standard logger.
type stdLogger struct {
	l       *log.Logger
	counter int32
	start   time.Time
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

// ------------------------------------------------------------------------

// Log logs an event.
func (l *stdLogger) Log(level Level, e *Event) {
	i := atomic.AddInt32(&l.counter, 1)
	l.l.Printf("%s: [%06d] %d [%6d - %s] %q (%s)\n", levelNames[level], i, e.CollectorID, e.RequestID, e.Type, e.Values, time.Since(l.start))
}
