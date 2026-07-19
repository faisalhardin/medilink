package procedure

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// ProcedureDB is the data-access contract for mdl_trx_visit_procedure.
// Mutating methods honour an active xorm session from the request context
// (see internal/library/db/xorm.SetDBSession) so the usecase can orchestrate
// the diff → soft-delete → insert/update transaction atomically.
type ProcedureDB interface {
	// GetActiveByVisitID returns every non-soft-deleted procedure for the visit,
	// ordered by rank ASC. Uses the slave DB (read path).
	GetActiveByVisitID(ctx context.Context, institutionID, visitID int64) ([]model.TrxVisitProcedure, error)

	// BulkInsert persists new procedure rows and lets DB autoincrement assign IDs.
	BulkInsert(ctx context.Context, rows []model.TrxVisitProcedure) error

	// BulkUpdate overwrites the mutable columns for a batch of existing rows.
	BulkUpdate(ctx context.Context, rows []model.TrxVisitProcedure) error

	// SoftDeleteByIDs marks the given ids as deleted_at = NOW(), scoped by
	// (institution_id, visit_id). Idempotent for already-deleted rows.
	SoftDeleteByIDs(ctx context.Context, institutionID, visitID int64, ids []int64) error

	// SoftDeleteByID is the single-row variant used by DELETE endpoint.
	// Returns found=false when the row is already deleted or does not exist.
	SoftDeleteByID(ctx context.Context, institutionID, visitID, procedureID int64) (found bool, err error)

	// GetPatientProcedureHistory returns paginated cross-visit procedure records
	// for a patient, ordered by visit date descending.
	GetPatientProcedureHistory(ctx context.Context, institutionID int64, patientUUID string, limit, offset int) ([]model.ProcedureHistoryEntry, int, error)
}

// ICD9CMDB is the data-access contract for mdl_ref_icd9cm.
type ICD9CMDB interface {
	// Search returns up to limit rows matching q by code prefix or display
	// substring. Leaf nodes are ranked first.
	Search(ctx context.Context, q string, limit int) ([]model.ICD9CMOption, error)

	// GetByCode returns the single row matching the given code, or nil if not found.
	GetByCode(ctx context.Context, code string) (*model.RefICD9CM, error)
}
