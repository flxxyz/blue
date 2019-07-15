package blue

import (
	"bytes"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	emptyMessage = bytes.NewBuffer([]byte{})
)

type Client struct {
	ID      *bytes.Buffer   //当前连接的唯一id
	conn    *websocket.Conn //websocket连接
	Request *http.Request   //保存当前请求的一些信息，貌似也没什么用
	packet  chan *Msg       //消息负载
	locker  *sync.RWMutex   //读写锁
	heart   *Heart
	exit    bool
}

func (c *Client) writeMessage(m *Msg) {
	if c.closed() {
		c.heart.errorHandler(c, errors.New("写入到一个已关闭的会话"))
		return
	}

	select {
	case c.packet <- m:
	default:
		c.heart.errorHandler(c, errors.New("packet的buffer已经写满了"))
	}
}

func (c *Client) writeRaw(m *Msg) error {
	if c.closed() {
		return errors.New("写入到一个已关闭的会话")
	}

	_ = c.conn.SetWriteDeadline(time.Now().Add(c.heart.conf.WriteDeadline))
	err := c.conn.WriteMessage(m.T, m.Body.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) readPump() {
	c.conn.SetReadLimit(c.heart.conf.Message.MaxBufferSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(c.heart.conf.ReadDeadline))

	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(c.heart.conf.ReadDeadline))
		c.heart.pongHandler(c)
		return nil
	})

	if c.heart.closeHandler != nil {
		c.conn.SetCloseHandler(func(code int, text string) error {
			return c.heart.closeHandler(c, code, text)
		})
	}

	for {
		t, message, err := c.conn.ReadMessage()
		if err != nil {
			//Todo: 客户端退出必有code:1001的error
			c.heart.errorHandler(c, err)
			return
		}

		if t == textMessage {
			c.heart.messageHandler(c, bytes.NewBuffer(message))
		}

		if t == binaryMessage {
			c.heart.messageBinaryHandler(c, bytes.NewBuffer(message))
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker((c.heart.conf.ReadDeadline * 9) / 10)
	defer ticker.Stop()

	for {
		select {
		case m, ok := <-c.packet:
			if !ok {
				_ = c.conn.WriteMessage(closeMessage, emptyMessage.Bytes())
				c.heart.disconnectHandler(c)
				return
			}

			err := c.writeRaw(m)
			if err != nil {
				c.heart.errorHandler(c, err)
				return
			}

			if m.T == closeMessage {
				return
			}

			if m.T == textMessage {
				c.heart.messageSentHandler(c, m.Body)
			}

			if m.T == binaryMessage {
				c.heart.messageBinarySentHandler(c, m.Body)
			}
		case <-ticker.C:
			_ = c.writeRaw(NewMsg(pingMessage, emptyMessage, nil))
		}
	}
}

func (c *Client) Write(data *bytes.Buffer) error {
	if c.closed() {
		return errors.New("会话已经关闭了")
	}

	c.writeMessage(NewMsg(
		textMessage,
		data,
		nil,
	))

	return nil
}

func (c *Client) WriteBinary(data *bytes.Buffer) error {
	if c.closed() {
		return errors.New("会话已经关闭了")
	}

	c.writeMessage(NewMsg(
		binaryMessage,
		data,
		nil,
	))

	return nil
}

func (c *Client) Close() error {
	if c.closed() {
		return errors.New("会话已经关闭了")
	}

	c.writeMessage(NewMsg(
		closeMessage,
		emptyMessage,
		nil,
	))

	return nil
}

func (c *Client) closed() bool {
	c.locker.RLock()
	defer c.locker.RUnlock()

	return c.exit
}

func (c *Client) close() {
	c.locker.Lock()
	if !c.closed() {
		c.exit = true
		c.conn.Close()
		close(c.packet)
		delete(c.heart.clients, c)
		log.Println("client close() 触发成功")
	}
	log.Println("client close() 触发失败")
	c.locker.Unlock()
}
