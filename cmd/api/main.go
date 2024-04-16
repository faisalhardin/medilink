package main

import (
	"fmt"
	"log"

	"github.com/faisalhardin/auth-vessel/internal/config"
	authhandler "github.com/faisalhardin/auth-vessel/internal/http/auth"
	authrepo "github.com/faisalhardin/auth-vessel/internal/repo/auth"
	"github.com/faisalhardin/auth-vessel/internal/server"
	"github.com/markbates/goth/providers/google"
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

	authRepo := authrepo.New(&authrepo.Config{Cfg: cfg},
		google.New(cfg.Vault.GoogleAuth.ClientID, cfg.Vault.GoogleAuth.ClientSecret, cfg.GoogleAuthConfig.CallbackURL),
	)

	authHandler := authhandler.New(&authhandler.AuthHandler{
		Cfg:      cfg,
		AuthRepo: authRepo,
	})

	server := server.NewServer(server.RegisterRoutes(authHandler))

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
