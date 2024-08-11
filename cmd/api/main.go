package main

import (
	"fmt"
	"log"

	ilog "github.com/faisalhardin/medilink/cmd/log"
	"github.com/faisalhardin/medilink/internal/repo/auth"

	httpHandler "github.com/faisalhardin/medilink/internal/entity/http"

	"github.com/faisalhardin/medilink/internal/config"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	institutionrepo "github.com/faisalhardin/medilink/internal/repo/institution"
	patientrepo "github.com/faisalhardin/medilink/internal/repo/patient"

	authUC "github.com/faisalhardin/medilink/internal/usecase/auth"
	institutionUC "github.com/faisalhardin/medilink/internal/usecase/institution"
	patientUC "github.com/faisalhardin/medilink/internal/usecase/patient"

	authHandler "github.com/faisalhardin/medilink/internal/http/auth"
	institutionHandler "github.com/faisalhardin/medilink/internal/http/institution"
	patientHandler "github.com/faisalhardin/medilink/internal/http/patient"

	"github.com/faisalhardin/medilink/internal/server"
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

	// repo block start
	institutionDB := institutionrepo.NewInstitutionDB(&institutionrepo.Conn{
		DB: db,
	})

	patientDB := patientrepo.NewPatientDB(&patientrepo.Conn{
		DB: db,
	})

	authRepo, err := auth.New(&auth.Options{
		Cfg: cfg,
		Str: auth.MockRedisClient{},
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

	authUC := authUC.New(&authUC.AuthUC{
		Cfg:      *cfg,
		AuthRepo: *authRepo,
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
	})

	authHandler := authHandler.New(&authHandler.AuthHandler{
		Cfg:    cfg,
		AuthUC: authUC,
	})
	// httphandler block end

	modules := server.LoadModules(&httpHandler.Handlers{
		InstitutionHandler: institutionHandler,
		PatientHandler:     patientHandler,
		AuthHandler:        authHandler,
	})

	server := server.NewServer(server.RegisterRoutes(modules))

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
