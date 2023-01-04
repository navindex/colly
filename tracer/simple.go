package tracer

import (
	"context"
	"crypto/tls"
	"net/http/httptrace"
	"time"
)

// ------------------------------------------------------------------------

// simpleTracer provides a simple data structure for storing an http trace.
type simpleTracer struct {
	start        time.Time
	connectStart time.Duration
	connectDone  time.Duration
	tlsStart     time.Duration
	tlsDone      time.Duration
	firstByte    time.Duration
	ct           *httptrace.ClientTrace
}

// ------------------------------------------------------------------------

// NewSimpleTracer returns a pointer to a newly created simple tracer.
func NewSimpleTracer() *simpleTracer {
	t := &simpleTracer{}

	t.ct = &httptrace.ClientTrace{
		GetConn:              func(_ string) { t.start = time.Now() },
		ConnectStart:         func(_, _ string) { t.connectStart = time.Since(t.start) },
		ConnectDone:          func(_, _ string, _ error) { t.connectDone = time.Since(t.start) },
		TLSHandshakeStart:    func() { t.tlsStart = time.Since(t.start) },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t.tlsDone = time.Since(t.start) },
		GotFirstResponseByte: func() { t.firstByte = time.Since(t.start) },
	}

	return t
}

// ------------------------------------------------------------------------

// WithContext returns a new context based on the provided parent context.
// HTTP client requests made with the returned context will use the provided trace hooks,
// in addition to any previous hooks registered with ctx.
// Any hooks defined in the provided trace will be called first.
func (t *simpleTracer) WithContext(ctx context.Context) context.Context {
	return httptrace.WithClientTrace(ctx, t.ct)
}
