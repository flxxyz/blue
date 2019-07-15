package blue_test

import (
	"bytes"
	"github.com/flxxyz/blue"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/quick"
	"time"
)

type TestServer struct {
	*blue.Heart
}

func NewTestServer(fn func(c *blue.Client, data *bytes.Buffer)) *TestServer {
	heart := blue.NewHeart()
	heart.HandlerMessage(fn)
	return &TestServer{
		heart,
	}
}

func NewDialer(url string) (*websocket.Conn, error) {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial(strings.Replace(url, "http", "ws", 1), nil)
	return conn, err
}

func TestHeart_HandlerConnect(t *testing.T) {
	testServer := NewTestServer(func(c *blue.Client, data *bytes.Buffer) {})
	server := httptest.NewServer(testServer)
	defer server.Close()

	fn := func(data string) bool {
		socket, err := NewDialer(server.URL)
		defer socket.Close()

		if err != nil {
			t.Error(err)
			return false
		}

		socket.WriteMessage(websocket.TextMessage, []byte(data))
		testServer.HandlerConnect(func(c *blue.Client) {
			buffer := bytes.NewBufferString("AMD, yes!")
			err := c.Write(buffer)
			if err != nil {
				t.Error("发送消息出错")
			}
		})

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Error(err)
	}
}

func TestHeart_HandlerDisconnect(t *testing.T) {
	testServer := NewTestServer(func(c *blue.Client, data *bytes.Buffer) {})
	server := httptest.NewServer(testServer)
	defer server.Close()

	fn := func(data string) bool {
		socket, err := NewDialer(server.URL)
		defer socket.Close()

		if err != nil {
			t.Error(err)
			return false
		}

		socket.WriteMessage(websocket.TextMessage, []byte(data))
		testServer.HandlerConnect(func(c *blue.Client) {
			c.Close()
		})
		testServer.HandlerDisconnect(func(c *blue.Client) {
			buffer := bytes.NewBufferString("AMD, yes!")
			err := c.Write(buffer)
			if err != nil {
				t.Error("发送消息出错")
			}
		})

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Error(err)
	}
}

func TestClientManage_Len(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	n := 100
	sockets := make([]*websocket.Conn, n)
	defer func() {
		for _, socket := range sockets {
			if socket != nil {
				socket.Close()
			}
		}
	}()

	testServer := NewTestServer(func(c *blue.Client, data *bytes.Buffer) {})
	server := httptest.NewServer(testServer)
	defer server.Close()

	disconnected := 0
	for i := 0; i < n; i++ {
		socket, err := NewDialer(server.URL)
		if err != nil {
			t.Error(err)
		}

		//一般概率产生失败连接，达成随机失败
		if (rand.Int() % 2) == 1 {
			sockets[i] = nil
			disconnected++
			socket.Close()
			continue
		}

		sockets[i] = socket
	}

	time.Sleep(time.Millisecond)

	connected := n - disconnected

	if testServer.Len() != connected {
		t.Errorf("框架自带的Len(): %d, 测试计算出的连接数: %d", testServer.Len(), connected)
	}
}

func TestHeart_Broadcast(t *testing.T) {
	testServer := NewTestServer(func(c *blue.Client, data *bytes.Buffer) {})
	testServer.HandlerMessage(func(c *blue.Client, data *bytes.Buffer) {
		testServer.Broadcast(data)
	})
	server := httptest.NewServer(testServer)
	defer server.Close()

	fn := func(data string) bool {
		conn, err := NewDialer(server.URL)
		defer conn.Close()

		if err != nil {
			t.Error(err)
			return false
		}

		n := 10
		sockets := make([]*websocket.Conn, n)
		for i := 0; i < n; i++ {
			socket, _ := NewDialer(server.URL)
			sockets[i] = socket
			defer sockets[i].Close()
		}

		conn.WriteMessage(websocket.TextMessage, []byte(data))

		for i := 0; i < n; i++ {
			_, message, err := sockets[i].ReadMessage()
			if err != nil {
				t.Error(err)
				return false
			}

			m := string(message)
			if data != m {
				t.Errorf("old: %s, new: %s", data, m)
				return false
			}
		}

		return true
	}

	//if !fn("AMD, yes!") {
	//    t.Error("error")
	//}
	if err := quick.Check(fn, nil); err != nil {
		t.Error(err)
	}
}
