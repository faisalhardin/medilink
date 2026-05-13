package satusehat

import (
	"context"
	"time"

	satusehatmodel "github.com/faisalhardin/medilink/internal/entity/model/satusehat"
)

// QueueDB is the data-access contract for the SatuSehat transactional outbox.
//
// Enqueue runs inside the primary write transaction (diagnosis/anamnesa save).
// Claim / MarkDone / MarkFailed are used later by the background worker and are
// declared now so the Phase 2 usecase can depend on the full interface without
// another churn in Phase 3.
type QueueDB interface {
	// Enqueue inserts a pending outbox row. Must be called inside the caller's
	// transaction so the queue row is only visible after the primary write commits.
	Enqueue(ctx context.Context, entry *satusehatmodel.SatuSehatQueueEntry) error

	// ClaimPendingBatch atomically marks up to `limit` pending rows as
	// 'processing' (where process_after <= now) and returns them to the worker.
	// Implemented with SELECT … FOR UPDATE SKIP LOCKED so multiple worker
	// replicas never grab the same row.
	ClaimPendingBatch(ctx context.Context, limit int) ([]satusehatmodel.SatuSehatQueueEntry, error)

	// MarkDone transitions a claimed row to 'done' on successful FHIR submission.
	MarkDone(ctx context.Context, id string) error

	// MarkFailed increments attempts, records the error, and reschedules the row.
	// When attempts exceeds the worker's retry budget the caller can pass
	// status='failed'; otherwise the repo keeps it 'pending' with a new
	// process_after = nextRunAt.
	MarkFailed(ctx context.Context, id, errMsg string, nextRunAt time.Time, terminal bool) error
}
