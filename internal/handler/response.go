package handler

import "github.com/gin-gonic/gin"

func ok(c *gin.Context, data any) {
	c.JSON(200, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func fail(c *gin.Context, status int, code int, message string) {
	c.JSON(status, gin.H{
		"code":    code,
		"message": message,
		"data":    nil,
	})
}
