package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/docs"
)

// OpenAPIDoc returns the generated OpenAPI document.
func OpenAPIDoc(c *gin.Context) {
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
}
