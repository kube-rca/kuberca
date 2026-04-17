package model

import "time"

type AuthRequest struct {
	ID                string `json:"id"`
	Password          string `json:"password"`
	PreferredLanguage string `json:"preferredLanguage,omitempty"`
}

type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

type AuthConfigResponse struct {
	AllowSignup  bool   `json:"allowSignup"`
	OIDCEnabled  bool   `json:"oidcEnabled"`
	OIDCLoginURL string `json:"oidcLoginUrl,omitempty"`
	OIDCProvider string `json:"oidcProvider,omitempty"`
}

type AuthUser struct {
	ID                int64
	LoginID           string
	PreferredLanguage string
}

type User struct {
	ID                int64
	LoginID           string
	PasswordHash      string
	AuthProvider      string
	OIDCSub           *string
	Email             *string
	DisplayName       *string
	PictureURL        *string
	PreferredLanguage string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type UpdatePreferredLanguageRequest struct {
	PreferredLanguage string `json:"preferredLanguage"`
}

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}
