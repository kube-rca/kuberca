package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/handler"
	"github.com/kube-rca/backend/internal/service"
)

func main() {
	// 의존성: handler → service → client
	// 초기화 순서: client → service → handler

	// 1. DB 연결 풀 초기화
	ctx := context.Background()
	dbPool, err := db.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer dbPool.Close()

	// 2. 외부 서비스 클라이언트 초기화
	// Slack Bot Token과 Channel ID를 환경변수에서 읽어옴
	slackClient := client.NewSlackClient()
	// Agent URL을 환경변수에서 읽어옴
	agentClient := client.NewAgentClient()

	// 3. 비즈니스 로직 서비스 초기화
	// AgentService: Agent 요청 및 Slack 쓰레드 응답 처리
	agentService := service.NewAgentService(agentClient, slackClient)
	// AlertService: 알림 필터링 및 Slack 전송 로직 담당
	alertService := service.NewAlertService(slackClient, agentService)

	// 4. HTTP 핸들러 초기화
	// Alertmanager 웹훅 요청 수신 및 응답 처리
	alertHandler := handler.NewAlertHandler(alertService)

	// HTTP 라우터 설정
	router := gin.Default()

	// Health Check 엔드포인트
	// - GET /ping: 서버 상태 확인용
	// - GET /: 루트 경로
	router.GET("/ping", handler.Ping)
	router.GET("/", handler.Root)

	// Alertmanager 웹훅 엔드포인트
	// - POST /webhook/alertmanager: Alertmanager에서 알림 수신
	router.POST("/webhook/alertmanager", alertHandler.Webhook)

	// 8080 서버 실행
	log.Println("Starting kube-rca-backend on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
