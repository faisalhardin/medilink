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

	// MissingCodes returns the subset of `codes` that do NOT exist in ref_icd10.
	// An empty return slice means every code was found. Used by the diagnosis
	// usecase to build per-field 422 messages before opening a transaction.
	MissingCodes(ctx context.Context, codes []string) ([]string, error)
}
