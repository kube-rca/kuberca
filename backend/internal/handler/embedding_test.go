package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/service"
)

func TestEmbeddingsHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	var svc *service.EmbeddingService
	r.POST("/api/v1/embeddings", NewEmbeddingHandler(svc).CreateEmbedding)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/embeddings", bytes.NewBufferString(`{"incident_id":"","incident_summary":""}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
