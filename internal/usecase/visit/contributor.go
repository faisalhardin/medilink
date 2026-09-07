package visit

import (
	"context"
	"net/http"
	"sort"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	visituc "github.com/faisalhardin/medilink/internal/entity/usecase/visit"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

const (
	wrapVisitContributorUCPrefix  = "VisitContributorUC."
	wrapMsgListVisitContributors  = wrapVisitContributorUCPrefix + "ListVisitContributors"
	wrapMsgAddVisitContributor    = wrapVisitContributorUCPrefix + "AddVisitContributor"
	wrapMsgDeleteVisitContributor = wrapVisitContributorUCPrefix + "DeleteVisitContributor"
	errVisitNotFound              = "visit_not_found"
	msgVisitNotFound              = "visit was not found in this institution"
	errVisitCompensationLocked    = "VISIT_COMPENSATION_LOCKED"
	msgVisitCompensationLocked    = "visit compensation is locked"
	errContributorAlreadyAdded    = "contributor_already_added"
	msgContributorAlreadyAdded    = "staff is already a contributor on this visit"
	errContributorNotFound        = "contributor_not_found"
	msgContributorNotFound        = "contributor was not found on this visit"
	labelSourceProductName        = "product_name"
	labelSourceICD10Display       = "icd10_display"
)

var _ visituc.VisitContributorUC = (*VisitContributorUC)(nil)

type visitGetter interface {
	GetPatientVisitsByID(ctx context.Context, visitID int64) (model.TrxPatientVisit, error)
}

type staffGetter interface {
	GetStaffByUUID(ctx context.Context, institutionID int64, uuid string, includeInactive bool) (model.StaffWithRolesResponse, error)
}

type VisitContributorUC struct {
	PatientDB     visitGetter
	StaffDB       staffGetter
	ContributorDB compensationrepo.ContributorDB
}

func NewVisitContributorUC(uc *VisitContributorUC) *VisitContributorUC {
	return uc
}

func (u *VisitContributorUC) ListVisitContributors(ctx context.Context, visitID int64) (model.ListVisitContributorsResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.ListVisitContributorsResponse{}, commonerr.SetNewUnauthorizedAPICall()
	}

	visit, err := u.PatientDB.GetPatientVisitsByID(ctx, visitID)
	if err != nil {
		return model.ListVisitContributorsResponse{}, errors.Wrap(err, wrapMsgListVisitContributors)
	}
	if visit.ID == 0 || visit.IDMstInstitution != userDetail.InstitutionID {
		return model.ListVisitContributorsResponse{}, commonerr.SetNewError(http.StatusNotFound, errVisitNotFound, msgVisitNotFound)
	}

	detected, err := u.ContributorDB.DetectForVisit(ctx, userDetail.InstitutionID, visitID)
	if err != nil {
		return model.ListVisitContributorsResponse{}, errors.Wrap(err, wrapMsgListVisitContributors)
	}

	merged := mergeAttributions(detected)
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Name != merged[j].Name {
			return merged[i].Name < merged[j].Name
		}
		return merged[i].StaffID < merged[j].StaffID
	})

	return model.ListVisitContributorsResponse{Contributors: merged}, nil
}

func (u *VisitContributorUC) AddVisitContributor(ctx context.Context, visitID int64, staffID string) (model.AddVisitContributorResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.AddVisitContributorResponse{}, commonerr.SetNewUnauthorizedAPICall()
	}

	if _, err := uuid.Parse(staffID); err != nil {
		return model.AddVisitContributorResponse{}, commonerr.SetNewBadRequest("invalid", "Invalid Staff ID")
	}

	visit, err := u.PatientDB.GetPatientVisitsByID(ctx, visitID)
	if err != nil {
		return model.AddVisitContributorResponse{}, errors.Wrap(err, wrapMsgAddVisitContributor)
	}
	if visit.ID == 0 || visit.IDMstInstitution != userDetail.InstitutionID {
		return model.AddVisitContributorResponse{}, commonerr.SetNewError(http.StatusNotFound, errVisitNotFound, msgVisitNotFound)
	}
	if visit.CompensationLockedAt.Valid {
		return model.AddVisitContributorResponse{}, commonerr.SetNewError(http.StatusForbidden, errVisitCompensationLocked, msgVisitCompensationLocked)
	}

	staff, err := u.StaffDB.GetStaffByUUID(ctx, userDetail.InstitutionID, staffID, false)
	if err != nil {
		return model.AddVisitContributorResponse{}, err
	}

	err = u.ContributorDB.UpsertManualContributor(ctx, model.MapVisitContributor{
		VisitID:       visitID,
		StaffID:       staffID,
		InstitutionID: userDetail.InstitutionID,
		AddedBy:       userDetail.UUID,
	})
	if err != nil {
		if errors.Is(err, compensationrepo.ErrContributorAlreadyAdded) {
			return model.AddVisitContributorResponse{}, commonerr.SetNewError(http.StatusConflict, errContributorAlreadyAdded, msgContributorAlreadyAdded)
		}
		return model.AddVisitContributorResponse{}, errors.Wrap(err, wrapMsgAddVisitContributor)
	}

	return model.AddVisitContributorResponse{
		Contributor: model.VisitContributorResponse{
			StaffID:       staff.UUID,
			Name:          staff.Name,
			Source:        model.ContributionSource{Type: model.ContributionSourceTypeManual},
			AddedManually: true,
		},
	}, nil
}

