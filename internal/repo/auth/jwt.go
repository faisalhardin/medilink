package auth

import (
	"context"
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

func (opt *Options) generateToken(ctx context.Context, claims any, timeNow, timeExpired time.Time) (tokenStr string, err error) {

	// Build and sign token
	builder := jwt.NewBuilder(opt.JwtOpt.jwtSigner)
	token, err := builder.Build(&claims)
	if err != nil {

		return
	}

	return token.String(), nil
}
