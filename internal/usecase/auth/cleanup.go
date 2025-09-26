package auth

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/faisalhardin/medilink/internal/repo/auth"
)

type CleanupUC struct {
	SessionRepo *auth.SessionRepository
	stopChan    chan struct{}
	running     bool
	mu          sync.RWMutex
}

func NewCleanupUC(sessionRepo *auth.SessionRepository) *CleanupUC {
	return &CleanupUC{
		SessionRepo: sessionRepo,
		stopChan:    make(chan struct{}),
		running:     false,
	}
}

// CleanupExpiredSessions marks expired sessions as expired
func (u *CleanupUC) CleanupExpiredSessions(ctx context.Context) error {
	return u.SessionRepo.CleanupExpiredSessions(ctx)
}

// DeleteOldSessions permanently deletes old expired sessions
func (u *CleanupUC) DeleteOldSessions(ctx context.Context) error {
	return u.SessionRepo.DeleteExpiredSessions(ctx)
}

// RunCleanupJob runs the cleanup job periodically
func (u *CleanupUC) RunCleanupJob(ctx context.Context) {
	u.mu.Lock()
	if u.running {
		u.mu.Unlock()
		return
	}
	u.running = true
	u.mu.Unlock()

	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	log.Println("Cleanup job started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Cleanup job stopped (context cancelled)")
			u.mu.Lock()
			u.running = false
			u.mu.Unlock()
			return
		case <-u.stopChan:
			log.Println("Cleanup job stopped (manual stop)")
			u.mu.Lock()
			u.running = false
			u.mu.Unlock()
			return
		case <-ticker.C:
			log.Println("Running cleanup job...")

			// Mark expired sessions as expired
			if err := u.CleanupExpiredSessions(ctx); err != nil {
				log.Printf("Cleanup expired sessions error: %v", err)
			} else {
				log.Println("Successfully marked expired sessions as expired")
			}

			// Delete old expired sessions (older than 30 days)
			if err := u.DeleteOldSessions(ctx); err != nil {
				log.Printf("Delete old sessions error: %v", err)
			} else {
				log.Println("Successfully deleted old expired sessions")
			}

			log.Println("Cleanup job completed")
		}
	}
}

// Stop gracefully stops the cleanup job
func (u *CleanupUC) Stop() {
	u.mu.RLock()
	if !u.running {
		u.mu.RUnlock()
		return
	}
	u.mu.RUnlock()

	close(u.stopChan)
}

// IsRunning returns whether the cleanup job is currently running
func (u *CleanupUC) IsRunning() bool {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.running
}

// RunCleanupJobOnce runs the cleanup job once (useful for testing or manual execution)
func (u *CleanupUC) RunCleanupJobOnce(ctx context.Context) error {
	log.Println("Running cleanup job once...")

	// Mark expired sessions as expired
	if err := u.CleanupExpiredSessions(ctx); err != nil {
		log.Printf("Cleanup expired sessions error: %v", err)
		return err
	}
	log.Println("Successfully marked expired sessions as expired")

	// Delete old expired sessions (older than 30 days)
	if err := u.DeleteOldSessions(ctx); err != nil {
		log.Printf("Delete old sessions error: %v", err)
		return err
	}
	log.Println("Successfully deleted old expired sessions")

	log.Println("Cleanup job completed successfully")
	return nil
}
