package main

import (
	"context"
	"flag"
	"log"

	"github.com/faisalhardin/medilink/internal/config"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/repo/auth"
	authCleanup "github.com/faisalhardin/medilink/internal/usecase/auth"
	_ "github.com/lib/pq"
)

const (
	repoName = "medilink"
)

func main() {
	var (
		runOnce = flag.Bool("once", false, "Run cleanup once and exit")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Initialize config
	cfg, err := config.New(repoName)
	if err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}

	vault, err := config.NewVault()
	if err != nil {
		log.Fatalf("Failed to init vault: %v", err)
	}
	cfg.Vault = vault.Data

	// Initialize database
	db, err := xormlib.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}
	defer db.CloseDBConnection()

	// Create session repository
	sessionRepo := auth.NewSessionRepository(db)

	// Create cleanup UC
	cleanupUC := authCleanup.NewCleanupUC(sessionRepo)

	ctx := context.Background()

	if *runOnce {
		log.Println("Running cleanup once...")
		if err := cleanupUC.RunCleanupJobOnce(ctx); err != nil {
			log.Fatalf("Cleanup failed: %v", err)
		}
		log.Println("Cleanup completed successfully")
		return
	}

	// Run cleanup job continuously
	log.Println("Starting cleanup job...")
	cleanupUC.RunCleanupJob(ctx)
}
