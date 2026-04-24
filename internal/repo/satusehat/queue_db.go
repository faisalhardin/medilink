package satusehat

import (
	"context"
	"time"

	satusehatmodel "github.com/faisalhardin/medilink/internal/entity/model/satusehat"
	satusehatrepo "github.com/faisalhardin/medilink/internal/entity/repo/satusehat"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix    = "SatuSehatQueueDB."
	WrapMsgEnqueue      = WrapErrMsgPrefix + "Enqueue"
	WrapMsgClaim        = WrapErrMsgPrefix + "ClaimPendingBatch"
	WrapMsgMarkDone     = WrapErrMsgPrefix + "MarkDone"
	WrapMsgMarkFailed   = WrapErrMsgPrefix + "MarkFailed"
	WrapMsgGenerateUUID = WrapErrMsgPrefix + "GenerateUUID"

	defaultClaimBatch = 20
	maxClaimBatch     = 100
)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewQueueDB returns a QueueDB bound to the xorm connection.
func NewQueueDB(db *xormlib.DBConnect) satusehatrepo.QueueDB {
	return &Conn{DB: db}
}

// writeSession returns the active TX session if one is on ctx; otherwise a
// fresh master-engine session. Enqueue MUST see the caller's TX — it is the
// entire point of the transactional outbox pattern.
func (c *Conn) writeSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

// Enqueue writes a pending outbox row. Status defaults to 'pending'; the
// worker will claim it once the primary TX commits.
func (c *Conn) Enqueue(ctx context.Context, entry *satusehatmodel.SatuSehatQueueEntry) error {
	if entry == nil {
		return errors.New(WrapMsgEnqueue + ": nil entry")
	}
	if entry.ID == "" {
		newID, err := uuid.NewV4()
		if err != nil {
			return errors.Wrap(err, WrapMsgGenerateUUID)
		}
		entry.ID = newID.String()
	}
	if entry.Status == "" {
		entry.Status = satusehatmodel.QueueStatusPending
	}
	if len(entry.Payload) == 0 {
		entry.Payload = []byte("{}")
	}

	const sql = `
		INSERT INTO mdl_trx_satusehat_queue
		(id, visit_id, institution_id, event_type, payload, status, attempts,
		 last_error, process_after, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?::jsonb, ?, ?, ?, COALESCE(?, NOW()), NOW(), NOW())
		RETURNING created_at, updated_at, process_after
	`

	res, err := c.writeSession(ctx).SQL(sql,
		entry.ID,
		entry.VisitID,
		entry.InstitutionID,
		entry.EventType,
		string(entry.Payload),
		entry.Status,
		entry.Attempts,
		entry.LastError,
		nullTime(entry.ProcessAfter),
	).QueryInterface()
	if err != nil {
		return errors.Wrap(err, WrapMsgEnqueue)
	}
	if len(res) > 0 {
		r := res[0]
		entry.CreatedAt = toTime(r["created_at"])
		entry.UpdatedAt = toTime(r["updated_at"])
		entry.ProcessAfter = toTime(r["process_after"])
	}
	return nil
}

// ClaimPendingBatch uses SELECT … FOR UPDATE SKIP LOCKED inside a single
// statement via a CTE so multiple worker replicas can run concurrently
// without fighting over the same rows.
func (c *Conn) ClaimPendingBatch(ctx context.Context, limit int) ([]satusehatmodel.SatuSehatQueueEntry, error) {
	if limit <= 0 {
		limit = defaultClaimBatch
	}
	if limit > maxClaimBatch {
		limit = maxClaimBatch
	}

	const sql = `
		WITH claimed AS (
			SELECT id
			FROM mdl_trx_satusehat_queue
			WHERE status = 'pending'
			  AND process_after <= NOW()
			ORDER BY process_after ASC
			FOR UPDATE SKIP LOCKED
			LIMIT ?
		)
		UPDATE mdl_trx_satusehat_queue q
		SET status     = 'processing',
		    attempts   = q.attempts + 1,
		    updated_at = NOW()
		FROM claimed c
		WHERE q.id = c.id
		RETURNING q.id, q.visit_id, q.institution_id, q.event_type, q.payload,
		          q.status, q.attempts, q.last_error, q.process_after,
		          q.created_at, q.updated_at
	`

	var rows []satusehatmodel.SatuSehatQueueEntry
	err := c.writeSession(ctx).SQL(sql, limit).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgClaim)
	}
	return rows, nil
}

// MarkDone transitions a claimed row to the terminal 'done' state.
func (c *Conn) MarkDone(ctx context.Context, id string) error {
	const sql = `
		UPDATE mdl_trx_satusehat_queue
		SET status     = 'done',
		    last_error = NULL,
		    updated_at = NOW()
		WHERE id = ?
	`
	if _, err := c.writeSession(ctx).Exec(sql, id); err != nil {
		return errors.Wrap(err, WrapMsgMarkDone)
	}
	return nil
}

// MarkFailed records the error and either terminates the row ('failed') or
// reschedules it back to 'pending' with a new process_after for retry.
func (c *Conn) MarkFailed(ctx context.Context, id, errMsg string, nextRunAt time.Time, terminal bool) error {
	status := satusehatmodel.QueueStatusPending
	if terminal {
		status = satusehatmodel.QueueStatusFailed
	}

	const sql = `
		UPDATE mdl_trx_satusehat_queue
		SET status        = ?,
		    last_error    = ?,
		    process_after = COALESCE(?, process_after),
		    updated_at    = NOW()
		WHERE id = ?
	`
	if _, err := c.writeSession(ctx).Exec(sql, status, errMsg, nullTime(nextRunAt), id); err != nil {
		return errors.Wrap(err, WrapMsgMarkFailed)
	}
	return nil
}

func nullTime(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}

func toTime(v interface{}) time.Time {
	if t, ok := v.(time.Time); ok {
		return t
	}
	return time.Time{}
}
