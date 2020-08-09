package middleware

import (
	"context"

	"github.com/viktordanov/go-youtube-sync/pkg/ws/middleware/priority"
)

type Middleware interface {
	Process(context.Context, Message) (context.Context, Message, error)
	GetPriority() priority.Event
	GetName() string
}

type ByPriority []Middleware

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].GetPriority() < a[j].GetPriority() }
