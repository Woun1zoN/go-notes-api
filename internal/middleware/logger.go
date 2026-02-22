package middleware

import (
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := GetRequestID(r)

		log.Printf(
			"ID Request: %s | START | [%s] %s | [IP:PORT: %s]",
			id, r.Method, r.URL.Path, r.RemoteAddr,
		)

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Printf(
			"ID Request: %s | END   | [%s] %s | Status: %d | Duration: %v | [IP:PORT: %s]",
			id, r.Method, r.URL.Path, rw.status, duration, r.RemoteAddr,
		)
	})
}