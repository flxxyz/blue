package blue

import "time"

type Conf struct {
	ReadDeadline  time.Duration
	WriteDeadline time.Duration
	Message       message
}

type message struct {
	PacketBufferSize int
	MaxBufferSize    int64
	WriteBufferSize  int
	ReadBufferSize   int
}

func NewConf() *Conf {
	return &Conf{
		ReadDeadline:  time.Second * 60,
		WriteDeadline: time.Second * 10,
		Message: message{
			PacketBufferSize: 256,
			MaxBufferSize:    512,
			WriteBufferSize:  1024,
			ReadBufferSize:   1024,
		},
	}
}
