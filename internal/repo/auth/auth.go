package auth

import (
	"log"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type Config struct {
	ClientID     string
	ClientSecret string
	Key          string
	MaxAge       int
	Path         string
	IsHttpOnly   bool
	IsSecure     bool
	CallbackURL  string
}

func NewAuth(cfg *Config) *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error load env file")
	}

	store := sessions.NewCookieStore([]byte(cfg.Key))
	store.MaxAge(cfg.MaxAge)

	store.Options.Path = cfg.Path
	store.Options.HttpOnly = cfg.IsHttpOnly
	store.Options.Secure = cfg.IsSecure

	gothic.Store = store

	goth.UseProviders(
		google.New(cfg.ClientID, cfg.ClientSecret, cfg.CallbackURL),
	)

	return cfg
}
