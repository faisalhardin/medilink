package auth

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/config"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type Options struct {
	Cfg *config.Config
}

func New(opt *Options, providers ...goth.Provider) *Options {

	store := sessions.NewCookieStore([]byte(opt.Cfg.Vault.GoogleAuth.Key))
	store.MaxAge(opt.Cfg.GoogleAuthConfig.MaxAge)

	store.Options.Path = opt.Cfg.GoogleAuthConfig.CookiePath
	store.Options.HttpOnly = opt.Cfg.GoogleAuthConfig.HttpOnly
	store.Options.Secure = opt.Cfg.GoogleAuthConfig.IsProd

	gothic.Store = store

	goth.UseProviders(
		providers...,
	)

	return opt
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
