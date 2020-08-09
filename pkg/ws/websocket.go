package ws

import (
	"context"
	"github.com/viktordanov/go-youtube-sync/cmd/cli/chat"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/model"
	"log"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/io"
	m "github.com/viktordanov/go-youtube-sync/pkg/ws/middleware"
	p "github.com/viktordanov/go-youtube-sync/pkg/ws/middleware/priority"
	"nhooyr.io/websocket"
)

type WSServer struct {
	AcceptOptions   *websocket.AcceptOptions
	middlewares     map[p.Event][]m.Middleware
	initMiddlewares map[p.Event][]m.Middleware
	users           map[string]*model.User
	rooms           map[string]*model.Room
}

func NewWSServer() *WSServer {
	options := &websocket.AcceptOptions{InsecureSkipVerify: true}
	return &WSServer{
		AcceptOptions:   options,
		middlewares:     make(map[p.Event][]m.Middleware),
		initMiddlewares: make(map[p.Event][]m.Middleware),
		users:           make(map[string]*model.User),
		rooms:           make(map[string]*model.Room),
	}
}

func (wss *WSServer) AddMiddleware(m m.Middleware) {
	wss.middlewares[m.GetPriority()] = append(wss.middlewares[m.GetPriority()], m)
}

func (wss *WSServer) AddInitMiddleware(m m.Middleware) {
	wss.initMiddlewares[m.GetPriority()] = append(wss.initMiddlewares[m.GetPriority()], m)
}

func (wss *WSServer) Serve() error {
	return http.ListenAndServe("127.0.0.1:9999", m.WithSessionID(wss.wsHandler))
}

func (wss *WSServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch generated session ID from context
	sessionID, ok := r.Context().Value(m.SessionIDKey()).(int64)
	if !ok {
		panic("Impossible edge case")
	}

	c, err := websocket.Accept(w, r, wss.AcceptOptions)
	if err != nil {
		log.Printf("Session %d terminated. %v\n", sessionID, err)
		return
	}
	log.SetPrefix("<~> ")
	log.Printf("Session %d initiated.\n", sessionID)
	log.SetPrefix("< > ")

	defer c.Close(websocket.StatusInternalError, "Internal error")

	var v m.Message

	failedAttempts := 0
	for {
		if failedAttempts == 3 {
			c.Close(websocket.StatusNormalClosure, "Maximum attempts exceeded")
			return
		}

		raw, err := io.Reader(r.Context(), c)
		if err != nil {
			c.Close(websocket.StatusNormalClosure, "")
			return
		}
		err = jsoniter.Unmarshal(raw, &v)
		if err != nil {
			log.Printf("Received unsupported message format\n%#v\n", err)
			c.Close(websocket.StatusNormalClosure, "Received unsupported message format")
			return
		}
		if v.Channel != "init" {
			log.Printf("Received incorrect init frame\n")
			io.Writer(r.Context(), c, "init-error", "Incorrect init frame")
			failedAttempts++
			continue
		}

		if _, ok := wss.users[v.Message]; ok {
			log.Printf("Username already used\n")
			io.Writer(r.Context(), c, "init-error", "Username used")
			failedAttempts++
			continue
		}

		io.Writer(r.Context(), c, "init-reply", v.Message)
		wss.users[v.Message] = model.NewUser(v.Message, c)
		break
	}

	timeBegin := time.Now()
	for {
		raw, err := io.Reader(r.Context(), c)
		if err != nil {
			log.SetPrefix("<!> ")
			log.Printf("Session %d terminated by client after %v.\n", sessionID, time.Since(timeBegin))
			log.SetPrefix("< > ")

			break
		}
		var v m.Message
		err = jsoniter.Unmarshal(raw, &v)
		if err != nil {
			log.Printf("Received unsupported message format\n%#v\n", err)
			continue
		}

		ctx, cancel := context.WithCancel(
			m.ContextWithUsers(
				m.ContextWithRooms(
					m.ContextWithConnection(
						r.Context(), c), wss.rooms), wss.users))

		for priority := p.LOW; priority <= p.HIGH; priority++ {
			wares := wss.initMiddlewares[priority]
			for _, ware := range wares {
				ctx, v, err = ware.Process(ctx, v)
				if err != nil {
					log.Printf("Middleware \"%s\" :: \"%v\"\n", ware.GetName(), err)
					cancel()
				}
			}
		}

		for priority := p.LOW; priority <= p.HIGH; priority++ {
			wares := wss.middlewares[priority]
			for _, ware := range wares {
				ctx, v, err = ware.Process(ctx, v)

				if err != nil {
					log.Printf("Middleware \"%s\" :: \"%v\"\n", ware.GetName(), err)
					cancel()
				}
			}
		}

		time.Sleep(time.Millisecond * 4)
		cancel()
		select {
		case <-ctx.Done():
			break
		}
	}
	for _, room := range wss.rooms {
		if room.Owner == v.Message {
			room.NextOwner()
			if len(room.Users) > 0 {
				chat.SendChatMessage(r.Context(), room.Users[room.Owner].Connection, "info", "You are the owner of the room")
			}
		}
		delete(room.Users, v.Message)
		for _, user := range room.Users {
			chat.SendChatMessage(r.Context(), user.Connection, "info", v.Message+" left")
		}
	}

	delete(wss.users, v.Message)

	c.Close(websocket.StatusNormalClosure, "")
}
