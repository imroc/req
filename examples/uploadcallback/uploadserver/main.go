package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func main() {
	router := gin.Default()
	router.POST("/upload", func(c *gin.Context) {
		body := c.Request.Body
		io.Copy(io.Discard, body)
		c.String(http.StatusOK, "ok")
	})
	router.Run(":8888")
}
