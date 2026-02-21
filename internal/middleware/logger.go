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

		rw := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		id := GetRequestID(r)

		log.Printf(
			"ID Request: %s\n\n[%s] | {%s} | [Status: %d] %v | [IP:PORT - %s]\nUser Agent: %s\n",
			id, r.Method, r.URL.Path, rw.status, duration, r.RemoteAddr, r.UserAgent(),
		)
	})
}