package main

import (
	"fmt"
	"log"
	"net/http"

	ilog "github.com/faisalhardin/medilink/cmd/log"
	"github.com/faisalhardin/medilink/internal/repo/auth"
	"github.com/faisalhardin/medilink/internal/repo/cache"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	httpHandler "github.com/faisalhardin/medilink/internal/entity/http"

	"github.com/faisalhardin/medilink/internal/config"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	institutionrepo "github.com/faisalhardin/medilink/internal/repo/institution"
	patientrepo "github.com/faisalhardin/medilink/internal/repo/patient"
	staffrepo "github.com/faisalhardin/medilink/internal/repo/staff"

	authUC "github.com/faisalhardin/medilink/internal/usecase/auth"
	institutionUC "github.com/faisalhardin/medilink/internal/usecase/institution"
	patientUC "github.com/faisalhardin/medilink/internal/usecase/patient"
	visituc "github.com/faisalhardin/medilink/internal/usecase/visit"

	authHandler "github.com/faisalhardin/medilink/internal/http/auth"
	institutionHandler "github.com/faisalhardin/medilink/internal/http/institution"
	patientHandler "github.com/faisalhardin/medilink/internal/http/patient"

	authmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/faisalhardin/medilink/internal/server"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

const (
	repoName = "medilink"
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

	ilog.SetupLogging(cfg)

	db, err := xormlib.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
		return
	}
	defer db.CloseDBConnection()

	store := sessions.NewCookieStore([]byte(cfg.Vault.GoogleAuth.Key))
	store.MaxAge(cfg.GoogleAuthConfig.MaxAge * 30)

	store.Options.Path = cfg.GoogleAuthConfig.CookiePath
	store.Options.HttpOnly = cfg.GoogleAuthConfig.HttpOnly
	store.Options.Secure = cfg.GoogleAuthConfig.IsProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	goth.UseProviders(
		auth.GoogleProvider(cfg),
	)

	redis := cache.New(cfg)

	// repo block start
	institutionDB := institutionrepo.NewInstitutionDB(&institutionrepo.Conn{
		DB: db,
	})

	patientDB := patientrepo.NewPatientDB(&patientrepo.Conn{
		DB: db,
	})

	staffDB := staffrepo.New(staffrepo.Conn{
		DB: db,
	})

	authRepo, err := auth.New(&auth.Options{
		Cfg:     cfg,
		Storage: redis,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	// repo block end

	// usecase block start
	institutionUC := institutionUC.NewInstitutionUC(&institutionUC.InstitutionUC{
		InstitutionRepo: institutionDB,
	})

	patientUC := patientUC.NewPatientUC(&patientUC.PatientUC{
		PatientDB: patientDB,
	})

	visitUC := visituc.NewVisitUC(&visituc.VisitUC{
		PatientDB: patientDB,
	})

	authUC := authUC.New(&authUC.AuthUC{
		Cfg:       *cfg,
		AuthRepo:  *authRepo,
		StaffRepo: staffDB,
	})

	// usecase block end

	// httphandler block start

	institutionHandler := institutionHandler.New(
		&institutionHandler.InstitutionHandler{
			InstitutionUC: institutionUC,
		},
	)

	patientHandler := patientHandler.New(&patientHandler.PatientHandler{
		PatientUC: patientUC,
		VisitUC:   visitUC,
	})

	authHandler := authHandler.New(&authHandler.AuthHandler{
		Cfg:      cfg,
		AuthUC:   authUC,
		AuthRepo: authRepo,
	})
	// httphandler block end

	// module block start
	authModule := authmodule.NewAuthModule(&authmodule.Module{
		Cfg:    cfg,
		AuthUC: authUC,
	})
	// module block end

	modules := server.LoadModules(cfg,
		&httpHandler.Handlers{
			InstitutionHandler: institutionHandler,
			PatientHandler:     patientHandler,
			AuthHandler:        authHandler,
		},
		authModule,
	)

	server := server.NewServer(server.RegisterRoutes(modules))

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}

func setGothRepo(cfg *config.Config) {

	store := sessions.NewCookieStore([]byte(cfg.Vault.GoogleAuth.Key))
	store.MaxAge(cfg.GoogleAuthConfig.MaxAge * 30)

	store.Options.Path = cfg.GoogleAuthConfig.CookiePath
	store.Options.HttpOnly = cfg.GoogleAuthConfig.HttpOnly
	store.Options.Secure = cfg.GoogleAuthConfig.IsProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	goth.UseProviders(
		auth.GoogleProvider(cfg),
	)
}
