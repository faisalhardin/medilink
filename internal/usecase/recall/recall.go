package recall

import (
	"context"
	"fmt"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	recallRepo "github.com/faisalhardin/medilink/internal/entity/repo/recall"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	wrapMsgCreateRecall           = "RecallUC.CreateRecall"
	wrapMsgUpdateRecall           = "RecallUC.UpdateRecall"
	wrapMsgGetNextRecallByPatient = "RecallUC.GetNextRecallByPatient"
	wrapMsgListRecalls            = "RecallUC.ListRecalls"
	defaultListLimit              = 50
)

type RecallUC struct {
	RecallDB  recallRepo.RecallDB
	PatientDB patientRepo.PatientDB
}

func NewRecallUC(u *RecallUC) *RecallUC {
	return u
}

func (u *RecallUC) CreateRecall(ctx context.Context, req model.CreateRecallRequest) (model.RecallResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.RecallResponse{}, commonerr.SetNewUnauthorizedAPICall()
	}

	patient, err := u.PatientDB.GetPatientByParams(ctx, model.MstPatientInstitution{
		UUID:          req.PatientUUID,
		InstitutionID: userDetail.InstitutionID,
	})
	if err != nil {
		return model.RecallResponse{}, errors.Wrap(err, wrapMsgCreateRecall)
	}
	if patient.ID == 0 {
		return model.RecallResponse{}, commonerr.SetNewBadRequest("patient not found",
			fmt.Sprintf("no patient with uuid %s in this institution", req.PatientUUID))
	}

	rec := model.TrxRecall{
		IDMstPatient:        patient.ID,
		IDMstInstitution:    userDetail.InstitutionID,
		ScheduledAt:         req.ScheduledAt,
		RecallType:          req.RecallType,
		Notes:               req.Notes,
		CreatedByIDMstStaff: userDetail.UserID,
		IDTrxPatientVisit:   req.IDTrxPatientVisit,
	}
	if err := u.RecallDB.Insert(ctx, &rec); err != nil {
		return model.RecallResponse{}, errors.Wrap(err, wrapMsgCreateRecall)
	}

	return model.RecallResponse{
		ID:                rec.ID,
		PatientUUID:       patient.UUID,
		PatientName:       patient.Name,
		ScheduledAt:       rec.ScheduledAt.Time(),
		RecallType:        rec.RecallType,
		Notes:             rec.Notes,
		IDTrxPatientVisit: rec.IDTrxPatientVisit,
		CreateTime:        rec.CreateTime,
	}, nil
}

func (u *RecallUC) UpdateRecall(ctx context.Context, req model.UpdateRecallRequest) error {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return commonerr.SetNewUnauthorizedAPICall()
	}

	existing, err := u.RecallDB.GetByID(ctx, req.IDMstRecall, userDetail.InstitutionID)
	if err != nil {
		return errors.Wrap(err, wrapMsgUpdateRecall)
	}
	if existing.TrxRecall.ID == 0 {
		return commonerr.SetNewBadRequest("recall not found",
			fmt.Sprintf("no recall with id %d in this institution", req.IDMstRecall))
	}

	if err := u.RecallDB.Update(ctx, req.IDMstRecall, userDetail.InstitutionID, req); err != nil {
		return errors.Wrap(err, wrapMsgUpdateRecall)
	}

	return nil
}

func (u *RecallUC) GetNextRecallByPatient(ctx context.Context, patientUUID string) (model.NextRecallResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.NextRecallResponse{}, commonerr.SetNewUnauthorizedAPICall()
	}

	rec, found, err := u.RecallDB.GetNextByPatient(ctx, patientUUID, userDetail.InstitutionID)
	if err != nil {
		return model.NextRecallResponse{}, errors.Wrap(err, wrapMsgGetNextRecallByPatient)
	}

	if !found {
		// Distinguish "patient not found" from "no upcoming recall"
		patient, _ := u.PatientDB.GetPatientByParams(ctx, model.MstPatientInstitution{
			UUID:          patientUUID,
			InstitutionID: userDetail.InstitutionID,
		})
		if patient.ID == 0 {
			return model.NextRecallResponse{}, commonerr.SetNewBadRequest("patient not found",
				fmt.Sprintf("no patient with uuid %s in this institution", patientUUID))
		}
		return model.NextRecallResponse{RecallResponse: model.RecallResponse{}, HasNext: false}, nil
	}

	return model.NextRecallResponse{
		RecallResponse: model.RecallResponse{
			ID:                rec.TrxRecall.ID,
			PatientUUID:       rec.MstPatientInstitution.UUID,
			PatientName:       rec.MstPatientInstitution.Name,
			ScheduledAt:       rec.TrxRecall.ScheduledAt.Time(),
			RecallType:        rec.TrxRecall.RecallType,
			Notes:             rec.TrxRecall.Notes,
			IDTrxPatientVisit: rec.TrxRecall.IDTrxPatientVisit,
			CreateTime:        rec.TrxRecall.CreateTime,
		},
		HasNext: false,
	}, nil
}

func (u *RecallUC) ListRecalls(ctx context.Context, params model.GetRecallParams) ([]model.RecallResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return nil, commonerr.SetNewUnauthorizedAPICall()
	}

	if params.Limit <= 0 {
		params.Limit = defaultListLimit
	}
	// Default to upcoming recalls (scheduled from now onward) when no time range given
	if params.FromTime.Time().IsZero() {
		params.FromTime = model.Time(time.Now())
	}

	list, err := u.RecallDB.ListUpcoming(ctx, params, userDetail.InstitutionID)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgListRecalls)
	}

	result := make([]model.RecallResponse, 0, len(list))
	for _, r := range list {
		result = append(result, model.RecallResponse{
			ID:                r.TrxRecall.ID,
			PatientUUID:       r.MstPatientInstitution.UUID,
			PatientName:       r.MstPatientInstitution.Name,
			ScheduledAt:       r.TrxRecall.ScheduledAt.Time(),
			RecallType:        r.RecallType,
			Notes:             r.Notes,
			IDTrxPatientVisit: r.IDTrxPatientVisit,
			CreateTime:        r.TrxRecall.CreateTime,
		})
	}
	return result, nil
}
