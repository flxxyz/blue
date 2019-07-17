package blue

import (
	"sync"
)

type ClientManage struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Msg
	locker     *sync.RWMutex
	state      bool
}

//起一个goroutine读取新增，退出，广播的channel
func (cm *ClientManage) run() {
	for {
		select {
		case c := <-cm.register:
			cm.locker.Lock()
			cm.clients[c] = true
			cm.locker.Unlock()
		case c := <-cm.unregister:
			if _, ok := cm.clients[c]; ok {
				cm.locker.Lock()
				delete(cm.clients, c)
				//close(c.packet)
				cm.locker.Unlock()
			}
		case m := <-cm.broadcast:
			cm.locker.RLock()
			for c := range cm.clients {
				if m.Filter != nil {
					if m.Filter(c) {
						c.writeMessage(m)
					}
				} else {
					c.writeMessage(m)
				}
			}
			cm.locker.RUnlock()
		}
	}
}

func (cm *ClientManage) pushRegister(c *Client) {
	cm.register <- c
}

func (cm *ClientManage) pushUnregister(c *Client) {
	cm.unregister <- c
}

//读取当前人员数(读锁保障)
func (cm *ClientManage) Len() int {
	cm.locker.RLock()
	defer cm.locker.RUnlock()

	return len(cm.clients)
}

func (cm *ClientManage) closed() bool {
	cm.locker.RLock()
	defer cm.locker.RUnlock()

	return !cm.state
}

//实例客户端管理器
func NewClientManage() *ClientManage {
	return &ClientManage{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Msg),
		locker:     &sync.RWMutex{},
		state:      true,
	}
}
