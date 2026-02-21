package middleware

import (
	"net/http"
	"context"

	"github.com/google/uuid"
)

type requestKey string

const requestIDKey requestKey = "requestID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("Request-ID")
		if _, err := uuid.Parse(id); err != nil {
            id = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		r = r.WithContext(ctx)

		w.Header().Set("Request-ID", id)

		next.ServeHTTP(w, r)
	})
}

func GetRequestID(r *http.Request) string {
	id, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		return "Unknown"
	}
	return id
}