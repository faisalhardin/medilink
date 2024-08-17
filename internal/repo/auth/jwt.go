package auth

import (
	"context"
	"time"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/entity/model"
)

type JwtOpt struct {
	JWTPrivateKey string
	jwtSigner     *jwt.HSAlg
}

type Claims struct {
	jwt.RegisteredClaims
	Payload interface{} `json:"payload"`
}

func (opt *Options) CreateJWTToken(ctx context.Context, payload model.UserJWTPayload, timeNow, timeExpired time.Time) (tokenStr string, err error) {
	return opt.generateToken(ctx, payload, timeNow, timeExpired)
}

func (opt *Options) generateToken(ctx context.Context, payload interface{}, timeNow, timeExpired time.Time) (tokenStr string, err error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    opt.Cfg.Server.Host,
			Audience:  jwt.Audience{opt.Cfg.Server.Host},
			ExpiresAt: jwt.NewNumericDate(timeExpired),
			IssuedAt:  jwt.NewNumericDate(timeNow),
		},
		Payload: payload,
	}

	// Build and sign token
	builder := jwt.NewBuilder(opt.JwtOpt.jwtSigner)
	token, err := builder.Build(&claims)
	if err != nil {

		return
	}

	return token.String(), nil
}
