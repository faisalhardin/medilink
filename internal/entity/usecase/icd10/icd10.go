package icd10

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// ICD10UC is the usecase contract for ICD-10 reference lookups.
// Search results are global reference data — not scoped by institution.
type ICD10UC interface {
	Search(ctx context.Context, req model.ICD10SearchRequest) ([]model.RefICD10, error)
}
