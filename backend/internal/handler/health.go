package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 헬스체크 엔드포인트
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// 루트 엔드포인트
func Root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Gin basic API server is running",
	})
}
