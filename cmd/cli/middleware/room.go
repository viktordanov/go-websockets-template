package middleware

import (
	"context"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/viktordanov/go-youtube-sync/cmd/cli/chat"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/io"
	mdw "github.com/viktordanov/go-youtube-sync/pkg/ws/middleware"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/middleware/priority"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/model"
	"nhooyr.io/websocket"
	"strconv"
)

type gatewayPlaylist struct {
	Playlist []string `json:"playlist"`
	Current  int      `json:"current"`
}

type Room struct {
	priority priority.Event
	name     string
}

func NewRoom(priority priority.Event, name string) *Room {
	return &Room{priority: priority, name: name}
}

func (m *Room) Process(ctx context.Context, msg mdw.Message) (context.Context, mdw.Message, error) {
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

	users, ok := ctx.Value(mdw.UsersKey()).(map[string]*model.User)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve rooms from context")
	}

	rooms, ok := ctx.Value(mdw.RoomsKey()).(map[string]*model.Room)
	if !ok {
		return ctx, msg, errors.New("Couldn't retrieve rooms from context")
	}

	switch msg.Channel {
	case "add-room":
		return m.HandleAddRoom(ctx, msg, c, users, rooms)
	case "remove-room":
		return m.HandleRemoveRoom(ctx, msg, c, users, rooms)
	case "join-room":
		return m.HandleJoinRoom(ctx, msg, c, users, rooms)
	case "leave-room":
		return m.HandleLeaveRoom(ctx, msg, c, users, rooms)
	case "command":
		return m.HandleCommand(ctx, msg, c, users, rooms)
	case "get-url":
		return m.HandleGetURL(ctx, msg, c, users, rooms)
	case "get-playlist":
		return m.HandleGetPlaylist(ctx, msg, c, users, rooms)
	}

	return ctx, msg, nil
}

func (m *Room) GetPriority() priority.Event {
	return m.priority
}

func (m *Room) GetName() string {
	return m.name
}

func (m *Room) HandledChannels() []string {
	return []string{"add-room", "remove-room", "join-room", "leave-room", "command", "get-url", "get-playlist"}
}

func (m *Room) HandleAddRoom(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	roomName := msg.Message
	if _, ok := rooms[roomName]; ok {
		chat.SendChatMessage(ctx, c, "error", "A room with the same name exists!")
		return ctx, msg, nil
	}

	var username string
	for s, user := range users {
		if user.Connection == c {
			username = s
		}
	}

	if len(roomName) < 3 {
		chat.SendChatMessage(ctx, c, "error", "Room name too short!")
		return ctx, msg, nil
	}

	rooms[roomName] = model.NewRoom(username, roomName, "")
	chat.SendChatMessage(ctx, c, "info", "Created room "+roomName)

	for _, user := range users {
		BroadcastRooms(ctx, user.Connection, rooms)
	}

	return ctx, msg, nil
}

func (m *Room) HandleRemoveRoom(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	roomName := msg.Message
	if _, ok := rooms[roomName]; !ok {
		chat.SendChatMessage(ctx, c, "error", "Room doesn't exist!")
		return ctx, msg, nil
	}

	var username string
	for s, user := range users {
		if user.Connection == c {
			username = s
		}
	}

	if username != rooms[roomName].Owner {
		chat.SendChatMessage(ctx, c, "error", "Unauthorized")
		return ctx, msg, nil
	}

	delete(rooms, roomName)
	chat.SendChatMessage(ctx, c, "info", "Deleted room "+roomName)

	for _, user := range users {
		BroadcastRooms(ctx, user.Connection, rooms)
	}
	return ctx, msg, nil
}

func (m *Room) HandleJoinRoom(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	roomName := msg.Message

	if _, ok := rooms[roomName]; !ok {
		chat.SendChatMessage(ctx, c, "error", "Room doesn't exist!")
		return ctx, msg, nil
	}
	var userName string
	for s, user := range users {
		if user.Connection == c {
			userName = s
		}
	}

	if _, ok := rooms[roomName].Users[userName]; ok {
		chat.SendChatMessage(ctx, c, "error", "Already in room")
		return ctx, msg, nil
	}

	for _, user := range rooms[roomName].Users {
		chat.SendChatMessage(ctx, user.Connection, "info", userName+" joined")
	}

	for _, room := range rooms {
		for _, user := range room.Users {
			if user.Name == userName {
				delete(room.Users, userName)
				chat.SendChatMessage(ctx, c, "info", "Left room "+room.Name)

				if userName == room.Owner {
					room.NextOwner()
					if len(room.Users) > 0 {
						chat.SendChatMessage(ctx, room.Users[room.Owner].Connection, "info", "You are the owner of the room")
					}
				}
			}
		}
	}

	if len(rooms[roomName].Users) == 0 {
		rooms[roomName].NextOwner()
		if len(rooms[roomName].Users) > 0 {
			chat.SendChatMessage(ctx, rooms[roomName].Users[rooms[roomName].Owner].Connection, "info", "You are the owner of the room")
		}
	}
	rooms[roomName].Users[userName] = users[userName]
	chat.SendChatMessage(ctx, c, "info", "Joined room "+roomName)
	io.Writer(ctx, c, "room-join", roomName)

	for _, user := range users {
		BroadcastRooms(ctx, user.Connection, rooms)
	}
	return ctx, msg, nil
}

