package main

import (
	"bytes"
	"github.com/flxxyz/blue"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

var (
	addr  = ":7777"
	heart = blue.NewHeart()
)

func init() {
	log.Println("监听的地址: ", addr)

	ticker := time.NewTicker(time.Second * 10)

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("当前客户端的数量: %d", heart.Len())
			}
		}
	}()
}

func main() {
	heart.HandlerConnect(func(c *blue.Client) {
		log.Printf("[connect   ] [%s] addr: %s", c.ID.String(), c.Request.RemoteAddr)
	})

	heart.HandlerMessage(func(c *blue.Client, data *bytes.Buffer) {
		heart.Broadcast(data)
		log.Printf("[message   ] [%s] body: %s", c.ID.String(), data.String())
	})

	heart.HandlerDisconnect(func(c *blue.Client) {
		heart.BroadcastFilter(c.ID, func(client *blue.Client) bool {
			return c.ID.String() != client.ID.String()
		})
		log.Printf("[disconnect] [%s] 退出房间", c.ID.String())
		log.Printf("实时客户端的数量: %d", heart.Len())
	})

	//原生类库支持
	//httpserver()

	//gin支持
	ginserver()
}

func httpserver() {
	http.HandleFunc("/", heart.ServeHTTP)

	if err := http.ListenAndServe(addr, nil); err == nil {
		log.Fatal("GG")
	}
}

func ginserver() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		heart.ServeHTTP(c.Writer, c.Request)
	})

	r.Run(addr)
}