func (u *VisitContributorUC) DeleteVisitContributor(ctx context.Context, visitID int64, staffID string) (model.DeleteVisitContributorResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.DeleteVisitContributorResponse{}, commonerr.SetNewUnauthorizedAPICall()
	}

	if _, err := uuid.Parse(staffID); err != nil {
		return model.DeleteVisitContributorResponse{}, commonerr.SetNewBadRequest("invalid", "Invalid Staff ID")
	}

	visit, err := u.PatientDB.GetPatientVisitsByID(ctx, visitID)
	if err != nil {
		return model.DeleteVisitContributorResponse{}, errors.Wrap(err, wrapMsgDeleteVisitContributor)
	}
	if visit.ID == 0 || visit.IDMstInstitution != userDetail.InstitutionID {
		return model.DeleteVisitContributorResponse{}, commonerr.SetNewError(http.StatusNotFound, errVisitNotFound, msgVisitNotFound)
	}
	if visit.CompensationLockedAt.Valid {
		return model.DeleteVisitContributorResponse{}, commonerr.SetNewError(http.StatusForbidden, errVisitCompensationLocked, msgVisitCompensationLocked)
	}

	deleted, err := u.ContributorDB.DeleteManualContributor(ctx, userDetail.InstitutionID, visitID, staffID)
	if err != nil {
		return model.DeleteVisitContributorResponse{}, errors.Wrap(err, wrapMsgDeleteVisitContributor)
	}
	if !deleted {
		return model.DeleteVisitContributorResponse{}, commonerr.SetNewError(http.StatusNotFound, errContributorNotFound, msgContributorNotFound)
	}

	return model.DeleteVisitContributorResponse{Success: true}, nil
}

type mergedStaff struct {
	staffID       string
	name          string
	source        compensationrepo.DetectedAttribution
	addedManually bool
}

func mergeAttributions(detected []compensationrepo.DetectedAttribution) []model.VisitContributorResponse {
	byStaff := make(map[string]*mergedStaff)

	for _, attr := range detected {
		if attr.Type == model.ContributionSourceTypeManual {
			continue
		}
		existing, ok := byStaff[attr.StaffID]
		if !ok {
			cp := attr
			byStaff[attr.StaffID] = &mergedStaff{
				staffID: attr.StaffID,
				name:    attr.Name,
				source:  cp,
			}
			continue
		}
		if beats(attr, existing.source) {
			existing.source = attr
			existing.name = attr.Name
		}
	}

	for _, attr := range detected {
		if attr.Type != model.ContributionSourceTypeManual {
			continue
		}
		existing, ok := byStaff[attr.StaffID]
		if !ok {
			cp := attr
			byStaff[attr.StaffID] = &mergedStaff{
				staffID:       attr.StaffID,
				name:          attr.Name,
				source:        cp,
				addedManually: true,
			}
			continue
		}
		existing.addedManually = true
	}

	out := make([]model.VisitContributorResponse, 0, len(byStaff))
	for _, row := range byStaff {
		out = append(out, model.VisitContributorResponse{
			StaffID:       row.staffID,
			Name:          row.name,
			Source:        contributionSourceFrom(row.source),
			AddedManually: row.addedManually,
		})
	}
	return out
}

func beats(candidate, current compensationrepo.DetectedAttribution) bool {
	cp, curp := sourcePriority(candidate.Type), sourcePriority(current.Type)
	if cp != curp {
		return cp < curp
	}
	return candidate.ClinicalRowID < current.ClinicalRowID
}

func sourcePriority(t model.ContributionSourceType) int {
	switch t {
	case model.ContributionSourceTypeProcedure:
		return 0
	case model.ContributionSourceTypeDiagnosis:
		return 1
	case model.ContributionSourceTypeAnamnesa:
		return 2
	case model.ContributionSourceTypeManual:
		return 3
	default:
		return 99
	}
}

func contributionSourceFrom(attr compensationrepo.DetectedAttribution) model.ContributionSource {
	src := model.ContributionSource{Type: attr.Type}
	switch attr.Type {
	case model.ContributionSourceTypeProcedure:
		if attr.ProcedureID != 0 {
			src.ProcedureID = null.Int64{Int64: attr.ProcedureID, Valid: true}
		}
		if attr.ProductID.Valid {
			src.ProductID = null.Int64{Int64: attr.ProductID.Int64, Valid: true}
		}
		if attr.Label.Valid {
			src.Label = null.String{String: attr.Label.String, Valid: true}
			src.LabelSource = null.String{String: labelSourceProductName, Valid: true}
		}
	case model.ContributionSourceTypeDiagnosis:
		if attr.DiagnosisID != 0 {
			src.DiagnosisID = null.Int64{Int64: attr.DiagnosisID, Valid: true}
		}
		if attr.Label.Valid {
			src.Label = null.String{String: attr.Label.String, Valid: true}
			src.LabelSource = null.String{String: labelSourceICD10Display, Valid: true}
		}
	}
	return src
}
