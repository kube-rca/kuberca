package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kube-rca/backend/internal/config"
	"github.com/kube-rca/backend/internal/db"
	"github.com/kube-rca/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

const (
	refreshCookieName = "kube_rca_refresh"
	minLoginIDLength  = 3
	minPasswordLength = 8
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrMisconfigured = errors.New("auth config invalid")
)

type CookieConfig struct {
	Name     string
	Path     string
	Domain   string
	Secure   bool
	SameSite http.SameSite
	MaxAge   int
}

type AuthService struct {
	repo        *db.Postgres
	jwtSecret   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	allowSignup bool
	cookieCfg   CookieConfig
}

type authClaims struct {
	LoginID string `json:"loginId"`
	jwt.RegisteredClaims
}

func NewAuthService(repo *db.Postgres, cfg config.AuthConfig) (*AuthService, error) {
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("%w: JWT_SECRET is required", ErrMisconfigured)
	}

	accessTTL, err := time.ParseDuration(cfg.JWTAccessTTL)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid JWT_ACCESS_TTL", ErrMisconfigured)
	}

	refreshTTL, err := time.ParseDuration(cfg.JWTRefreshTTL)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid JWT_REFRESH_TTL", ErrMisconfigured)
	}

	allowSignup, err := parseBool(cfg.AllowSignup, false)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid ALLOW_SIGNUP", ErrMisconfigured)
	}

	cookieSecure, err := parseBool(cfg.CookieSecure, true)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid AUTH_COOKIE_SECURE", ErrMisconfigured)
	}

	cookieSameSite, err := parseSameSite(cfg.CookieSameSite)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid AUTH_COOKIE_SAMESITE", ErrMisconfigured)
	}

	if cookieSameSite == http.SameSiteNoneMode && !cookieSecure {
		return nil, fmt.Errorf("%w: SameSite=None requires Secure cookie", ErrMisconfigured)
	}

	cookiePath := cfg.CookiePath
	if strings.TrimSpace(cookiePath) == "" {
		cookiePath = "/"
	}

	return &AuthService{
		repo:        repo,
		jwtSecret:   []byte(cfg.JWTSecret),
		accessTTL:   accessTTL,
		refreshTTL:  refreshTTL,
		allowSignup: allowSignup,
		cookieCfg: CookieConfig{
			Name:     refreshCookieName,
			Path:     cookiePath,
			Domain:   cfg.CookieDomain,
			Secure:   cookieSecure,
			SameSite: cookieSameSite,
			MaxAge:   int(refreshTTL.Seconds()),
		},
	}, nil
}

func (s *AuthService) EnsureSchema(ctx context.Context) error {
	return s.repo.EnsureAuthSchema(ctx)
}

func (s *AuthService) EnsureAdmin(ctx context.Context, loginID, password string) error {
	if strings.TrimSpace(loginID) == "" || strings.TrimSpace(password) == "" {
		return fmt.Errorf("%w: ADMIN_USERNAME/ADMIN_PASSWORD are required", ErrMisconfigured)
	}

	_, err := s.repo.GetUserByLoginID(ctx, loginID)
	if err == nil {
		return nil
	}
	if !db.IsNoRows(err) {
		return err
	}

	if err := validateCredentials(loginID, password); err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.repo.CreateUser(ctx, loginID, string(hash))
	return err
}

func (s *AuthService) AllowSignup() bool {
	return s.allowSignup
}

func (s *AuthService) CookieConfig() CookieConfig {
	return s.cookieCfg
}

func (s *AuthService) Register(ctx context.Context, loginID, password string) (string, string, int64, error) {
	if !s.allowSignup {
		return "", "", 0, ErrForbidden
	}

	if err := validateCredentials(loginID, password); err != nil {
		return "", "", 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", 0, err
	}

	user, err := s.repo.CreateUser(ctx, loginID, string(hash))
	if err != nil {
		if isUniqueViolation(err) {
			return "", "", 0, ErrConflict
		}
		return "", "", 0, err
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, loginID, password string) (string, string, int64, error) {
	if err := validateCredentials(loginID, password); err != nil {
		return "", "", 0, err
	}

	user, err := s.repo.GetUserByLoginID(ctx, loginID)
	if err != nil {
		if db.IsNoRows(err) {
			return "", "", 0, ErrUnauthorized
		}
		return "", "", 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", 0, ErrUnauthorized
	}

	return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, int64, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", "", 0, ErrUnauthorized
	}

	hash := hashRefreshToken(refreshToken)
	record, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		if db.IsNoRows(err) {
			return "", "", 0, ErrUnauthorized
		}
		return "", "", 0, err
	}

	if record.RevokedAt != nil || time.Now().After(record.ExpiresAt) {
		return "", "", 0, ErrUnauthorized
	}

	user, err := s.repo.GetUserByID(ctx, record.UserID)
	if err != nil {
		return "", "", 0, err
	}

	newRefreshToken, newHash, err := newRefreshToken()
	if err != nil {
		return "", "", 0, err
	}

	if err := s.repo.RotateRefreshToken(ctx, record.ID, record.UserID, newHash, time.Now().Add(s.refreshTTL)); err != nil {
		return "", "", 0, err
	}

	accessToken, expiresIn, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, newRefreshToken, expiresIn, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return nil
	}

	hash := hashRefreshToken(refreshToken)
	return s.repo.RevokeRefreshTokenByHash(ctx, hash)
}

func (s *AuthService) ParseAccessToken(tokenStr string) (*model.AuthUser, error) {
	claims := &authClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return &model.AuthUser{
		ID:      userID,
		LoginID: claims.LoginID,
	}, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *model.User) (string, string, int64, error) {
	accessToken, expiresIn, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", 0, err
	}

	refreshToken, refreshHash, err := newRefreshToken()
	if err != nil {
		return "", "", 0, err
	}

	if err := s.repo.InsertRefreshToken(ctx, user.ID, refreshHash, time.Now().Add(s.refreshTTL)); err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, expiresIn, nil
}

func (s *AuthService) generateAccessToken(user *model.User) (string, int64, error) {
	now := time.Now()
	claims := authClaims{
		LoginID: user.LoginID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return signed, int64(s.accessTTL.Seconds()), nil
}

func validateCredentials(loginID, password string) error {
	loginID = strings.TrimSpace(loginID)
	password = strings.TrimSpace(password)

	if len(loginID) < minLoginIDLength || len(loginID) > 64 {
		return ErrInvalidInput
	}
	if len(password) < minPasswordLength || len(password) > 128 {
		return ErrInvalidInput
	}
	return nil
}

func parseBool(value string, fallback bool) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return parsed, nil
}

func parseSameSite(value string) (http.SameSite, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return http.SameSiteLaxMode, nil
	}
	switch value {
	case "lax":
		return http.SameSiteLaxMode, nil
	case "strict":
		return http.SameSiteStrictMode, nil
	case "none":
		return http.SameSiteNoneMode, nil
	default:
		return 0, ErrInvalidInput
	}
}

func newRefreshToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, hashRefreshToken(token), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
