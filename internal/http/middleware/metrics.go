package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Dragodui/diploma-server/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track in-flight requests
		metrics.HttpRequestsInFlight.Inc()
		defer metrics.HttpRequestsInFlight.Dec()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)

		// Record request count
		metrics.HttpRequestsTotal.WithLabelValues(
			r.Method,
			r.URL.Path,
			status,
		).Inc()

		// Record request duration
		metrics.HttpRequestDuration.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(duration)

		// Record response size
		metrics.HttpResponseSizeBytes.WithLabelValues(
			r.Method,
			r.URL.Path,
		).Observe(float64(wrapped.bytesWritten))
	})
}
