package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/agent/internal/handler"
	"github.com/kube-rca/agent/internal/service"
)

func main() {
	analysisService := service.NewAnalysisService()
	analysisHandler := handler.NewAnalysisHandler(analysisService)

	router := gin.Default()

	router.GET("/ping", handler.Ping)
	router.GET("/healthz", handler.Healthz)
	router.GET("/", handler.Root)
	router.POST("/analyze/alertmanager", analysisHandler.AnalyzeAlertRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	addr := ":" + port

	log.Printf("Starting kube-rca-agent on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
