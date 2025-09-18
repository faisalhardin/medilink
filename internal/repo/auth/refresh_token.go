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
