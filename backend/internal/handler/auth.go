package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/model"
	"github.com/kube-rca/backend/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req model.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	accessToken, refreshToken, expiresIn, err := h.svc.Register(c.Request.Context(), req.ID, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, model.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	accessToken, refreshToken, expiresIn, err := h.svc.Login(c.Request.Context(), req.ID, req.Password)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, refreshToken)
	c.JSON(http.StatusOK, model.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, _ := c.Cookie(h.svc.CookieConfig().Name)
	accessToken, newRefreshToken, expiresIn, err := h.svc.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		writeAuthError(c, err)
		return
	}

	h.setRefreshCookie(c, newRefreshToken)
	c.JSON(http.StatusOK, model.AuthResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(h.svc.CookieConfig().Name)
	_ = h.svc.Logout(c.Request.Context(), refreshToken)
	h.clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"status": "logged_out"})
}

func (h *AuthHandler) Config(c *gin.Context) {
	c.JSON(http.StatusOK, model.AuthConfigResponse{
		AllowSignup: h.svc.AllowSignup(),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user := GetAuthUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"userId":  user.ID,
		"loginId": user.LoginID,
	})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string) {
	cfg := h.svc.CookieConfig()
	c.SetSameSite(cfg.SameSite)
	c.SetCookie(cfg.Name, token, cfg.MaxAge, cfg.Path, cfg.Domain, cfg.Secure, true)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	cfg := h.svc.CookieConfig()
	c.SetSameSite(cfg.SameSite)
	c.SetCookie(cfg.Name, "", -1, cfg.Path, cfg.Domain, cfg.Secure, true)
}

func writeAuthError(c *gin.Context, err error) {
	switch err {
	case service.ErrInvalidInput:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case service.ErrUnauthorized:
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case service.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "signup disabled"})
	case service.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
	}
}
