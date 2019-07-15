package main

import (
	"bytes"
	"github.com/flxxyz/blue"
	"github.com/flxxyz/crontab"
	"log"
	"net/http"
	"time"
)

func init() {
	crontab.Init()
}

func main() {
	addr := ":30000"
	log.Println("监听的地址: ", addr)

	//gin.SetMode(gin.ReleaseMode)
	//r := gin.Default()
	heart := blue.NewHeart()

	//r.GET("/", func(c *gin.Context) {
	//    heart.ServeHTTP(c.Writer, c.Request)
	//})

	heart.HandlerConnect(func(c *blue.Client) {
		log.Printf("[connect   ] [%s] addr: %s", c.ID.String(), c.Request.RemoteAddr)
	})
	heart.HandlerMessage(func(c *blue.Client, data *bytes.Buffer) {
		heart.Broadcast(data)
		log.Printf("[message   ] [%s] body: %s", c.ID.String(), data.String())
	})
	heart.HandlerDisconnect(func(c *blue.Client) {
		c.ID.WriteString("退出房间")
		heart.BroadcastFilter(c.ID, func(client *blue.Client) bool {
			return c.ID.String() != client.ID.String()
		})
		log.Printf("[disconnect] [%s]", c.ID.String())
		log.Printf("实时客户端的数量: %d", heart.Len())
	})

	crontab.NewTicker(time.Second*60, func() {
		log.Printf("当前客户端的数量: %d", heart.Len())
	})

	//r.Run(addr)

	http.HandleFunc("/", heart.ServeHTTP)
	if err := http.ListenAndServe(addr, nil); err == nil {
		log.Fatal("GG")
	}
}
