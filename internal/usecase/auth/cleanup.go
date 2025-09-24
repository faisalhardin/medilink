package auth

import (
	"context"
	"time"

	"github.com/faisalhardin/medilink/internal/repo/auth"
)

type CleanupUC struct {
	SessionRepo *auth.SessionRepository
}

func NewCleanupUC(sessionRepo *auth.SessionRepository) *CleanupUC {
	return &CleanupUC{
		SessionRepo: sessionRepo,
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
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Mark expired sessions as expired
			if err := u.CleanupExpiredSessions(ctx); err != nil {
				// Log error but continue
				continue
			}

			// Delete old expired sessions (older than 30 days)
			if err := u.DeleteOldSessions(ctx); err != nil {
				// Log error but continue
				continue
			}
		}
	}
}
