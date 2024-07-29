package main

import (
	"fmt"
	"log"

	ilog "github.com/faisalhardin/medilink/cmd/log"

	httpHandler "github.com/faisalhardin/medilink/internal/entity/http"

	"github.com/faisalhardin/medilink/internal/config"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	institutionrepo "github.com/faisalhardin/medilink/internal/repo/institution"
	patientrepo "github.com/faisalhardin/medilink/internal/repo/patient"

	institutionUC "github.com/faisalhardin/medilink/internal/usecase/institution"
	patientUC "github.com/faisalhardin/medilink/internal/usecase/patient"

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
	// repo block end

	// usecase block start
	institutionUC := institutionUC.NewInstitutionUC(&institutionUC.InstitutionUC{
		InstitutionRepo: institutionDB,
	})

	patientUC := patientUC.NewPatientUC(&patientUC.PatientUC{
		PatientDB: patientDB,
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
	// httphandler block end

	modules := server.LoadModules(&httpHandler.Handlers{
		InstitutionHandler: institutionHandler,
		PatientHandler:     patientHandler,
	})

	server := server.NewServer(server.RegisterRoutes(modules))

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
