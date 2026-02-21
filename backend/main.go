package main

import (
	"context"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/client"
	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/handler"
	"github.com/kube-rca/backend/internal/service"
)

func main() {
	// 의존성: handler → service → client
	// 초기화 순서: client → service → handler

	cfg := config.Load()
	log.Printf("check embedding model: %s", cfg.Embedding.Model)

	// 1. DB 연결 풀 초기화
	ctx := context.Background()
	dbPool, err := db.NewPostgresPool(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer dbPool.Close()

	// --- 추가된 부분: RCA 관련 초기화 ---
	// DB 구조체에 Pool 연결 (기존 db 패키지의 Postgres 구조체 사용)
	pgRepo := &db.Postgres{Pool: dbPool}

	// Incident 스키마 생성 (장애 단위)
	if err := pgRepo.EnsureIncidentSchema(); err != nil {
		log.Fatalf("Failed to ensure incident schema: %v", err)
	}

	// Alert 스키마 생성 (개별 알람 단위)
	if err := pgRepo.EnsureAlertSchema(); err != nil {
		log.Fatalf("Failed to ensure alert schema: %v", err)
	}

	// Alert 분석/근거 스키마 생성
	if err := pgRepo.EnsureAlertAnalysisSchema(); err != nil {
		log.Fatalf("Failed to ensure alert analysis schema: %v", err)
	}

	// 피드백(투표/코멘트) 스키마 생성
	if err := pgRepo.EnsureFeedbackSchema(); err != nil {
		log.Fatalf("Failed to ensure feedback schema: %v", err)
	}

	// Webhook 설정 스키마 생성
	if err := pgRepo.EnsureWebhookSchema(); err != nil {
		log.Fatalf("Failed to ensure webhook schema: %v", err)
	}

	// Embedding 스키마 생성 (pgvector 확장 및 embeddings 테이블)
	// todo: pgvector 확장 먼저 db에 설치해아함
	if err := pgRepo.EnsureEmbeddingSchema(ctx); err != nil {
		log.Fatalf("Failed to ensure embedding schema: %v", err)
	}

	authService, err := service.NewAuthService(pgRepo, cfg.Auth)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}
	if err := authService.EnsureSchema(ctx); err != nil {
		log.Fatalf("Failed to ensure auth schema: %v", err)
	}
	if err := authService.EnsureAdmin(ctx, cfg.Auth.AdminUsername, cfg.Auth.AdminPassword); err != nil {
		log.Fatalf("Failed to ensure admin user: %v", err)
	}
	authHandler := handler.NewAuthHandler(authService)

	embeddingClient, err := client.NewEmbeddingClient(cfg.Embedding)
	if err != nil {
		log.Fatalf("Failed to initialize embedding client: %v", err)
	}
	embeddingService := service.NewEmbeddingService(pgRepo, embeddingClient)
	embeddingHandler := handler.NewEmbeddingHandler(embeddingService)

	// 2. 외부 서비스 클라이언트 초기화
	slackClient := client.NewSlackClient(cfg.Slack)
	agentClient := client.NewAgentClient(cfg.Agent)

	// 3. 비즈니스 로직 서비스 초기화
	// AgentService: Agent 요청 및 Slack 쓰레드 응답 처리 + DB 저장
	agentService := service.NewAgentService(agentClient, slackClient, pgRepo)
	chatService := service.NewChatService(pgRepo, agentClient)
	// AlertService: 알림 필터링 및 Slack 전송 로직 담당 + DB 저장
	alertService := service.NewAlertService(slackClient, agentService, pgRepo)
	// RcaService: Incident/Alert 조회 및 종료 처리 + Agent 최종 분석 요청 + 임베딩 생성
	rcaSvc := service.NewRcaService(pgRepo, agentService, embeddingService)
	chatHandler := handler.NewChatHandler(chatService)
	webhookSvc := service.NewWebhookService(pgRepo)

	// 4. HTTP 핸들러 초기화
	// Alertmanager 웹훅 요청 수신 및 응답 처리
	alertHandler := handler.NewAlertHandler(alertService)
	rcaHndlr := handler.NewRcaHandler(rcaSvc)
	webhookHndlr := handler.NewWebhookSettingsHandler(webhookSvc)

	// HTTP 라우터 설정
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping", "/", "/openapi.json"},
	}))

	corsOrigins := splitOrigins(cfg.Auth.CorsAllowedOrigins)
	if len(corsOrigins) > 0 {
		router.Use(handler.CORSMiddleware(corsOrigins, true))
	}

	// Health Check 엔드포인트
	// - GET /ping: 서버 상태 확인용
	// - GET /: 루트 경로
	router.GET("/ping", handler.Ping)
	router.GET("/", handler.Root)
	router.GET("/openapi.json", handler.OpenAPIDoc)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", authHandler.Logout)
		auth.GET("/config", authHandler.Config)
		auth.GET("/me", handler.AuthMiddleware(authService), authHandler.Me)

		protected := v1.Group("")
		protected.Use(handler.AuthMiddleware(authService))
		protected.GET("/incidents", rcaHndlr.GetIncidents)
		protected.GET("/incidents/:id", rcaHndlr.GetIncidentDetail)
		protected.PUT("/incidents/:id", rcaHndlr.UpdateIncident)
		protected.PATCH("/incidents/:id", rcaHndlr.HideIncident)
		protected.GET("/incidents/hidden", rcaHndlr.GetHiddenIncidents)
		protected.PATCH("/incidents/:id/unhide", rcaHndlr.UnhideIncident)
		protected.POST("/incidents/:id/resolve", rcaHndlr.ResolveIncident)
		protected.GET("/incidents/:id/alerts", rcaHndlr.GetIncidentAlerts)
		protected.POST("/incidents/mock", rcaHndlr.CreateMockIncident)
		protected.GET("/incidents/:id/feedback", rcaHndlr.GetIncidentFeedback)
		protected.POST("/incidents/:id/comments", rcaHndlr.CreateIncidentComment)
		protected.PUT("/incidents/:id/comments/:commentId", rcaHndlr.UpdateIncidentComment)
		protected.DELETE("/incidents/:id/comments/:commentId", rcaHndlr.DeleteIncidentComment)
		protected.POST("/incidents/:id/vote", rcaHndlr.VoteIncidentFeedback)

		// Alert 엔드포인트
		protected.GET("/alerts", rcaHndlr.GetAlerts)
		protected.GET("/alerts/:id", rcaHndlr.GetAlertDetail)
		protected.PUT("/alerts/:id/incident", rcaHndlr.UpdateAlertIncident)
		protected.GET("/alerts/:id/feedback", rcaHndlr.GetAlertFeedback)
		protected.POST("/alerts/:id/comments", rcaHndlr.CreateAlertComment)
		protected.PUT("/alerts/:id/comments/:commentId", rcaHndlr.UpdateAlertComment)
		protected.DELETE("/alerts/:id/comments/:commentId", rcaHndlr.DeleteAlertComment)
		protected.POST("/alerts/:id/vote", rcaHndlr.VoteAlertFeedback)

		protected.POST("/embeddings", embeddingHandler.CreateEmbedding)
		protected.POST("/embeddings/search", embeddingHandler.SearchEmbeddings)
		protected.POST("/chat", chatHandler.Chat)

		// Settings 엔드포인트 (Webhook 설정 CRUD)
		protected.GET("/settings/webhooks", webhookHndlr.ListWebhookConfigs)
		protected.POST("/settings/webhooks", webhookHndlr.CreateWebhookConfig)
		protected.GET("/settings/webhooks/:id", webhookHndlr.GetWebhookConfig)
		protected.PUT("/settings/webhooks/:id", webhookHndlr.UpdateWebhookConfig)
		protected.DELETE("/settings/webhooks/:id", webhookHndlr.DeleteWebhookConfig)
	}

	// Alertmanager 웹훅 엔드포인트
	// - POST /webhook/alertmanager: Alertmanager에서 알림 수신
	router.POST("/webhook/alertmanager", alertHandler.Webhook)

	// 8080 서버 실행
	log.Println("Starting kube-rca-backend on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func splitOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		origins = append(origins, trimmed)
	}
	return origins
}
