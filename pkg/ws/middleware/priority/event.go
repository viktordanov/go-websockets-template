package priority

type Event byte

const (
	LOW Event = iota
	NORMAL
	HIGH
)
