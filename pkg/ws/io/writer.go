package io

import (
	"context"
	"strings"

	"nhooyr.io/websocket"
)

const (
	PART_ONE       = "{\"channel\":\""
	PART_TWO       = "\",\"message\":\""
	PART_THREE     = "\"}"
	PART_THREE_ALT = "\",\"metadata\":\""
	PART_FOUR      = "\",\"metadata2\":\""
)

// Writer writes a reply to a connection
func Writer(ctx context.Context, c *websocket.Conn, channel, message string) error {
	s := make([]byte, 0, 27+len(channel)+len(message))
	s = append(s, PART_ONE...)
	s = append(s, channel...)
	s = append(s, PART_TWO...)
	s = append(s, message...)
	s = append(s, PART_THREE...)

	out, err := inflate(s)
	if err != nil {
		return err
	}
	c.Write(ctx, websocket.MessageText, out)
	return nil
}

// WriterJSON writes an escaped reply to a connection
func WriterJSON(ctx context.Context, c *websocket.Conn, channel, unescapedJSON string) error {
	return Writer(ctx, c, channel, escape(unescapedJSON))
}

// WriterJSONMeta writes an escaped reply to a connection
func WriterJSONMeta(ctx context.Context, c *websocket.Conn, channel, unescapedJSON, meta string) error {
	return WriterMeta(ctx, c, channel, escape(unescapedJSON), meta)
}

// WriterJSONFull writes an escaped reply to a connection
func WriterJSONFull(ctx context.Context, c *websocket.Conn, channel, message, meta, meta2 string) error {
	return WriterFull(ctx, c, channel, escape(message), escape(meta), escape(meta2))
}

// WriterMeta writes an escaped reply to a connection
func WriterMeta(ctx context.Context, c *websocket.Conn, channel, message, meta string) error {
	s := make([]byte, 0, 27+len(channel)+len(message))
	s = append(s, PART_ONE...)
	s = append(s, channel...)
	s = append(s, PART_TWO...)
	s = append(s, message...)
	s = append(s, PART_THREE_ALT...)
	s = append(s, meta...)
	s = append(s, PART_THREE...)

	out, err := inflate(s)
	if err != nil {
		return err
	}
	c.Write(ctx, websocket.MessageText, out)
	return nil
}

func WriterFull(ctx context.Context, c *websocket.Conn, channel, message, meta, meta2 string) error {
	s := make([]byte, 0, 27+len(channel)+len(message))
	s = append(s, PART_ONE...)
	s = append(s, channel...)
	s = append(s, PART_TWO...)
	s = append(s, message...)
	s = append(s, PART_THREE_ALT...)
	s = append(s, meta...)
	s = append(s, PART_FOUR...)
	s = append(s, meta2...)
	s = append(s, PART_THREE...)

	/*	out, err := inflate(s)
		if err != nil {
			return err
		}
	*/
	c.Write(ctx, websocket.MessageText, s)
	return nil
}

func inflate(data []byte) ([]byte, error) {
	/*sR := bytes.NewBuffer([]byte{})
	flateWriter, err := flate.NewWriter(sR, 2)
	if err != nil {
		return nil, err
	}
	flateWriter.Write(data)
	flateWriter.Close()
	return sR.Bytes(), nil*/
	return data, nil
}

func escape(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")
	str = strings.ReplaceAll(str, "\"", "\\\"")
	str = strings.ReplaceAll(str, "'", "\\'")
	str = strings.ReplaceAll(str, "/", "\\/")
	str = strings.ReplaceAll(str, "\r", "\\\r")
	str = strings.ReplaceAll(str, "\b", "\\\b")
	str = strings.ReplaceAll(str, "\f", "\\\f")
	str = strings.ReplaceAll(str, "\n", "\\\n")
	str = strings.ReplaceAll(str, "\t", "\\\t")
	return str
}