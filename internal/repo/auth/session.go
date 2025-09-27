package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

type SessionRepository struct {
	db *xorm.DBConnect
}

func NewSessionRepository(db *xorm.DBConnect) *SessionRepository {
	return &SessionRepository{db: db}
}

// CreateSession creates a new user session in the database
func (r *SessionRepository) CreateSession(ctx context.Context, session *model.UserSession) error {
	_, err := r.db.MasterDB.Context(ctx).Table(model.MstUserSession).Insert(session)
	if err != nil {
		return errors.Wrap(err, "CreateSession")
	}
	return nil
}

// GetSessionByKey retrieves a session by its session key
func (r *SessionRepository) GetSessionByKey(ctx context.Context, sessionKey string) (*model.UserSession, error) {
	var session model.UserSession
	has, err := r.db.SlaveDB.Context(ctx).Table(model.MstUserSession).
		Where("session_key = ? AND deleted_at IS NULL", sessionKey).
		Get(&session)
	if err != nil {
		return nil, errors.Wrap(err, "GetSessionByKey")
	}
	if !has {
		return nil, errors.New("session not found")
	}
	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by its refresh token hash
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (*model.UserSession, error) {
	var session model.UserSession
	has, err := r.db.SlaveDB.Context(ctx).Table(model.MstUserSession).
		Where("refresh_token_hash = ? AND deleted_at IS NULL", refreshTokenHash).
		Get(&session)
	if err != nil {
		return nil, errors.Wrap(err, "GetSessionByRefreshToken")
	}
	if !has {
		return nil, commonerr.SetNewRevokedSessionError()
	}
	return &session, nil
}

// UpdateSessionStatus updates the status of a session
func (r *SessionRepository) UpdateSessionStatus(ctx context.Context, sessionKey, status string) error {
	sess := r.db.MasterDB.Context(ctx).Table(model.MstUserSession)
	_, err := sess.
		Where("session_key = ? AND deleted_at IS NULL", sessionKey).
		Update(&model.UserSession{Status: status})
	if err != nil {
		return errors.Wrap(err, "UpdateSessionStatus")
	}
	return nil
}

// UpdateSessionAccessToken updates the access token hash and expiration for a session
func (r *SessionRepository) UpdateSessionAccessToken(ctx context.Context, session *model.UserSession) error {
	sess := r.db.MasterDB.Context(ctx).Table(model.MstUserSession)

	_, err := sess.Where("session_key = ?", session.SessionKey).
		Update(session)
	if err != nil {
		return errors.Wrap(err, "UpdateSessionAccessToken")
	}
	return nil
}

// UpdateLastAccessed updates the last accessed time for a session
func (r *SessionRepository) UpdateLastAccessed(ctx context.Context, sessionKey string) error {
	_, err := r.db.MasterDB.Context(ctx).Table(model.MstUserSession).
		Where("session_key = ? AND deleted_at IS NULL", sessionKey).
		Update(map[string]interface{}{
			"last_accessed_at": time.Now(),
			"updated_at":       time.Now(),
		})
	if err != nil {
		return errors.Wrap(err, "UpdateLastAccessed")
	}
	return nil
}

// RevokeAllUserSessions revokes all active sessions for a user
func (r *SessionRepository) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	_, err := r.db.MasterDB.Context(ctx).Table(model.MstUserSession).
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, string(model.SessionStatusActive)).
		Update(map[string]interface{}{
			"status":     string(model.SessionStatusRevoked),
			"updated_at": time.Now(),
		})
	if err != nil {
		return errors.Wrap(err, "RevokeAllUserSessions")
	}
	return nil
}

// GetUserSessions retrieves all sessions for a user
func (r *SessionRepository) GetUserSessions(ctx context.Context, userID int64) ([]model.SessionInfo, error) {
	var sessions []model.SessionInfo
	err := r.db.SlaveDB.Context(ctx).Table(model.MstUserSession).
		Select("session_key, status, expires_at, last_accessed_at, ip_address, user_agent, created_at").
		Where("user_id = ? AND deleted_at IS NULL", userID).
		OrderBy("last_accessed_at DESC").
		Find(&sessions)
	if err != nil {
		return nil, errors.Wrap(err, "GetUserSessions")
	}
	return sessions, nil
}

// CleanupExpiredSessions marks expired sessions as expired
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	now := time.Now()
	nowStr := now.Format(time.RFC3339)
	_, err := r.db.MasterDB.Context(ctx).Table(model.MstUserSession).
		Where("status = ? AND (expires_at < ? OR refresh_expires_at < ?) AND deleted_at IS NULL",
			string(model.SessionStatusActive), nowStr, nowStr).
		Update(&model.UserSession{Status: string(model.SessionStatusExpired)})
	if err != nil {
		return errors.Wrap(err, "CleanupExpiredSessions")
	}
	return nil
}

// DeleteExpiredSessions permanently deletes old expired sessions (older than 30 days)
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -30) // 30 days ago
	_, err := r.db.MasterDB.Context(ctx).Table(model.MstUserSession).
		Where("(status = ? OR status = ?) AND updated_at < ?", string(model.SessionStatusExpired), string(model.SessionStatusRevoked), cutoffDate.Format(time.RFC3339)).
		Unscoped().
		Delete(&model.UserSession{})
	if err != nil {
		return errors.Wrap(err, "DeleteExpiredSessions")
	}
	return nil
}

// HashToken creates a SHA256 hash of a token for secure storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateRefreshToken generates a secure random refresh token
func GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", errors.Wrap(err, "GenerateRefreshToken")
	}
	return hex.EncodeToString(bytes), nil
}
