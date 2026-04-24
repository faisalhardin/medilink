package practitioner

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// PractitionerDB is the data-access contract for doctor / nurse reference lookups.
// All methods are scoped by institution — practitioners are tenant-local master data.
type PractitionerDB interface {
	// SearchDoctors returns up to `limit` doctors in the given institution
	// matching `q` by name prefix or full-text tsvector match. Rows with a
	// linked staff account are ordered before external practitioners.
	SearchDoctors(ctx context.Context, institutionID int64, q string, limit int) ([]model.DoctorSearchResult, error)

	// SearchNurses is SearchDoctors for the nurse table; `role` filters to a
	// specific nurse_role_type ("nurse"|"midwife"|"paramedic") when non-nil.
	SearchNurses(ctx context.Context, institutionID int64, role *string, q string, limit int) ([]model.NurseSearchResult, error)

	// MissingDoctorIDs returns the subset of `ids` that do NOT exist (active=true)
	// in the given institution. Used by the diagnosis usecase for pre-TX validation.
	MissingDoctorIDs(ctx context.Context, institutionID int64, ids []string) ([]string, error)
}
