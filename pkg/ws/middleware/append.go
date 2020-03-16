package middleware

import (
	"context"

	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware/priority"
)

// type Middleware interface {
// 	Process(Message) (Message, error)
// 	GetPriority() priority.Event
// 	GetName() string
// }

type Append struct {
	priority priority.Event
	name     string
	toAppend string
}

func NewAppend(priority priority.Event, name, toAppend string) *Append {
	return &Append{priority: priority, name: name, toAppend: toAppend}
}

func (m *Append) Process(ctx context.Context, msg Message) (context.Context, Message, error) {
	msg.Message += m.toAppend
	return ctx, msg, nil
}

func (m *Append) GetPriority() priority.Event {
	return m.priority
}
func (m *Append) GetName() string {
	return m.name
}
