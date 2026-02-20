package middleware

import (
    "net/http"
	"log"
	"runtime/debug"
)

func MiddlewareRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			stack := debug.Stack()

			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n[%s] | {%s} | [IP:PORT - %s]\nStack:\n%s\n", err, r.Method, r.URL.String(), r.RemoteAddr, stack)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}