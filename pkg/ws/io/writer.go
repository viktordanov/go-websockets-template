package io

import (
	"context"

	"nhooyr.io/websocket"
)

const (
	PART_ONE   = "{\"channel\":\""
	PART_TWO   = "\",\"message\":\""
	PART_THREE = "\"}"
)

func Writer(ctx context.Context, c *websocket.Conn, channel, message string) error {
	s := make([]byte, 0, 27+len(channel)+len(message))
	s = append(s, PART_ONE...)
	s = append(s, channel...)
	s = append(s, PART_TWO...)
	s = append(s, message...)
	s = append(s, PART_THREE...)
	c.Write(ctx, websocket.MessageText, s)
	return nil
}
