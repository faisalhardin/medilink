package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/cristalhq/jwt/v5"
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/repo/cache"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"golang.org/x/oauth2"
	oauthgoogle "golang.org/x/oauth2/google"
)

type Options struct {
	Cfg        *config.Config
	Storage    cache.Caching
	GoogleAuth *oauth2.Config

	JwtOpt JwtOpt
}

func GoogleProvider(cfg *config.Config) goth.Provider {
	return google.New(cfg.Vault.GoogleAuth.ClientID, cfg.Vault.GoogleAuth.ClientSecret, fmt.Sprintf("http://%s/v1/auth/google/callback", cfg.Server.BaseURL))
}

func New(opt *Options, providers ...goth.Provider) (*Options, error) {

	opt.JwtOpt = JwtOpt{
		JWTPrivateKey: opt.Cfg.Vault.JWTCredential.Secret,
	}
	opt.GoogleAuth = &oauth2.Config{
		ClientID:     opt.Cfg.Vault.GoogleAuth.ClientID,
		ClientSecret: opt.Cfg.Vault.GoogleAuth.ClientSecret,
		RedirectURL:  opt.Cfg.WebConfig.Host,
		Scopes:       []string{"profile", "email"},
		Endpoint:     oauthgoogle.Endpoint,
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

func (opt *Options) GetGoogleAuthCallback(ctx context.Context, code string) (tokens *oauth2.Token, err error) {
	tokens, err = opt.GoogleAuth.Exchange(ctx, code)
	if err != nil {
		err = errors.Wrap(err, "GetGoogleAuthCallback")
		return
	}

	return
}

func (opt *Options) GetUserInfo(ctx context.Context, accessToken string) (user model.GoogleUser, err error) {
	userInfoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s?access_token=%s", userInfoEndpoint, accessToken), nil)
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	var userInfo model.GoogleUser
	if err = json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return
	}

	return userInfo, nil
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

func (opt *Options) StoreKeySession(ctx context.Context, tokenKey, token string, expiresIn time.Duration) (err error) {
	if expiresIn.Hours() < 0 || expiresIn.Hours() > 8 {
		return errors.New("invalid token expiration data")
	}

	_, err = opt.Storage.SetWithExpire(tokenKey, token, int(expiresIn.Abs().Seconds()))
	if err != nil {
		return err
	}

	return nil
}

func (opt *Options) GetTokenFromKeyToken(ctx context.Context, key string) (token string, err error) {

	token, err = opt.Storage.Get(key)
	if err != nil {
		return "", err
	}

	_, err = opt.Storage.Del(key)
	if err != nil {
		return "", err
	}

	return token, nil
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
