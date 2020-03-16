package middleware

import (
	"context"
	"errors"
	"log"

	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware/priority"
)

// type Middleware interface {
// 	Process(Message) (Message, error)
// 	GetPriority() priority.Event
// 	GetName() string
// }

type Print struct {
	priority priority.Event
	name     string
}

func NewPrint(priority priority.Event, name string) *Print {
	return &Print{priority: priority, name: name}
}

func (m *Print) Process(ctx context.Context, msg Message) (context.Context, Message, error) {
	id, ok := ctx.Value(sessionIDKey).(int64)
	if !ok {
		return ctx, msg, errors.New("Unexpected value cast error")
	}
	log.Printf("%d %s\n", id, msg.Message)
	return ctx, msg, nil
}

func (m *Print) GetPriority() priority.Event {
	return m.priority
}
func (m *Print) GetName() string {
	return m.name
}
