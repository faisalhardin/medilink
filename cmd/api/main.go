package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	ilog "github.com/faisalhardin/medilink/cmd/log"
	"github.com/faisalhardin/medilink/internal/repo/auth"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"

	httpHandler "github.com/faisalhardin/medilink/internal/entity/http"

	"github.com/faisalhardin/medilink/internal/config"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	inmemory "github.com/faisalhardin/medilink/internal/repo/cache/inmemory"
	institutionrepo "github.com/faisalhardin/medilink/internal/repo/institution"
	journeyrepo "github.com/faisalhardin/medilink/internal/repo/journey"
	custjourneyrepo "github.com/faisalhardin/medilink/internal/repo/journey/customerjourney"
	patientrepo "github.com/faisalhardin/medilink/internal/repo/patient"
	productrepo "github.com/faisalhardin/medilink/internal/repo/product"
	staffrepo "github.com/faisalhardin/medilink/internal/repo/staff"

	authUC "github.com/faisalhardin/medilink/internal/usecase/auth"
	institutionUC "github.com/faisalhardin/medilink/internal/usecase/institution"
	journeyuc "github.com/faisalhardin/medilink/internal/usecase/journey"
	patientUC "github.com/faisalhardin/medilink/internal/usecase/patient"
	productuc "github.com/faisalhardin/medilink/internal/usecase/product"
	visituc "github.com/faisalhardin/medilink/internal/usecase/visit"

	authHandler "github.com/faisalhardin/medilink/internal/http/auth"
	institutionHandler "github.com/faisalhardin/medilink/internal/http/institution"
	journeyhandler "github.com/faisalhardin/medilink/internal/http/journey"
	patientHandler "github.com/faisalhardin/medilink/internal/http/patient"
	producthandler "github.com/faisalhardin/medilink/internal/http/product"

	mwmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/faisalhardin/medilink/internal/server"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

const (
	repoName = "medilink"
)

func main() {

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}
	time.Local = loc

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

	inMemoryCaching := inmemory.New(inmemory.Options{
		MaxIdle:   cfg.Redis.MaxIdle,
		MaxActive: cfg.Redis.MaxActive,
		Timeout:   cfg.Redis.TimeOutInSecond,
		Wait:      true,
	})

	// repo block start
	transaction := xormlib.NewTransaction(db)

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
		Storage: inMemoryCaching,
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	productDB := productrepo.NewProductDB(&productrepo.Conn{
		DB: db,
	})

	journeyDB := journeyrepo.NewJourneyDB(&journeyrepo.JourneyDB{
		DB: db,
	})

	customerJourneyDB := custjourneyrepo.NewUserJourneyDB(&custjourneyrepo.UserJourneyDB{
		JourneyDB: journeyDB,
	})

	// repo block end

	// usecase block start
	institutionUC := institutionUC.NewInstitutionUC(&institutionUC.InstitutionUC{
		InstitutionRepo: institutionDB,
		Transaction:     transaction,
	})

	patientUC := patientUC.NewPatientUC(&patientUC.PatientUC{
		PatientDB: patientDB,
	})

	visitUC := visituc.NewVisitUC(&visituc.VisitUC{
		PatientDB:       patientDB,
		InstitutionRepo: institutionDB,
		Transaction:     transaction,
		JourneyDB:       customerJourneyDB,
	})

	authUC := authUC.New(&authUC.AuthUC{
		Cfg:         *cfg,
		AuthRepo:    *authRepo,
		StaffRepo:   staffDB,
		JourneyRepo: journeyDB,
	})

	productUC := productuc.NewProductUC(&productuc.ProductUC{
		ProductDB: productDB,
	})

	journeyUC := journeyuc.NewJourneyUC(&journeyuc.JourneyUC{
		JourneyDB:   customerJourneyDB,
		PatientDB:   patientDB,
		Transaction: transaction,
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

	productHandler := producthandler.New(&producthandler.ProductHandler{
		ProductUC: productUC,
	})

	journeyHandler := journeyhandler.New(&journeyhandler.JourneyHandler{
		JourneyUC: journeyUC,
	})
	// httphandler block end

	// module block start
	middlewareModule := mwmodule.NewMiddlewareModule(&mwmodule.Module{
		Cfg:    cfg,
		AuthUC: authUC,
	})
	// module block end

	modules := server.LoadModules(cfg,
		&httpHandler.Handlers{
			InstitutionHandler: institutionHandler,
			PatientHandler:     patientHandler,
			AuthHandler:        authHandler,
			ProductHandler:     productHandler,
			JourneyHandler:     journeyHandler,
		},
		middlewareModule,
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
