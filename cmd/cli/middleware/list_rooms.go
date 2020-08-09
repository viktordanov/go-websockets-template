package middleware

import (
	"context"
	"errors"
	jsoniter "github.com/json-iterator/go"
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

type gatewayRoom struct {
	Owner string   `json:"owner"`
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

type roomsReply struct {
	Rooms []gatewayRoom `json:"rooms"`
}

type ListRooms struct {
	priority priority.Event
	name     string
}

func NewListRooms(priority priority.Event, name string) *ListRooms {
	return &ListRooms{priority: priority, name: name}
}

func (m *ListRooms) Process(ctx context.Context, msg mdw.Message) (context.Context, mdw.Message, error) {
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

	c, ok := ctx.Value(mdw.ConnectionPtrKey()).(*websocket.Conn)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve connection pointer from context")
	}

	rooms, ok := ctx.Value(mdw.RoomsKey()).(map[string]*model.Room)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve rooms from context")
	}

	reply := roomsReply{Rooms: []gatewayRoom{}}

	for _, room := range rooms {
		reply.Rooms = append(reply.Rooms, gatewayRoom{
			Owner: room.Owner,
			Name:  room.Name,
			Users: room.UserNames(),
		})
	}

	replyJson, err := jsoniter.Marshal(reply)
	if err != nil {
		return ctx, msg, err
	}

	io.WriterJSON(ctx, c, "list-rooms-reply", string(replyJson))
	return ctx, msg, nil
}

func (m *ListRooms) GetPriority() priority.Event {
	return m.priority
}

func (m *ListRooms) GetName() string {
	return m.name
}

func (m *ListRooms) HandledChannels() []string {
	return []string{"list-rooms"}
}
