package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
	"golang.org/x/oauth2"
)

var ErrOIDCNotAllowed = fmt.Errorf("oidc: email not in allowlist")

type OIDCService struct {
	provider       *oidc.Provider
	oauth2Config   oauth2.Config
	verifier       *oidc.IDTokenVerifier
	authService    *AuthService
	repo           *db.Postgres
	allowedDomains []string
	allowedEmails  []string
	enabled        bool
}

func NewOIDCService(ctx context.Context, cfg config.OIDCConfig, authSvc *AuthService, repo *db.Postgres) (*OIDCService, error) {
	enabled := cfg.Enabled
	if !enabled {
		return &OIDCService{enabled: false}, nil
	}

	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURI == "" {
		return nil, fmt.Errorf("%w: OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_REDIRECT_URI are required when OIDC is enabled", ErrMisconfigured)
	}

	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to initialize provider (%s): %w", cfg.Issuer, err)
	}

	oauth2Cfg := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	allowedDomains := splitAndTrim(cfg.AllowedDomains)
	allowedEmails := splitAndTrim(cfg.AllowedEmails)

	if len(allowedDomains) == 0 && len(allowedEmails) == 0 {
		log.Printf("WARNING: OIDC enabled but no allowedDomains or allowedEmails configured - all OIDC logins will be denied")
	}
	log.Printf("OIDC enabled: issuer=%s, allowedDomains=%v, allowedEmails=%d entries", cfg.Issuer, allowedDomains, len(allowedEmails))

	return &OIDCService{
		provider:       provider,
		oauth2Config:   oauth2Cfg,
		verifier:       verifier,
		authService:    authSvc,
		repo:           repo,
		allowedDomains: allowedDomains,
		allowedEmails:  allowedEmails,
		enabled:        true,
	}, nil
}

func (s *OIDCService) Enabled() bool {
	return s.enabled
}

// ProviderName returns a short identifier for the OIDC provider based on the issuer URL.
func (s *OIDCService) ProviderName() string {
	if !s.enabled || s.provider == nil {
		return ""
	}
	issuer := s.provider.Endpoint().AuthURL
	switch {
	case strings.Contains(issuer, "accounts.google.com"):
		return "google"
	case strings.Contains(issuer, "/realms/"):
		return "keycloak"
	case strings.Contains(issuer, "okta.com"):
		return "okta"
	case strings.Contains(issuer, "login.microsoftonline.com"):
		return "azure"
	case strings.Contains(issuer, "gitlab"):
		return "gitlab"
	default:
		return "oidc"
	}
}

func (s *OIDCService) AuthURL(state, nonce string) string {
	return s.oauth2Config.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("access_type", "online"),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
}

func (s *OIDCService) AuthURLWithPKCE(state, nonce, codeChallenge string) string {
	return s.oauth2Config.AuthCodeURL(state,
		oidc.Nonce(nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("access_type", "online"),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
}

type oidcClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Sub           string `json:"sub"`
}

func (s *OIDCService) HandleCallback(ctx context.Context, code, codeVerifier, expectedNonce string) (*model.User, error) {
	opts := []oauth2.AuthCodeOption{}
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	oauth2Token, err := s.oauth2Config.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("oidc: code exchange failed: %w", err)
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("oidc: no id_token in response")
	}

	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("oidc: id_token verification failed: %w", err)
	}

	if idToken.Nonce != expectedNonce {
		return nil, fmt.Errorf("oidc: nonce mismatch")
	}

	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("oidc: failed to parse claims: %w", err)
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("oidc: email not verified")
	}

	if !s.checkAllowlist(claims.Email) {
		return nil, ErrOIDCNotAllowed
	}

	user, err := s.findOrCreateUser(ctx, claims)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *OIDCService) IssueTokens(ctx context.Context, user *model.User) (string, string, int64, error) {
	return s.authService.IssueTokens(ctx, user)
}

func (s *OIDCService) CookieConfig() CookieConfig {
	return s.authService.CookieConfig()
}

func (s *OIDCService) checkAllowlist(email string) bool {
	if len(s.allowedDomains) == 0 && len(s.allowedEmails) == 0 {
		return false
	}

	emailLower := strings.ToLower(email)

	parts := strings.SplitN(emailLower, "@", 2)
	if len(parts) == 2 {
		domain := parts[1]
		for _, d := range s.allowedDomains {
			if strings.EqualFold(d, domain) {
				return true
			}
		}
	}

	for _, e := range s.allowedEmails {
		if strings.EqualFold(e, emailLower) {
			return true
		}
	}

	return false
}

func (s *OIDCService) findOrCreateUser(ctx context.Context, claims oidcClaims) (*model.User, error) {
	user, err := s.repo.GetUserByOIDCSub(ctx, "oidc", claims.Sub)
	if err == nil {
		return user, nil
	}
	if !db.IsNoRows(err) {
		return nil, err
	}

	loginID := claims.Email
	if claims.Name != "" {
		loginID = fmt.Sprintf("%s (%s)", claims.Name, claims.Email)
	}
	pictureURL := claims.Picture
	if pictureURL != "" && !strings.HasPrefix(pictureURL, "https://") {
		pictureURL = ""
	}
	user, err = s.repo.CreateOIDCUser(ctx, loginID, "oidc", claims.Sub, claims.Email, claims.Name, pictureURL)
	if err != nil {
		if isUniqueViolation(err) {
			user, err = s.repo.GetUserByOIDCSub(ctx, "oidc", claims.Sub)
			if err != nil {
				return nil, err
			}
			return user, nil
		}
		return nil, err
	}

	return user, nil
}

func splitAndTrim(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
