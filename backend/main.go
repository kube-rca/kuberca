package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/handler"
)

func main() {
	// Gin의 기본 라우터 생성
	router := gin.Default()

	// Health Check & Root 엔드포인트
	router.GET("/ping", handler.Ping)
	router.GET("/", handler.Root)

	// Alert Manager 웹훅 엔드포인트
	router.POST("/webhook/alertmanager", handler.AlertmanagerWebhook)

	// 기본 포트 :8080 으로 서버 시작
	log.Println("Starting kube-rca-backend on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
