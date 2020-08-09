package middleware

import (
	"context"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/model"
	"math/rand"
	"net/http"

	"nhooyr.io/websocket"
)

type contextKey int8

const sessionIDKey = contextKey(0)
const connectionPtrKey = contextKey(1)
const usersKey = contextKey(2)
const roomsKey = contextKey(3)

func SessionIDKey() contextKey {
	return sessionIDKey
}

func ConnectionPtrKey() contextKey {
	return connectionPtrKey
}

func UsersKey() contextKey {
	return usersKey
}
func RoomsKey() contextKey {
	return roomsKey
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

func ContextWithUsers(ctx context.Context, users map[string]*model.User) context.Context {
	return context.WithValue(ctx, usersKey, users)
}

func ContextWithRooms(ctx context.Context, rooms map[string]*model.Room) context.Context {
	return context.WithValue(ctx, roomsKey, rooms)
}
