package middleware

import (
	"net/http"
	"time"

	"github.com/Dragodui/diploma-server/internal/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func RequestResponseLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}
		next.ServeHTTP(rec, r)

		if r.URL.Path == "/" {
			return
		}

		duration := time.Since(start).Milliseconds()
		logger.Info.Printf(
			"%s %s -> %d (%dms)",
			r.Method,
			r.URL.RequestURI(),
			rec.status,
			duration,
		)
	})
}
