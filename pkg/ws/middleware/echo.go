package middleware

import (
	"context"
	"errors"

	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/io"
	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware/priority"
	"nhooyr.io/websocket"
)

// type Middleware interface {
// 	Process(Message) (Message, error)
// 	GetPriority() priority.Event
// 	GetName() string
// }

type Echo struct {
	priority priority.Event
	name     string
}

func NewEcho(priority priority.Event, name string) *Echo {
	return &Echo{priority: priority, name: name}
}

func (m *Echo) Process(ctx context.Context, msg Message) (context.Context, Message, error) {
	c, ok := ctx.Value(connectionPtrKey).(*websocket.Conn)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve connection pointer from context")
	}
	io.Writer(ctx, c, msg.Channel, msg.Message)
	return ctx, msg, nil
}

func (m *Echo) GetPriority() priority.Event {
	return m.priority
}
func (m *Echo) GetName() string {
	return m.name
}
