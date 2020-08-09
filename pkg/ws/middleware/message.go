package middleware

type Message struct {
	Channel   string `json:"channel"`
	Message   string `json:"message"`
	Metadata  string `json:"metadata"`
	Metadata2 string `json:"metadata2"`
}
