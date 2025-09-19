package auth

import (
	"context"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

func (opt *Options) StoreRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = opt.DB.MasterDB.Context(ctx)
	}

	_, err := session.Table("mdl_refresh_tokens").InsertOne(token)
	if err != nil {
		return errors.Wrap(err, "failed to store refresh token")
	}

	return nil
}

func (opt *Options) GetRefreshToken(ctx context.Context, token string) (*model.RefreshToken, error) {
	session := opt.DB.SlaveDB.Context(ctx)

	refreshToken := &model.RefreshToken{}
	found, err := session.Table("mdl_refresh_tokens").
		Where("token = ? AND is_revoked = false AND expires_at > ?", token, time.Now()).
		Get(refreshToken)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get refresh token")
	}

	if !found {
		return nil, errors.New("refresh token not found or expired")
	}

	return refreshToken, nil
}

func (opt *Options) RevokeRefreshToken(ctx context.Context, token string) error {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = opt.DB.MasterDB.Context(ctx)
	}

	now := time.Now()
	_, err := session.Table("mdl_refresh_tokens").
		Where("token = ?", token).
		Update(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if err != nil {
		return errors.Wrap(err, "failed to revoke refresh token")
	}

	return nil
}

func (opt *Options) RevokeAllUserRefreshTokens(ctx context.Context, userID int64) error {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = opt.DB.MasterDB.Context(ctx)
	}

	now := time.Now()
	_, err := session.Table("mdl_refresh_tokens").
		Where("user_id = ? AND is_revoked = false", userID).
		Update(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if err != nil {
		return errors.Wrap(err, "failed to revoke all user refresh tokens")
	}

	return nil
}

func (opt *Options) CleanupExpiredTokens(ctx context.Context) error {
	session := opt.DB.MasterDB.Context(ctx)

	_, err := session.Table("mdl_refresh_tokens").
		Where("expires_at < ? OR is_revoked = true", time.Now()).
		Delete(&model.RefreshToken{})

	if err != nil {
		return errors.Wrap(err, "failed to cleanup expired tokens")
	}

	return nil
}

func (opt *Options) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	session := opt.DB.SlaveDB.Context(ctx)

	count, err := session.Table("mdl_refresh_tokens").
		Where("token = ? AND is_revoked = true", token).
		Count(&model.RefreshToken{})

	if err != nil {
		return false, errors.Wrap(err, "failed to check token blacklist")
	}

	return count > 0, nil
}

func (opt *Options) GetActiveRefreshTokensByDevice(ctx context.Context, userID int64, deviceID string) ([]model.RefreshToken, error) {
	session := opt.DB.SlaveDB.Context(ctx)

	var tokens []model.RefreshToken
	err := session.Table("mdl_refresh_tokens").
		Where("user_id = ? AND device_id = ? AND is_revoked = false AND expires_at > ?", userID, deviceID, time.Now()).
		Find(&tokens)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get active refresh tokens by device")
	}

	return tokens, nil
}

func (opt *Options) RevokeRefreshTokensByDevice(ctx context.Context, userID int64, deviceID string) error {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = opt.DB.MasterDB.Context(ctx)
	}

	now := time.Now()
	_, err := session.Table("mdl_refresh_tokens").
		Where("user_id = ? AND device_id = ? AND is_revoked = false", userID, deviceID).
		Update(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if err != nil {
		return errors.Wrap(err, "failed to revoke refresh tokens by device")
	}

	return nil
}

// Login Tracking Methods

func (opt *Options) RecordLoginAttempt(ctx context.Context, loginReq model.LoginRequest) error {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = opt.DB.MasterDB.Context(ctx)
	}

	loginTracking := &model.MstLogin{
		UserID:        loginReq.UserID,
		InstitutionID: loginReq.InstitutionID,
		DeviceID:      loginReq.DeviceID,
		UserAgent:     loginReq.UserAgent,
		IPAddress:     loginReq.IPAddress,
		LoginType:     loginReq.LoginType,
		SessionID:     loginReq.SessionID,
		Status:        loginReq.Status,
		FailureReason: loginReq.FailureReason,
		LoginAt:       time.Now(),
		ExpiresAt:     &loginReq.ExpiresAt,
	}

	_, err := session.Table("mdl_mst_login").InsertOne(loginTracking)
	if err != nil {
		return errors.Wrap(err, "failed to record login attempt")
	}

	return nil
}

