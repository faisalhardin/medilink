package main

import (
	"fmt"
	"log"

	"github.com/faisalhardin/auth-vessel/internal/config"
	authhandler "github.com/faisalhardin/auth-vessel/internal/http/auth"
	xormlib "github.com/faisalhardin/auth-vessel/internal/library/db/xorm"
	authrepo "github.com/faisalhardin/auth-vessel/internal/repo/auth"
	userrepo "github.com/faisalhardin/auth-vessel/internal/repo/user"
	"github.com/faisalhardin/auth-vessel/internal/server"
	_ "github.com/lib/pq"
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

	cfg.Vault = vault.Data

	authRepo := authrepo.New(&authrepo.Options{Cfg: cfg},
		google.New(cfg.Vault.GoogleAuth.ClientID, cfg.Vault.GoogleAuth.ClientSecret, cfg.GoogleAuthConfig.CallbackURL),
	)

	db, err := xormlib.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
		return
	}
	defer db.CloseDBConnection()

	userRepo := userrepo.New(&userrepo.Conn{
		DB: db,
	})

	authHandler := authhandler.New(&authhandler.AuthHandler{
		Cfg:      cfg,
		AuthRepo: authRepo,
		UserRepo: *userRepo,
	})

	server := server.NewServer(server.RegisterRoutes(authHandler))

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
