package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Gin의 기본 라우터 생성
	router := gin.Default()

	// 건강 체크 및 테스트용 기본 엔드포인트
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 루트 엔드포인트 예시
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Gin basic API server is running",
		})
	})

	// 기본 포트 :8080 으로 서버 시작
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
