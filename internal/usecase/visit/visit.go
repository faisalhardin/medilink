package visit

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix             = "VisitUC."
	WrapMsgInsertNewVisit        = WrapErrMsgPrefix + "InsertNewVisit"
	WrapMsgGetPatientVisits      = WrapErrMsgPrefix + "GetPatientVisits"
	WrapMsgUpdatePatientVisit    = WrapErrMsgPrefix + "UpdatePatientVisit"
	WrapMsgInsertVisitTouchpoint = WrapErrMsgPrefix + "InsertVisitTouchpoint"
	WrapMsgUpdateVisitTouchpoint = WrapErrMsgPrefix + "UpdateVisitTouchpoint"
	WrapMsgGetVisitTouchpoint    = WrapErrMsgPrefix + "GetVisitTouchpoint"
)

type VisitUC struct {
	PatientDB patientRepo.PatientDB
}

func NewVisitUC(u *VisitUC) *VisitUC {
	return u
}

func (u *VisitUC) InsertNewVisit(ctx context.Context, req model.InsertNewVisitRequest) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstPatient, err := u.PatientDB.GetPatients(ctx, model.GetPatientParams{
		PatientUUIDs:  []string{req.PatientUUID},
		InstitutionID: userDetail.InstitutionID,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	if len(mstPatient) == 0 {
		return commonerr.SetNewBadRequest("patient is not found", "no patient with given uuid")
	}

	req.IDMstPatient = mstPatient[0].ID
	err = u.PatientDB.RecordPatientVisit(ctx, &req.TrxPatientVisit)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertNewVisit)
	}

	return nil
}

func (u *VisitUC) GetPatientVisits(ctx context.Context, req model.GetPatientVisitParams) (visits []model.TrxPatientVisit, err error) {
	visits, err = u.PatientDB.GetPatientVisits(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	return
}

func (u *VisitUC) UpdatePatientVisit(ctx context.Context, req model.UpdatePatientVisitRequest) (err error) {
	err = u.PatientDB.UpdatePatientVisit(ctx, req.TrxPatientVisit)
	if err != nil {
		return errors.Wrap(err, WrapMsgUpdatePatientVisit)
	}

	return
}

func (u *VisitUC) ValidatePatientVisitExist(ctx context.Context, req model.DtlPatientVisitRequest) (err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	visit, err := u.PatientDB.GetPatientVisits(ctx, model.GetPatientVisitParams{
		IDPatientVisit:   req.IDTrxPatientVisit,
		IDMstInstitution: userDetail.InstitutionID,
	})
	if err != nil {
		return err
	}
	if len(visit) == 0 {
		err = commonerr.SetNewBadRequest("invalid", "no patient visit found")
		return
	}

	return nil
}

func (u *VisitUC) InsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		return errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
	}

	err = u.PatientDB.InsertDtlPatientVisit(ctx, &model.DtlPatientVisit{
		IDTrxPatientVisit:  req.IDTrxPatientVisit,
		TouchpointName:     req.TouchpointName,
		TouchpointCategory: req.TouchpointCategory,
		Notes:              req.Notes,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertVisitTouchpoint)
	}
	return
}

func (u *VisitUC) UpdateVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		return errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
	}

	err = u.PatientDB.UpdateDtlPatientVisit(ctx, &model.DtlPatientVisit{
		ID:                 req.ID,
		TouchpointName:     req.TouchpointName,
		TouchpointCategory: req.TouchpointCategory,
		Notes:              req.Notes,
	})
	if err != nil {
		return errors.Wrap(err, WrapMsgUpdateVisitTouchpoint)
	}
	return
}

func (u *VisitUC) GetVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlVisit []model.DtlPatientVisit, err error) {

	if err = u.ValidatePatientVisitExist(ctx, req); err != nil {
		err = errors.Wrap(err, WrapMsgGetVisitTouchpoint)
		return
	}

	dtlVisit, err = u.PatientDB.GetDtlPatientVisit(ctx, model.DtlPatientVisit{
		ID:                req.ID,
		IDTrxPatientVisit: req.IDTrxPatientVisit,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetVisitTouchpoint)
		return
	}
	return
}
