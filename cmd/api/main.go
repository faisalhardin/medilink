package main

import (
	"fmt"
	"log"

	"github.com/faisalhardin/medilink/internal/entity/http"

	"github.com/faisalhardin/medilink/internal/config"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	institutionrepo "github.com/faisalhardin/medilink/internal/repo/institution"

	institutionUC "github.com/faisalhardin/medilink/internal/usecase/institution"

	institutionHandler "github.com/faisalhardin/medilink/internal/http/institution"

	"github.com/faisalhardin/medilink/internal/server"
	_ "github.com/lib/pq"
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
	// repo block end

	// usecase block start
	institutionUC := institutionUC.NewInstitutionUC(&institutionUC.InstitutionUC{
		InstitutionRepo: institutionDB,
	})

	// usecase block end

	// httphandler block start

	institutionHandler := institutionHandler.New(
		&institutionHandler.InstitutionHandler{
			InstitutionUC: institutionUC,
		},
	)
	// httphandler block end

	modules := server.LoadModules(&http.Handlers{
		InstitutionHandler: institutionHandler,
	})

	server := server.NewServer(server.RegisterRoutes(modules))

	err = server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
