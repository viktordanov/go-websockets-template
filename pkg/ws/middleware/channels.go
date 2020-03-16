package middleware

import (
	"context"

	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware/priority"
)

type Channels struct {
	priority priority.Event
	name     string
}

func NewChannels(priority priority.Event, name string) *Channels {
	return &Channels{priority: priority, name: name}
}

func (m *Channels) Process(ctx context.Context, msg Message) (context.Context, Message, error) {

	return ctx, msg, nil
}

func (m *Channels) GetPriority() priority.Event {
	return m.priority
}
func (m *Channels) GetName() string {
	return m.name
}
