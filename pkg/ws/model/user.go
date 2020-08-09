package model

import "nhooyr.io/websocket"

type User struct {
	Name       string `json:"name"`
	Ready      bool   `json:"ready"`
	Connection *websocket.Conn
}

func NewUser(name string, conn *websocket.Conn) *User {
	return &User{Name: name, Ready: false, Connection: conn}
}

func (u *User) SetReady(newState bool) {
	u.Ready = newState
}
