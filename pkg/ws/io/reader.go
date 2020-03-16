package io

import (
	"context"
	"fmt"

	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/internal/bpool"
	"github.com/viktordanov/itemIndexrWSSync/pkg/ws/internal/errd"
	"nhooyr.io/websocket"
)

func Reader(ctx context.Context, c *websocket.Conn) (_ []byte, err error) {
	defer errd.Wrap(&err, "failed to read JSON message")

	typ, r, err := c.Reader(ctx)
	if err != nil {
		return nil, err
	}

	if typ != websocket.MessageText {
		return nil, fmt.Errorf("expected text message for JSON but got: %v", typ)
	}

	b := bpool.Get()
	defer bpool.Put(b)

	_, err = b.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
