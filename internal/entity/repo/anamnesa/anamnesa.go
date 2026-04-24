package anamnesa

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// AnamnesaDB is the data-access contract for mdl_trx_anamnesa.
// Each visit has at most one active anamnesa row — the unique index on
// (institution_id, visit_id) is what makes Upsert safe.
type AnamnesaDB interface {
	// GetByVisitID loads the anamnesa row for the visit (if any). The boolean
	// discriminates "not found" from a nil-error empty struct so callers can
	// branch between POST (create) and PUT/replace semantics.
	GetByVisitID(ctx context.Context, institutionID, visitID int64) (*model.TrxAnamnesa, bool, error)

	// Upsert inserts a new row or overwrites the existing one (matched by the
	// unique index). Derived fields (vs_map, vs_bmi, vs_bmi_result) are expected
	// to be pre-computed by the usecase layer per BACKEND_SPEC §6.
	// Must run inside the caller's transaction so the SatuSehat outbox row can
	// be enqueued atomically.
	Upsert(ctx context.Context, row *model.TrxAnamnesa) error
}
