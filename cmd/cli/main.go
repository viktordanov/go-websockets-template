package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/viktordanov/itemIndexrWSSync/pkg/ws"
	m "github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware"
	p "github.com/viktordanov/itemIndexrWSSync/pkg/ws/middleware/priority"
)

func init() {
	rand.Seed(time.Now().UnixNano() + time.Now().UnixNano()/2)
	log.SetPrefix("< > ")
}

func main() {
	wss := ws.NewWSServer()
	wss.AddMiddleware(m.NewEcho(p.HIGH, "Echo middleware"))
	wss.Serve()
}
