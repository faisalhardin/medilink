package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/config"

	redisrepo "github.com/faisalhardin/medilink/internal/entity/repo/redis"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type Options struct {
	Cfg     *config.Config
	Storage redisrepo.Redis

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

func (opt *Options) StoreLoginInformation(ctx context.Context, key, staffDetail string, expiresIn time.Duration) (string, error) {

	if expiresIn.Hours() < 0 || expiresIn.Hours() > 8 {
		return "", errors.New("invalid token expiration data")
	}

	_, err := opt.Storage.SetWithExpire(key, staffDetail, int(expiresIn.Abs().Seconds()))
	if err != nil {
		return "", err
	}

	return key, nil
}

func (opt *Options) GetLoginInformation(ctx context.Context, key string) (string, error) {
	loginInformation, err := opt.Storage.Get(key)
	if err != nil {
		return "", err
	}

	return loginInformation, nil
}
