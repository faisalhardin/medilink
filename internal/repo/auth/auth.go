package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"

	redisrepo "github.com/faisalhardin/medilink/internal/entity/repo/redis"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Options struct {
	Cfg     *config.Config
	Storage redisrepo.Redis

	JwtOpt JwtOpt
}

func GoogleProvider(cfg *config.Config) goth.Provider {
	return google.New(cfg.Vault.GoogleAuth.ClientID, cfg.Vault.GoogleAuth.ClientSecret, "http://127.0.0.1:8080/v1/auth/google/callback")
}

func New(opt *Options, providers ...goth.Provider) (*Options, error) {

	opt.JwtOpt = JwtOpt{
		JWTPrivateKey: opt.Cfg.Vault.JWTCredential.Secret,
	}

	// Create signer
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(opt.Cfg.Vault.JWTCredential.Secret))
	if err != nil {
		return opt, errors.Wrap(err, "NewAuthOpt")
	}
	opt.JwtOpt.jwtSigner = signer

	verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(opt.Cfg.Vault.JWTCredential.Secret))
	if err != nil {
		return opt, errors.Wrap(err, "NewAuthOpt")
	}
	opt.JwtOpt.jwtVerifier = verifier

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

func (opt *Options) VerifyJWT(jwtToken string, claims any) (err error) {

	tokenParsed, err := jwt.Parse([]byte(jwtToken), opt.JwtOpt.jwtVerifier)
	if err != nil && errors.Is(err, jwt.ErrInvalidFormat) {
		return commonerr.SetNewBadRequest("authorization", err.Error())
	} else if err != nil {
		return err
	}

	marshalledToken, err := tokenParsed.Claims().MarshalJSON()
	if err != nil {
		return err
	}

	err = json.Unmarshal(marshalledToken, claims)
	if err != nil {
		return err
	}

	return nil
}