func (opt *Options) UpdateLoginLogout(ctx context.Context, sessionID string, logoutAt time.Time) error {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = opt.DB.MasterDB.Context(ctx)
	}

	_, err := session.Table("mdl_mst_login").
		Where("session_id = ? AND status = 'success' AND logout_at IS NULL", sessionID).
		Update(map[string]interface{}{
			"logout_at": logoutAt,
			"status":    "revoked",
		})

	if err != nil {
		return errors.Wrap(err, "failed to update login logout")
	}

	return nil
}

func (opt *Options) GetLoginHistory(ctx context.Context, params model.GetLoginHistoryParams) ([]model.LoginHistoryResponse, error) {
	session := opt.DB.SlaveDB.Context(ctx)

	var loginHistory []model.LoginHistoryResponse

	query := session.Table("mdl_mst_login").Alias("mlt").
		Join("LEFT", "mdl_mst_staff mms", "mms.id = mlt.user_id").
		Join("LEFT", "mdl_mst_institution mmi", "mmi.id = mlt.institution_id").
		Select("mlt.*, mms.name as user_name, mms.email as user_email, mmi.name as institution_name")

	// Apply filters
	if params.UserID > 0 {
		query = query.Where("mlt.user_id = ?", params.UserID)
	}
	if params.InstitutionID > 0 {
		query = query.Where("mlt.institution_id = ?", params.InstitutionID)
	}
	if params.DeviceID != "" {
		query = query.Where("mlt.device_id = ?", params.DeviceID)
	}
	if params.Status != "" {
		query = query.Where("mlt.status = ?", params.Status)
	}
	if params.LoginType != "" {
		query = query.Where("mlt.login_type = ?", params.LoginType)
	}
	if params.StartDate != "" {
		query = query.Where("mlt.login_at >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("mlt.login_at <= ?", params.EndDate)
	}

	// Apply pagination
	if params.Limit > 0 {
		query = query.Limit(params.Limit, params.Offset)
	}

	// Order by login time descending
	query = query.OrderBy("mlt.login_at DESC")

	err := query.Find(&loginHistory)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get login history")
	}

	return loginHistory, nil
}

func (opt *Options) GetActiveSessions(ctx context.Context, userID int64) ([]model.LoginHistoryResponse, error) {
	session := opt.DB.SlaveDB.Context(ctx)

	var activeSessions []model.LoginHistoryResponse

	err := session.Table("mdl_mst_login").Alias("mlt").
		Join("LEFT", "mdl_mst_staff mms", "mms.id = mlt.user_id").
		Join("LEFT", "mdl_mst_institution mmi", "mmi.id = mlt.institution_id").
		Select("mlt.*, mms.name as user_name, mms.email as user_email, mmi.name as institution_name").
		Where("mlt.user_id = ? AND mlt.status = 'success' AND mlt.logout_at IS NULL AND (mlt.expires_at IS NULL OR mlt.expires_at > ?)", userID, time.Now()).
		OrderBy("mlt.login_at DESC").
		Find(&activeSessions)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get active sessions")
	}

	return activeSessions, nil
}

func (opt *Options) GetDeviceLoginStats(ctx context.Context, userID int64) (map[string]int, error) {
	session := opt.DB.SlaveDB.Context(ctx)

	var results []struct {
		DeviceID string `xorm:"device_id"`
		Count    int    `xorm:"count"`
	}

	err := session.Table("mdl_mst_login").
		Select("device_id, COUNT(*) as count").
		Where("user_id = ? AND status = 'success'", userID).
		GroupBy("device_id").
		Find(&results)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get device login stats")
	}

	stats := make(map[string]int)
	for _, result := range results {
		stats[result.DeviceID] = result.Count
	}

	return stats, nil
}
