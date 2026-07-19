package procedure

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// ProcedureUC is the procedure endpoint orchestration contract.
type ProcedureUC interface {
	// SearchICD9CM returns ICD-9-CM code search results from the reference table.
	SearchICD9CM(ctx context.Context, q string, limit int) ([]model.ICD9CMOption, error)

	// GetByVisitID returns all non-deleted procedures for a visit, ordered by rank.
	GetByVisitID(ctx context.Context, visitID int64) ([]model.ProcedureEntry, error)

	// Save atomically replaces the full procedure set for a visit:
	// rows with id != nil are updated, rows with id == nil are inserted,
	// and existing rows absent from the payload are soft-deleted.
	Save(ctx context.Context, visitID int64, req model.SaveProceduresRequest) (model.SaveProceduresSummary, error)

	// Delete soft-deletes a single procedure line by ID.
	Delete(ctx context.Context, visitID, procedureID int64) error

	// GetPatientHistory returns paginated cross-visit procedure history for a patient.
	GetPatientHistory(ctx context.Context, patientUUID string, limit, offset int) (model.ProcedureHistoryResponse, error)
}
