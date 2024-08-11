package auth

import (
	"context"
	"time"

	"github.com/cristalhq/jwt/v5"
)

type JwtOpt struct {
	JWTPrivateKey string
	jwtSigner     *jwt.HSAlg
}

type Claims struct {
	jwt.RegisteredClaims
	Payload interface{} `json:"payload"`
}

func (opt *Options) CreateJWTToken(ctx context.Context, payload interface{}, expiredAfterInMinutes int64) (tokenStr string, err error) {

	currTime := time.Now()
	expireAt := time.Now().Add(time.Duration(expiredAfterInMinutes) * time.Minute)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    opt.Cfg.Server.Host,
			Audience:  jwt.Audience{opt.Cfg.Server.Host},
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(currTime),
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
