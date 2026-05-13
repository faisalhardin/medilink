package diagnosis

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// DiagnosisDB is the data-access contract for mdl_trx_diagnosis.
// Mutating methods honour an active xorm session from the request context
// (see internal/library/db/xorm.SetDBSession) so the usecase can orchestrate
// the diff → soft-delete → insert/update transaction atomically.
type DiagnosisDB interface {
	// GetActiveByVisitID joins ref_icd10 + mdl_mst_doctor and returns every
	// non-soft-deleted diagnosis for the visit. Uses the slave DB (read path).
	GetActiveByVisitID(ctx context.Context, institutionID, visitID int64) ([]model.TrxDiagnosisWithDoctor, error)

	// GetActiveByVisitIDs returns non-soft-deleted diagnoses for any of the given
	// visits, ordered by visit_id then created_at. Empty visitIDs yields empty slice.
	GetActiveByVisitIDs(ctx context.Context, institutionID int64, visitIDs []int64) ([]model.TrxDiagnosisWithDoctor, error)

	// SoftDeleteByIDs marks the given ids as deleted_at = NOW() for this visit.
	// Must be called within a transaction; the caller asserts id ownership by
	// scoping on (institution_id, visit_id).
	SoftDeleteByIDs(ctx context.Context, institutionID, visitID int64, ids []int64) error

	// SoftDeleteByID is the single-row variant used by DELETE /v1/visit/:visit_id/diagnosis/:id.
	// Returns found=false when the id is already deleted or does not exist for this tenant.
	SoftDeleteByID(ctx context.Context, institutionID, visitID int64, diagnosisID int64) (found bool, err error)

	// BulkInsert persists new diagnosis rows and lets DB autoincrement assign IDs.
	BulkInsert(ctx context.Context, rows []model.TrxDiagnosis) error

	// BulkUpdate overwrites the mutable columns (type, case, clinical_status,
	// verification_status, prognosis, icd10_code, note, onset_date) row-by-row
	// inside the caller's transaction.
	BulkUpdate(ctx context.Context, rows []model.TrxDiagnosis) error
}
