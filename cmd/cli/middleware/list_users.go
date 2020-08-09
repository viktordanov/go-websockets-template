package middleware

import (
	"context"
	"errors"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/io"
	mdw "github.com/viktordanov/go-youtube-sync/pkg/ws/middleware"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/middleware/priority"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/model"
	"nhooyr.io/websocket"
)

// type Middleware interface {
// 	Process(Message) (Message, error)
// 	GetPriority() priority.Event
// 	GetName() string
// }

type ListUsers struct {
	priority priority.Event
	name     string
}

func NewListUsers(priority priority.Event, name string) *ListUsers {
	return &ListUsers{priority: priority, name: name}
}

func (m *ListUsers) Process(ctx context.Context, msg mdw.Message) (context.Context, mdw.Message, error) {
	c, ok := ctx.Value(mdw.ConnectionPtrKey()).(*websocket.Conn)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve connection pointer from context")
	}

	users, ok := ctx.Value(mdw.UsersKey()).(map[string]*model.User)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve users from context")
	}

	// Filter handled channels
	found := false

	for _, channel := range m.HandledChannels() {
		if channel == msg.Channel {
			found = true
			break
		}
	}

	if !found {
		return ctx, msg, nil
	}

	usersString := ""
	userCount := len(users)
	i := 0
	for name := range users {
		usersString += name
		if i < userCount-1 {
			usersString += ", "
		}
		i++
	}

	io.Writer(ctx, c, "list-users-reply", usersString)
	return ctx, msg, nil
}

func (m *ListUsers) GetPriority() priority.Event {
	return m.priority
}

func (m *ListUsers) GetName() string {
	return m.name
}

func (m *ListUsers) HandledChannels() []string {
	return []string{"list-users"}
}
