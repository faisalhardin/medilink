package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
)

type JwtOpt struct {
	JWTPrivateKey string
	jwtSigner     *jwt.HSAlg
	jwtVerifier   *jwt.HSAlg
}

type Claims struct {
	jwt.RegisteredClaims
	Payload model.UserJWTPayload `json:"payload"`
}

func (claims Claims) Verify() (err error) {

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return commonerr.SetNewTokenExpiredError()
	}

	return nil
}

func (opt *Options) CreateJWTToken(ctx context.Context, payload model.UserJWTPayload, timeNow, timeExpired time.Time) (tokenStr string, err error) {

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    opt.Cfg.Server.Host,
			Audience:  jwt.Audience{opt.Cfg.Server.Host},
			ExpiresAt: jwt.NewNumericDate(timeExpired),
			IssuedAt:  jwt.NewNumericDate(timeNow),
		},
		Payload: payload,
	}

	return opt.generateToken(ctx, claims, timeNow, timeExpired)
}

func (opt *Options) generateToken(_ context.Context, claims any, _, _ time.Time) (tokenStr string, err error) {

	// Build and sign token
	builder := jwt.NewBuilder(opt.JwtOpt.jwtSigner)
	token, err := builder.Build(&claims)
	if err != nil {

		return
	}

	return token.String(), nil
}

func (opt *Options) CreateTokenPair(ctx context.Context, payload model.UserJWTPayload, deviceInfo model.DeviceInfo) (tokenPair model.TokenPair, err error) {
	now := time.Now()

	// Create access token (short-lived, 15-30 minutes)
	accessTokenExpiry := now.Add(time.Duration(opt.Cfg.JWTConfig.DurationInMinutes) * time.Minute)
	accessToken, err := opt.CreateJWTToken(ctx, payload, now, accessTokenExpiry)
	if err != nil {
		return tokenPair, err
	}

	// Create refresh token (long-lived, 7-30 days)
	refreshTokenExpiry := now.Add(time.Duration(opt.Cfg.JWTConfig.RefreshDurationInDays) * 24 * time.Hour)
	refreshToken, err := opt.generateRefreshToken()
	if err != nil {
		return tokenPair, err
	}

	// Store refresh token in database
	refreshTokenRecord := model.RefreshToken{
		Token:         refreshToken,
		UserID:        payload.UserID,
		InstitutionID: payload.InstitutionID,
		DeviceID:      deviceInfo.DeviceID,
		UserAgent:     deviceInfo.UserAgent,
		IPAddress:     deviceInfo.IPAddress,
		ExpiresAt:     refreshTokenExpiry,
	}

	err = opt.StoreRefreshToken(ctx, &refreshTokenRecord)
	if err != nil {
		return tokenPair, err
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(opt.Cfg.JWTConfig.DurationInMinutes * 60),
	}, nil
}

func (opt *Options) generateRefreshToken() (string, error) {
	// Generate a cryptographically secure random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}
