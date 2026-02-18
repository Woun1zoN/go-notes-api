package middleware

import (
    "net/http"
	"time"
	"log"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		log.Printf(
			"[%s] | {%s} | [Status: %d] %v | [IP:PORT - %s]\nUser Agent: %s", r.Method, r.URL.Path, rw.status, duration, r.RemoteAddr, r.UserAgent(),
		)
	})
}