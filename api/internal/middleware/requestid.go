package middleware

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

// Key to use when setting the request ID.
type ctxKeyRequestID int

// RequestIDKey is the key that holds the unique request ID in a request context.
const RequestIDKey ctxKeyRequestID = 0

//RequestIDHeaderName is the name of the header used to signal and set the incoming request ID if it's missing.
const RequestIDHeaderName = "x-devbot-request-id"

// GetRequestID provides the request ID of the request associated with the given context.
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

//RequestID creates an HTTP middleware for go-chi web framework that ensures a request ID is set on requests.
//goland:noinspection GoUnusedExportedFunction
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeaderName)
		if id == "" {
			id = uuid.New().String()
		}
		w.Header().Set(RequestIDHeaderName, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), RequestIDKey, id)))
	})
}
