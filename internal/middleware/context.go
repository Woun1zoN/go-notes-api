package middleware

import (
    "net/http"
	"context"
	"time"
)

func MiddlewareContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}