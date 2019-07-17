package blue

import (
	"bytes"
	"errors"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"log"
	"net/http"
	"sync"
)

type (
	clientHandleFunc  func(c *Client)
	messageHandleFunc func(c *Client, data *bytes.Buffer)
	errorHandleFunc   func(c *Client, err error)
	closeHandleFunc   func(c *Client, code int, text string) error
)

//心脏
type Heart struct {
	*ClientManage
	upgrader *websocket.Upgrader
	conf     *Conf

	// 这都是不开放的
	connectHandler           clientHandleFunc
	disconnectHandler        clientHandleFunc
	messageHandler           messageHandleFunc
	messageBinaryHandler     messageHandleFunc
	messageSentHandler       messageHandleFunc
	messageBinarySentHandler messageHandleFunc
	errorHandler             errorHandleFunc
	closeHandler             closeHandleFunc
	pongHandler              clientHandleFunc
}

//开放的处理链接函数
func (h *Heart) HandlerConnect(fn func(c *Client)) {
	h.connectHandler = fn
}

//开放的处理断开链接函数
func (h *Heart) HandlerDisconnect(fn func(c *Client)) {
	h.disconnectHandler = fn
}

//开放的处理消息函数
func (h *Heart) HandlerMessage(fn func(c *Client, data *bytes.Buffer)) {
	h.messageHandler = fn
}

//开放的处理二进制消息函数
func (h *Heart) HandlerMessageBinary(fn func(c *Client, data *bytes.Buffer)) {
	h.messageBinaryHandler = fn
}

//开放的处理发送消息函数
func (h *Heart) HandlerSendMessage(fn func(c *Client, data *bytes.Buffer)) {
	h.messageSentHandler = fn
}

//开放的处理发送二进制消息函数
func (h *Heart) HandlerSendMessageBinary(fn func(c *Client, data *bytes.Buffer)) {
	h.messageBinarySentHandler = fn
}

//开放的处理错误函数
func (h *Heart) HandlerError(fn func(c *Client, err error)) {
	h.errorHandler = fn
}

//开放的处理退出函数
func (h *Heart) HandlerClose(fn func(c *Client, code int, text string) error) {
	if fn != nil {
		h.closeHandler = fn
	}
}

//开放的处理pong函数
func (h *Heart) HandlerPong(fn func(c *Client)) {
	h.pongHandler = fn
}

//广播消息
func (h *Heart) Broadcast(data *bytes.Buffer) error {
	if h.closed() {
		return errors.New("manage已关闭")
	}

	h.broadcast <- NewMsg(textMessage, data, nil)

	return nil
}

//广播过滤的消息
func (h *Heart) BroadcastFilter(data *bytes.Buffer, fn func(*Client) bool) error {
	if h.closed() {
		return errors.New("manage已关闭")
	}

	h.broadcast <- NewMsg(textMessage, data, fn)

	return nil
}

//广播二进制消息
func (h *Heart) BroadcastBinary(data *bytes.Buffer) error {
	if h.closed() {
		return errors.New("manage已关闭")
	}

	h.broadcast <- NewMsg(binaryMessage, data, nil)

	return nil
}

//广播过滤的二进制消息
func (h *Heart) BroadcastBinaryFilter(data *bytes.Buffer, fn func(*Client) bool) error {
	if h.closed() {
		return errors.New("manage已关闭")
	}

	h.broadcast <- NewMsg(binaryMessage, data, fn)

	return nil
}

//处理请求升级
func (h *Heart) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		log.Println("启动失败！")
		return
	}

	clientId, _ := uuid.NewV4()
	c := &Client{
		ID:      bytes.NewBufferString(clientId.String()),
		Request: r,
		conn:    conn,
		heart:   h,
		packet:  make(chan *Msg, h.conf.Message.PacketBufferSize),
		locker:  &sync.RWMutex{},
		exit:    false,
	}

	h.pushRegister(c)

	h.connectHandler(c)

	go c.writePump()

	c.readPump()

	if !h.closed() {
		h.pushUnregister(c)
	}

	c.close()

	h.disconnectHandler(c)
}

//实例心脏
func NewHeart() *Heart {
	conf := NewConf()
	cm := NewClientManage()
	go cm.run()

	return &Heart{
		conf: conf,
		upgrader: &websocket.Upgrader{
			WriteBufferSize: conf.Message.WriteBufferSize,
			ReadBufferSize:  conf.Message.ReadBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				//允许其他域的操作
				return true
			},
		},
		ClientManage:             cm,
		errorHandler:             func(c *Client, err error) {},
		messageHandler:           func(c *Client, data *bytes.Buffer) {},
		messageBinaryHandler:     func(c *Client, data *bytes.Buffer) {},
		messageSentHandler:       func(c *Client, data *bytes.Buffer) {},
		messageBinarySentHandler: func(c *Client, data *bytes.Buffer) {},
		connectHandler:           func(c *Client) {},
		disconnectHandler:        func(c *Client) {},
		closeHandler:             nil,
		pongHandler:              func(c *Client) {},
	}
}
