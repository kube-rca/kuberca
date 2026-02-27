package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kube-rca/backend/internal/service"
)

const (
	oidcStateCookie    = "kube_rca_oidc_state"
	oidcNonceCookie    = "kube_rca_oidc_nonce"
	oidcVerifierCookie = "kube_rca_oidc_verifier"
	oidcCookieMaxAge   = 600 // 10 minutes
)

type OIDCHandler struct {
	svc *service.OIDCService
}

func NewOIDCHandler(svc *service.OIDCService) *OIDCHandler {
	return &OIDCHandler{svc: svc}
}

// Login godoc
// @Summary Initiate OIDC login
// @Description Generates state, nonce, PKCE and redirects to the OIDC provider.
// @Tags auth
// @Success 302 "Redirect to OIDC provider"
// @Router /api/v1/auth/oidc/login [get]
func (h *OIDCHandler) Login(c *gin.Context) {
	state, err := randomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	nonce, err := randomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}

	codeVerifier, err := randomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate code verifier"})
		return
	}
	codeChallenge := computeCodeChallenge(codeVerifier)

	cookieCfg := h.svc.CookieConfig()
	c.SetSameSite(cookieCfg.SameSite)
	c.SetCookie(oidcStateCookie, state, oidcCookieMaxAge, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)
	c.SetCookie(oidcNonceCookie, nonce, oidcCookieMaxAge, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)
	c.SetCookie(oidcVerifierCookie, codeVerifier, oidcCookieMaxAge, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)

	authURL := h.svc.AuthURLWithPKCE(state, nonce, codeChallenge)
	c.Redirect(http.StatusFound, authURL)
}

// Callback godoc
// @Summary OIDC callback
// @Description Handles the OIDC provider callback, verifies state/nonce, exchanges code for tokens.
// @Tags auth
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 302 "Redirect to frontend"
// @Failure 302 "Redirect to frontend with error"
// @Router /api/v1/auth/oidc/callback [get]
func (h *OIDCHandler) Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		log.Printf("OIDC callback: missing code or state")
		c.Redirect(http.StatusFound, "/?error=oidc_invalid_request")
		return
	}

	savedState, err := c.Cookie(oidcStateCookie)
	if err != nil || savedState != state {
		log.Printf("OIDC callback: state mismatch")
		c.Redirect(http.StatusFound, "/?error=oidc_state_mismatch")
		return
	}

	savedNonce, err := c.Cookie(oidcNonceCookie)
	if err != nil {
		log.Printf("OIDC callback: missing nonce cookie")
		c.Redirect(http.StatusFound, "/?error=oidc_invalid_request")
		return
	}

	savedVerifier, _ := c.Cookie(oidcVerifierCookie)

	h.clearOIDCCookies(c)

	user, err := h.svc.HandleCallback(c.Request.Context(), code, savedVerifier, savedNonce)
	if err != nil {
		log.Printf("OIDC callback error: %v", err)
		if errors.Is(err, service.ErrOIDCNotAllowed) {
			c.Redirect(http.StatusFound, "/?error=oidc_not_allowed")
			return
		}
		c.Redirect(http.StatusFound, "/?error=oidc_failed")
		return
	}

	_, refreshToken, _, err := h.svc.IssueTokens(c.Request.Context(), user)
	if err != nil {
		log.Printf("OIDC callback: failed to issue tokens: %v", err)
		c.Redirect(http.StatusFound, "/?error=oidc_failed")
		return
	}

	cookieCfg := h.svc.CookieConfig()
	c.SetSameSite(cookieCfg.SameSite)
	c.SetCookie(cookieCfg.Name, refreshToken, cookieCfg.MaxAge, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)

	c.Redirect(http.StatusFound, "/")
}

func (h *OIDCHandler) clearOIDCCookies(c *gin.Context) {
	cookieCfg := h.svc.CookieConfig()
	c.SetSameSite(cookieCfg.SameSite)
	c.SetCookie(oidcStateCookie, "", -1, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)
	c.SetCookie(oidcNonceCookie, "", -1, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)
	c.SetCookie(oidcVerifierCookie, "", -1, cookieCfg.Path, cookieCfg.Domain, cookieCfg.Secure, true)
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func computeCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
