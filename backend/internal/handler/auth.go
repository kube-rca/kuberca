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

// Register godoc
// @Summary Register a new user
// @Description Sign up when ALLOW_SIGNUP is true.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.AuthRequest true "Login ID and password"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 403 {object} model.ErrorResponse
// @Failure 409 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/register [post]
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

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body model.AuthRequest true "Login ID and password"
// @Success 200 {object} model.AuthResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/login [post]
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

// Refresh godoc
// @Summary Refresh access token
// @Description Uses refresh token cookie (kube_rca_refresh).
// @Tags auth
// @Produce json
// @Success 200 {object} model.AuthResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/auth/refresh [post]
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

// Logout godoc
// @Summary Logout
// @Description Revokes refresh token (if present) and clears cookie.
// @Tags auth
// @Produce json
// @Success 200 {object} model.AuthLogoutResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(h.svc.CookieConfig().Name)
	_ = h.svc.Logout(c.Request.Context(), refreshToken)
	h.clearRefreshCookie(c)
	c.JSON(http.StatusOK, model.AuthLogoutResponse{Status: "logged_out"})
}

// Config godoc
// @Summary Get auth config
// @Tags auth
// @Produce json
// @Success 200 {object} model.AuthConfigResponse
// @Router /api/v1/auth/config [get]
func (h *AuthHandler) Config(c *gin.Context) {
	c.JSON(http.StatusOK, model.AuthConfigResponse{
		AllowSignup: h.svc.AllowSignup(),
	})
}

// Me godoc
// @Summary Get current user
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} model.AuthMeResponse
// @Failure 401 {object} model.ErrorResponse
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	user := GetAuthUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, model.AuthMeResponse{
		UserID:  user.ID,
		LoginID: user.LoginID,
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
