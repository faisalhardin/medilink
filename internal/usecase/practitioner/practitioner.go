package practitioner

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	practitionerrepo "github.com/faisalhardin/medilink/internal/entity/repo/practitioner"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	wrapMsgSearchDoctors = "PractitionerUC.SearchDoctors"
	wrapMsgSearchNurses  = "PractitionerUC.SearchNurses"

	// defaultSearchLimit matches BACKEND_SPEC §4 (autocomplete returns up to 20).
	defaultSearchLimit = 20
)

// PractitionerUC is the concrete implementation of the doctor / nurse lookup usecase.
type PractitionerUC struct {
	PractitionerDB practitionerrepo.PractitionerDB
}

func NewPractitionerUC(u *PractitionerUC) *PractitionerUC {
	return u
}

// SearchDoctors returns up to req.Limit (default 20) doctors in the caller's
// institution matching req.Query. Empty result is a valid 200 — autocomplete
// must not return 404 for "no matches".
func (u *PractitionerUC) SearchDoctors(ctx context.Context, req model.DoctorSearchRequest) ([]model.DoctorSearchResult, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return nil, commonerr.SetNewUnauthorizedAPICall()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	rows, err := u.PractitionerDB.SearchDoctors(ctx, userDetail.InstitutionID, req.Query, limit)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgSearchDoctors)
	}
	if rows == nil {
		rows = []model.DoctorSearchResult{}
	}
	return rows, nil
}

// SearchNurses mirrors SearchDoctors. An empty req.Role becomes nil so the repo
// skips the role filter entirely; the validator already rejects unknown roles.
func (u *PractitionerUC) SearchNurses(ctx context.Context, req model.NurseSearchRequest) ([]model.NurseSearchResult, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return nil, commonerr.SetNewUnauthorizedAPICall()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	var role *string
	if req.Role != "" {
		role = &req.Role
	}

	rows, err := u.PractitionerDB.SearchNurses(ctx, userDetail.InstitutionID, role, req.Query, limit)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgSearchNurses)
	}
	if rows == nil {
		rows = []model.NurseSearchResult{}
	}
	return rows, nil
}
