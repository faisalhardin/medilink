package model

import "time"

const (
	MstUserSession = "mdl_mst_user_sessions"
)

type UserSession struct {
	ID               int64      `xorm:"'id' pk autoincr" json:"-"`
	SessionKey       string     `xorm:"'session_key' unique" json:"session_key"`
	UserID           int64      `xorm:"'user_id'" json:"user_id"`
	AccessTokenHash  string     `xorm:"'access_token_hash'" json:"-"`
	RefreshTokenHash string     `xorm:"'refresh_token_hash'" json:"-"`
	Status           string     `xorm:"'status'" json:"status"`
	ExpiresAt        time.Time  `xorm:"'expires_at'" json:"expires_at"`
	RefreshExpiresAt time.Time  `xorm:"'refresh_expires_at'" json:"refresh_expires_at"`
	LastAccessedAt   time.Time  `xorm:"'last_accessed_at'" json:"last_accessed_at"`
	IPAddress        string     `xorm:"'ip_address'" json:"ip_address"`
	UserAgent        string     `xorm:"'user_agent'" json:"user_agent"`
	CreatedAt        time.Time  `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt        time.Time  `xorm:"'updated_at' updated" json:"updated_at"`
	DeleteTime       *time.Time `xorm:"'deleted_at' deleted" json:"-"`
}

type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusRevoked SessionStatus = "revoked"
	SessionStatusExpired SessionStatus = "expired"
)

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type LogoutRequest struct {
	SessionKey  string `json:"session_key,omitempty"`
	AllSessions bool   `json:"all_sessions,omitempty"`
}

type SessionInfo struct {
	SessionKey     string    `json:"session_key"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
	IPAddress      string    `json:"ip_address"`
	UserAgent      string    `json:"user_agent"`
	CreatedAt      time.Time `json:"created_at"`
}
