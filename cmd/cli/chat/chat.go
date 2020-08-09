package chat

import (
	"context"
	jsoniter "github.com/json-iterator/go"
	"github.com/viktordanov/go-youtube-sync/pkg/ws/io"
	"nhooyr.io/websocket"
)

type ChatMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func SendChatMessage(ctx context.Context, c *websocket.Conn, msgType, msg string) {
	chatMessage := ChatMessage{
		Type:    msgType,
		Message: msg,
	}
	msgJSON, _ := jsoniter.Marshal(chatMessage)
	io.WriterJSON(ctx, c, "chat", string(msgJSON))
}
