package main

import (
	"github.com/viktordanov/go-youtube-sync/cmd/cli/middleware"
	"log"
	"math/rand"
	"time"

	"github.com/viktordanov/go-youtube-sync/pkg/ws"
	p "github.com/viktordanov/go-youtube-sync/pkg/ws/middleware/priority"
)

func init() {
	rand.Seed(time.Now().UnixNano() + time.Now().UnixNano()/2)
	log.SetPrefix("< > ")
}

func main() {
	wss := ws.NewWSServer()
	wss.AddMiddleware(middleware.NewListUsers(p.NORMAL, "List users middleware"))
	wss.AddMiddleware(middleware.NewListRooms(p.NORMAL, "List rooms middleware"))
	wss.AddMiddleware(middleware.NewRoom(p.NORMAL, "Room command middleware"))
	wss.AddMiddleware(middleware.NewChat(p.NORMAL, "Chat middleware"))
	wss.Serve()
}
