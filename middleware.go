package common

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"time"
)

// requestIDKey is the context key used to store a request ID.
type requestIDKey struct{}

// WithRequestID returns a new context carrying the given request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID returns the request ID stored in the context, or "" when absent.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// responseRecorder wraps an http.ResponseWriter to capture the response
// status code and the number of bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int64
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	n, err := r.ResponseWriter.Write(p)
	r.bytes += int64(n)
	return n, err
}

// Flush passes through to the underlying writer when it supports
// http.Flusher, keeping SSE and other streaming responses working.
func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack passes through to the underlying writer when it supports
// http.Hijacker; otherwise it reports http.ErrNotSupported.
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// RequestLogger returns middleware that logs one line per request with the
// request ID, method, path, status code, bytes written, and duration. An
// incoming X-Request-ID header is honored; otherwise a random ID is generated.
// The ID is also injected into the request context, retrievable via
// RequestID. It is net/http-generic and works with any router (chi, stdlib
// mux, etc.).
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = newRequestID()
		}

		rw := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r.WithContext(WithRequestID(r.Context(), id)))

		LogInfo("req",
			"req_id", id,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"bytes", rw.bytes,
			"dur_ms", time.Since(start).Milliseconds(),
		)
	})
}

// newRequestID generates a random 8-byte hex request ID. crypto/rand.Read
// never returns an error and always fills the slice completely.
func newRequestID() string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
