package icd10

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// ICD10DB is the data-access contract for the ref_icd10 reference table.
// ICD-10 codes are global reference data — not scoped by institution.
type ICD10DB interface {
	// Search returns up to `limit` rows matching `q` by code prefix or
	// full-text match on the display field. Code-prefix hits come first.
	Search(ctx context.Context, q string, limit int) ([]model.RefICD10, error)

	// GetByCodes returns the rows matching the given codes (in any order).
	// Callers detect missing codes by diffing the result against the request
	// and use the returned display text to snapshot it onto write-side rows
	// (e.g. mdl_trx_diagnosis.icd10_display) before opening a transaction.
	GetByCodes(ctx context.Context, codes []string) ([]model.RefICD10, error)
}
