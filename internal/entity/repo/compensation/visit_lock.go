package compensation

import (
	"context"
	"time"
)

// VisitLockDB writes compensation lock columns on visits.
// Mutating methods honour an active xorm session from the request context
// (see internal/library/db/xorm.GetDBSession).
type VisitLockDB interface {
	// LockVisits sets compensation_period_id and compensation_locked_at on the
	// given visits. It does not unlock. Returns the number of visits locked.
	LockVisits(ctx context.Context, periodID int64, visitIDs []int64, lockedAt time.Time) (int64, error)
}
