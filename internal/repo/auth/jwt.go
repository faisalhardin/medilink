package auth

import (
	"context"
	"time"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/config"
)

type JwtOpt struct {
	Cfg config.Config
}

type Claims struct {
	jwt.RegisteredClaims
	UserData interface{}
}

func (opt *Options) CreateJWTToken(ctx context.Context, userData interface{}) (err error) {

	expireAt := time.Now().Add(time.Duration(opt.Cfg.JWTConfig.DurationInHour) * time.Hour)

	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    opt.Cfg.Server.Host,
			Audience:  jwt.Audience{opt.Cfg.Server.Host},
			ExpiresAt: jwt.NewNumericDate(expireAt),
		},
		UserData: userData,
	}

	// Create signer
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(opt.Cfg.Vault.JWTCredential.Secret))
	if err != nil {

		return
	}

	// Build and sign token
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(&claims)
	if err != nil {

		return
	}
}
