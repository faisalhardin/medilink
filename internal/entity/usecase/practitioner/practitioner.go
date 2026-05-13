package practitioner

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// PractitionerUC is the usecase contract for doctor / nurse autocomplete lookups.
// Both methods scope by the caller's institution (resolved from the JWT) — the
// caller never supplies institution_id in the query string.
type PractitionerUC interface {
	SearchDoctors(ctx context.Context, req model.DoctorSearchRequest) ([]model.DoctorSearchResult, error)
	SearchNurses(ctx context.Context, req model.NurseSearchRequest) ([]model.NurseSearchResult, error)
}
