package blue

import (
	"bytes"
	"github.com/gorilla/websocket"
)

const (
	textMessage   = websocket.TextMessage
	binaryMessage = websocket.BinaryMessage
	closeMessage  = websocket.CloseMessage
	pingMessage   = websocket.PingMessage
	pongMessage   = websocket.PongMessage
)

type filterFunc func(c *Client) bool

//消息结构体
type Msg struct {
	T      int
	Body   *bytes.Buffer
	Filter filterFunc
}

func NewMsg(messageType int, buffer *bytes.Buffer, fn filterFunc) *Msg {
	return &Msg{
		T:      messageType,
		Body:   buffer,
		Filter: fn,
	}
}
