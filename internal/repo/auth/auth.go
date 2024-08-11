package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/go-redis/redis/v8"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type Options struct {
	Cfg *config.Config
	Str TokenStorage

	JwtOpt JwtOpt
}

func New(opt *Options, providers ...goth.Provider) (*Options, error) {

	store := sessions.NewCookieStore([]byte(opt.Cfg.Vault.GoogleAuth.Key))
	store.MaxAge(opt.Cfg.GoogleAuthConfig.MaxAge)

	store.Options.Path = opt.Cfg.GoogleAuthConfig.CookiePath
	store.Options.HttpOnly = opt.Cfg.GoogleAuthConfig.HttpOnly
	store.Options.Secure = opt.Cfg.GoogleAuthConfig.IsProd

	gothic.Store = store

	goth.UseProviders(
		providers...,
	)

	opt.JwtOpt = JwtOpt{
		JWTPrivateKey: opt.Cfg.Vault.JWTCredential.Secret,
	}

	// Create signer
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(opt.Cfg.Vault.JWTCredential.Secret))
	if err != nil {
		return opt, errors.Wrap(err, "NewAuthOpt")
	}
	opt.JwtOpt.jwtSigner = signer

	return opt, nil
}

func (opt *Options) GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) (goth.User, error) {
	return gothic.CompleteUserAuth(w, r)
}

func (opt *Options) BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (opt *Options) Logout(w http.ResponseWriter, r *http.Request) {
	gothic.Logout(w, r)
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (opt *Options) CreateToken(ctx context.Context, staffDetail string, expiresIn time.Duration) (string, error) {
	token, err := GenerateOpaqueToken()
	if err != nil {
		return "", err
	}

	err = opt.Str.Set(ctx, token, staffDetail, expiresIn)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (opt *Options) ValidateToken(ctx context.Context, token string) (string, error) {
	userID, err := opt.Str.Get(ctx, token)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", errors.New("invalid token")
		}
		return "", err
	}
	return userID, nil
}
