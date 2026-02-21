package middleware

import (
    "net/http"
	"log"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			stack := debug.Stack()
			id := GetRequestID(r)

			if err := recover(); err != nil {
				log.Printf("ID Request: %s\n\nPANIC: %v\n[%s] | {%s} | [IP:PORT - %s]\nStack:\n%s\n", 
				id, err, r.Method, r.URL.String(), r.RemoteAddr, stack)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}