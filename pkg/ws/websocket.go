package ws

import (
	"context"
	"log"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/io"
	m "github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware"
	p "github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware/priority"
	"nhooyr.io/websocket"
)

type WSServer struct {
	AcceptOptions *websocket.AcceptOptions
	middlewares   map[p.Event][]m.Middleware
}

func NewWSServer() *WSServer {
	options := &websocket.AcceptOptions{InsecureSkipVerify: true}
	return &WSServer{AcceptOptions: options, middlewares: make(map[p.Event][]m.Middleware)}
}
func (wss *WSServer) AddMiddleware(m m.Middleware) {
	wss.middlewares[m.GetPriority()] = append(wss.middlewares[m.GetPriority()], m)
}

func (wss *WSServer) Serve() error {
	return http.ListenAndServe(":8080", m.WithSessionID(wss.wsHandler))
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

		ctx, cancel := context.WithCancel(m.ContextWithConnection(r.Context(), c))

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
	c.Close(websocket.StatusNormalClosure, "")

}
