package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	//启动gin服务
	r := gin.Default()

	// 静态文件服务
	fs := http.FileServer(http.Dir("./dist"))
	// 中间件处理路由
	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next() // 继续处理API请求
			return
		}

		// 处理静态文件
		fs.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	})

	r.Run()
}
