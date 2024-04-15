package main

import (
	"fmt"
	"log"

	"github.com/faisalhardin/auth-vessel/internal/config"
	"github.com/faisalhardin/auth-vessel/internal/repo/auth"
	"github.com/faisalhardin/auth-vessel/internal/server"
)

const (
	repoName = "vessel-auth"
)

func main() {

	// init config
	cfg, err := config.New(repoName)
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}

	vault, err := config.NewVault()
	if err != nil {
		log.Fatalf("failed to init the vault: %v", err)
	}

	cfg.Vault = vault

	auth.NewAuth(&auth.Config{
		ClientID:     cfg.Vault.GoogleAuth.ClientID,
		ClientSecret: cfg.Vault.GoogleAuth.ClientSecret,
		Key:          cfg.Vault.GoogleAuth.Key,
		MaxAge:       cfg.GoogleAuthConfig.MaxAge,
		IsSecure:     cfg.GoogleAuthConfig.IsProd,
		Path:         cfg.GoogleAuthConfig.CookiePath,
		IsHttpOnly:   cfg.GoogleAuthConfig.HttpOnly,
		CallbackURL:  cfg.GoogleAuthConfig.CallbackURL,
	})

	server := server.NewServer()

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
