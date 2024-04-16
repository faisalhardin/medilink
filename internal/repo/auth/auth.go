package auth

import (
	"fmt"
	"net/http"

	"github.com/faisalhardin/auth-vessel/internal/config"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

type Config struct {
	Cfg *config.Config
}

func New(cfg *Config, providers ...goth.Provider) *Config {

	store := sessions.NewCookieStore([]byte(cfg.Cfg.Vault.GoogleAuth.Key))
	store.MaxAge(cfg.Cfg.GoogleAuthConfig.MaxAge)

	store.Options.Path = cfg.Cfg.GoogleAuthConfig.CookiePath
	store.Options.HttpOnly = cfg.Cfg.GoogleAuthConfig.HttpOnly
	store.Options.Secure = cfg.Cfg.GoogleAuthConfig.IsProd

	gothic.Store = store

	goth.UseProviders(
		providers...,
	)

	return cfg
}

func (h *Config) GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, r)
		return
	}

	fmt.Println(user)

}

func (h *Config) BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
	http.Redirect(w, r, h.Cfg.GoogleAuthConfig.HomepageRedirect, http.StatusFound)
}

func (h *Config) Logout(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