func (m *Room) HandleLeaveRoom(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	roomName := msg.Message

	var userName string
	for s, user := range users {
		if user.Connection == c {
			userName = s
		}
	}

	if _, ok := rooms[roomName]; !ok {
		chat.SendChatMessage(ctx, c, "error", "Room doesn't exist!")
		return ctx, msg, nil
	}
	if _, ok := rooms[roomName].Users[userName]; !ok {
		chat.SendChatMessage(ctx, c, "error", "User not in room!")
		return ctx, msg, nil
	}
	delete(rooms[roomName].Users, userName)
	chat.SendChatMessage(ctx, c, "info", "Left room "+roomName)

	if userName == rooms[roomName].Owner {
		if len(rooms[roomName].Users) > 1 {
			rooms[roomName].NextOwner()
			chat.SendChatMessage(ctx, rooms[roomName].Users[rooms[roomName].Owner].Connection, "info", "You are the owner of the room")
		}
	}

	for _, user := range rooms[roomName].Users {
		chat.SendChatMessage(ctx, user.Connection, "info", userName+" left")
	}

	io.Writer(ctx, c, "room-leave", roomName)
	for _, user := range users {
		BroadcastRooms(ctx, user.Connection, rooms)
	}
	return ctx, msg, nil
}

func (m *Room) HandleCommand(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	command := msg.Metadata
	content := msg.Message

	var user *model.User
	var room *model.Room
	for _, u := range users {
		if u.Connection == c {
			user = u
			break
		}
	}

	if user == nil {
		chat.SendChatMessage(ctx, c, "error", "User not found!")
		return ctx, msg, nil
	}
	for _, r := range rooms {
		found := false
		for _, u := range r.Users {
			if u.Name == user.Name {
				found = true
				room = r
				break
			}
		}
		if found {
			break
		}
	}

	if room == nil {
		chat.SendChatMessage(ctx, c, "error", "User room not found!")
		return ctx, msg, nil
	}

	if command == "add-play-url" {
		room.AddToPlaylist(content)
		room.GoToLast()
		BroadcastPlaylist(ctx, room)
	}

	if command == "add-url" {
		room.AddToPlaylist(content)
		BroadcastPlaylist(ctx, room)
	}

	if command == "remove-url" {
		room.RemoveFromPlaylist(content)
		BroadcastPlaylist(ctx, room)
	}

	if command == "next" {
		content = room.NextUrl()
		BroadcastPlaylist(ctx, room)
	}
	if command == "auto-next" {
		if user.Name != room.Owner {
			return ctx, msg, errors.New("User not owner")
		}
		command = "next"
		if room.CurrentIndex+1 < len(room.Playlist) {
			content = room.NextUrl()
			BroadcastPlaylist(ctx, room)
		}
	}

	if command == "prev" {
		content = room.PrevUrl()
		BroadcastPlaylist(ctx, room)
	}

	if command == "set-index" {
		index, err := strconv.Atoi(content)
		if err != nil {
			return ctx, msg, errors.New("Couldn't convert index to int")
		}
		room.SetCurrentIndex(index)
		content = room.CurrentUrl()
		BroadcastPlaylist(ctx, room)
	}

	for _, user := range room.Users {
		io.WriterFull(ctx, user.Connection, msg.Channel, content, command, msg.Metadata2)
	}

	return ctx, msg, nil
}

func (m *Room) HandleGetURL(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	var user *model.User
	var room *model.Room
	for _, u := range users {
		if u.Connection == c {
			user = u
		}
	}
	if user == nil {
		chat.SendChatMessage(ctx, c, "error", "User not found!")
		return ctx, msg, nil
	}
	for _, r := range rooms {
		found := false
		for _, u := range r.Users {
			if u.Name == user.Name {
				found = true
				room = r
				break
			}
		}
		if found {
			break
		}
	}
	if room == nil {
		chat.SendChatMessage(ctx, c, "error", "User room not found!")
		return ctx, msg, nil
	}

	io.Writer(ctx, c, msg.Channel, room.CurrentUrl())

	return ctx, msg, nil
}

func (m *Room) HandleGetPlaylist(ctx context.Context, msg mdw.Message, c *websocket.Conn, users map[string]*model.User, rooms map[string]*model.Room) (context.Context, mdw.Message, error) {
	var user *model.User
	var room *model.Room
	for _, u := range users {
		if u.Connection == c {
			user = u
		}
	}
	if user == nil {
		chat.SendChatMessage(ctx, c, "error", "User not found!")
		return ctx, msg, nil
	}
	for _, r := range rooms {
		found := false
		for _, u := range r.Users {
			if u.Name == user.Name {
				found = true
				room = r
				break
			}
		}
		if found {
			break
		}
	}
	if room == nil {
		chat.SendChatMessage(ctx, c, "error", "User room not found!")
		return ctx, msg, nil
	}

	message := gatewayPlaylist{
		Playlist: room.Playlist,
		Current:  room.CurrentIndex,
	}

	replyJson, _ := jsoniter.Marshal(message)
	replyStr := string(replyJson)

	io.WriterJSON(ctx, user.Connection, msg.Channel, replyStr)

	return ctx, msg, nil
}

func BroadcastRooms(ctx context.Context, c *websocket.Conn, rooms map[string]*model.Room) {
	reply := roomsReply{Rooms: []gatewayRoom{}}

	for _, room := range rooms {
		reply.Rooms = append(reply.Rooms, gatewayRoom{
			Owner: room.Owner,
			Name:  room.Name,
			Users: room.UserNames(),
		})
	}

	replyJson, _ := jsoniter.Marshal(reply)

	io.WriterJSON(ctx, c, "list-rooms-reply", string(replyJson))
}

func BroadcastPlaylist(ctx context.Context, r *model.Room) {
	message := gatewayPlaylist{
		Playlist: r.Playlist,
		Current:  r.CurrentIndex,
	}

	replyJson, _ := jsoniter.Marshal(message)
	replyStr := string(replyJson)
	for _, user := range r.Users {
		io.WriterJSON(ctx, user.Connection, "get-playlist", replyStr)
	}
}
