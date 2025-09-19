package model

import (
	"time"
)

type RefreshToken struct {
	ID            int64      `json:"id" xorm:"'id' pk autoincr"`
	Token         string     `json:"token" xorm:"'token' unique"`
	UserID        int64      `json:"user_id" xorm:"'user_id'"`
	InstitutionID int64      `json:"institution_id" xorm:"'institution_id'"`
	DeviceID      string     `json:"device_id" xorm:"'device_id'"`
	UserAgent     string     `json:"user_agent" xorm:"'user_agent'"`
	IPAddress     string     `json:"ip_address" xorm:"'ip_address'"`
	IsRevoked     bool       `json:"is_revoked" xorm:"'is_revoked' default false"`
	ExpiresAt     time.Time  `json:"expires_at" xorm:"'expires_at'"`
	CreatedAt     time.Time  `json:"created_at" xorm:"'created_at' created"`
	UpdatedAt     time.Time  `json:"updated_at" xorm:"'updated_at' updated"`
	RevokedAt     *time.Time `json:"revoked_at" xorm:"'revoked_at'"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validation:"required"`
	DeviceID     string `json:"device_id"`
	UserAgent    string `json:"user_agent"`
	IPAddress    string `json:"ip_address"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type DeviceInfo struct {
	DeviceID  string `json:"device_id"`
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}

type MstLogin struct {
	ID            int64      `json:"id" xorm:"'id' pk autoincr"`
	UserID        int64      `json:"user_id" xorm:"'user_id'"`
	InstitutionID int64      `json:"institution_id" xorm:"'institution_id'"`
	DeviceID      string     `json:"device_id" xorm:"'device_id'"`
	UserAgent     string     `json:"user_agent" xorm:"'user_agent'"`
	IPAddress     string     `json:"ip_address" xorm:"'ip_address'"`
	LoginType     string     `json:"login_type" xorm:"'login_type'"`
	SessionID     string     `json:"session_id" xorm:"'session_id'"`
	Status        string     `json:"status" xorm:"'status'"`
	FailureReason string     `json:"failure_reason" xorm:"'failure_reason'"`
	LoginAt       time.Time  `json:"login_at" xorm:"'login_at'"`
	LogoutAt      *time.Time `json:"logout_at" xorm:"'logout_at'"`
	ExpiresAt     *time.Time `json:"expires_at" xorm:"'expires_at'"`
	CreatedAt     time.Time  `json:"created_at" xorm:"'created_at' created"`
	UpdatedAt     time.Time  `json:"updated_at" xorm:"'updated_at' updated"`
}

type LoginRequest struct {
	UserID        int64     `json:"user_id"`
	InstitutionID int64     `json:"institution_id"`
	DeviceID      string    `json:"device_id"`
	UserAgent     string    `json:"user_agent"`
	IPAddress     string    `json:"ip_address"`
	LoginType     string    `json:"login_type"`
	SessionID     string    `json:"session_id"`
	Status        string    `json:"status"`
	FailureReason string    `json:"failure_reason"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type LoginHistoryResponse struct {
	MstLogin
	UserName        string `json:"user_name" xorm:"'user_name'"`
	UserEmail       string `json:"user_email" xorm:"'user_email'"`
	InstitutionName string `json:"institution_name" xorm:"'institution_name'"`
}

type GetLoginHistoryParams struct {
	UserID        int64  `json:"user_id" schema:"user_id"`
	InstitutionID int64  `json:"institution_id" schema:"institution_id"`
	DeviceID      string `json:"device_id" schema:"device_id"`
	Status        string `json:"status" schema:"status"`
	LoginType     string `json:"login_type" schema:"login_type"`
	StartDate     string `json:"start_date" schema:"start_date"`
	EndDate       string `json:"end_date" schema:"end_date"`
	CommonRequestPayload
}
