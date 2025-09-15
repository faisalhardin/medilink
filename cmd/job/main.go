package main

import (
	"context"
	"log"
	"time"

	ilog "github.com/faisalhardin/medilink/cmd/log"
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	journeyrepo "github.com/faisalhardin/medilink/internal/repo/journey"
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
	journeyDB := journeyrepo.NewJourneyDB(&journeyrepo.JourneyDB{
		DB: db,
	})

	fillShortID(journeyDB)

}

func fillShortID(journeyDB *journeyrepo.JourneyDB) {
	journeyPoints, err := journeyDB.ListJourneyPointsWithoutShortID(context.Background(), model.GetJourneyPointParams{
		CommonRequestPayload: model.CommonRequestPayload{
			Limit:  1000,
			Offset: 0,
		},
	})
	if err != nil {
		log.Fatalf("failed to list journey points: %v", err)
	}

	for _, journeyPoint := range journeyPoints {
		journeyPoint.BeforeInsert()
		err := journeyDB.UpdateJourneyPoint(context.Background(), &journeyPoint)
		if err != nil {
			log.Fatalf("failed to update journey point: %v", err)
		}
	}

}
