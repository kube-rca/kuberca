package model

import "time"

type AuthRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

type AuthConfigResponse struct {
	AllowSignup bool `json:"allowSignup"`
}

type AuthUser struct {
	ID      int64
	LoginID string
}

type User struct {
	ID           int64
	LoginID      string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
