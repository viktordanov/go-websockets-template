package test

import (
	"bytes"
	"strings"
	"testing"
	"unsafe"
)

const (
	PART_ONE   = "{\"channel\":\""
	PART_TWO   = "\",\"message\":\""
	PART_THREE = "\"}"
	channel    = "test"
	message    = "A very fucking long test message which will never be used in practice"
)

func BenchmarkWriterBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]byte, 0, 27+len(channel)+len(message))
		s = append(s, PART_ONE...)
		s = append(s, channel...)
		s = append(s, PART_TWO...)
		s = append(s, message...)
		s = append(s, PART_THREE...)
	}
}

func BenchmarkWriterStringBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s strings.Builder
		s.Grow(16 + int(unsafe.Sizeof(channel)+unsafe.Sizeof(message)))
		s.Write([]byte(PART_ONE))
		s.Write([]byte(channel))
		s.Write([]byte(PART_TWO))
		s.Write([]byte(message))
		s.Write([]byte(PART_THREE))
		_ = []byte(s.String())
	}
}

func BenchmarkWriterStrings(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s string
		s += PART_ONE
		s += channel
		s += PART_TWO
		s += message
		s += PART_THREE
	}
}

func BenchmarkWriterByteBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buffer bytes.Buffer
		buffer.Grow(16 + int(unsafe.Sizeof(channel)+unsafe.Sizeof(message)))
		buffer.Write([]byte(PART_ONE))
		buffer.Write([]byte(channel))
		buffer.Write([]byte(PART_TWO))
		buffer.Write([]byte(message))
		buffer.Write([]byte(PART_THREE))
		buffer.Bytes()
	}
}
