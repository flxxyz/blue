# blue
基于[gorilla/websocket](https://github.com/gorilla/websocket)构建的微型websocket框架

ps: 为什么叫这个名字呢，因为最近在听[the blue hearts](https://zh.wikipedia.org/zh/THE_BLUE_HEARTS)这个乐队

## 依赖要求
没有

## 安装
使用`go`命令获取类库

```bash
go get github.com/flxxyz/blue
```

## 例子

### net/http
```go
package main

import (
    "bytes"
    "github.com/flxxyz/blue"
    "log"
    "net/http"
)

func main() {
    heart := blue.NewHeart()
    heart.HandlerMessage(func(c *blue.Client, data *bytes.Buffer) {
        heart.Broadcast(data)
        log.Printf("[%s] body: %s", c.ID.String(), data.String())
    })
    
    http.HandleFunc("/", heart.ServeHTTP)
    if err := http.ListenAndServe(":7777", nil); err == nil {
        log.Fatal("GG")
    }
}
```

### gin
```go
package main

import (
    "bytes"
    "github.com/flxxyz/blue"
    "github.com/gin-gonic/gin"
    "log"
)

func main() {
    heart := blue.NewHeart()
    heart.HandlerMessage(func(c *blue.Client, data *bytes.Buffer) {
        heart.Broadcast(data)
        log.Printf("[%s] body: %s", c.ID.String(), data.String())
    })
    
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()
    r.GET("/", func(c *gin.Context) {
        heart.ServeHTTP(c.Writer, c.Request)
    })
    r.Run(":7777")
}
```

## 文档
[文档点这里](http://godoc.org/github.com/flxxyz/blue)

## 版权
blue包在MIT License下发布。有关详细信息，请参阅LICENSE。
