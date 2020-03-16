package middleware

import (
	"context"
	"math/rand"
	"net/http"

	"nhooyr.io/websocket"
)

type contextKey int8

const sessionIDKey = contextKey(0)
const connectionPtrKey = contextKey(1)

func SessionIDKey() contextKey {
	return sessionIDKey
}

func ConnectionPtrKey() contextKey {
	return connectionPtrKey
}

func WithSessionID(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, sessionIDKey, rand.Int63())
		f(w, r.WithContext(ctx))
	}
}

func ContextWithConnection(ctx context.Context, c *websocket.Conn) context.Context {
	return context.WithValue(ctx, connectionPtrKey, c)
}
