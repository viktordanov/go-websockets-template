package middleware

import (
	"context"
	"errors"
	mdw "github.com/viktordanov/go-youtube-sync/pkg/ws/middleware"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/model"

	"github.com/viktordanov/go-youtube-sync/pkg/ws/io"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/middleware/priority"
)

// type Middleware interface {
// 	Process(Message) (Message, error)
// 	GetPriority() priority.Event
// 	GetName() string
// }

type Chat struct {
	priority priority.Event
	name     string
}

func NewChat(priority priority.Event, name string) *Chat {
	return &Chat{priority: priority, name: name}
}

func (m *Chat) Process(ctx context.Context, msg mdw.Message) (context.Context, mdw.Message, error) {
	if msg.Channel != "chat" {
		return ctx, msg, nil
	}
	users, ok := ctx.Value(mdw.UsersKey()).(map[string]*model.User)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve users from context")
	}

	for _, user := range users {
		io.WriterMeta(ctx, user.Connection, "chat", msg.Message, msg.Metadata)
	}
	return ctx, msg, nil
}

func (m *Chat) GetPriority() priority.Event {
	return m.priority
}
func (m *Chat) GetName() string {
	return m.name
}
