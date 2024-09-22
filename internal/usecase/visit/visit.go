package visit

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix       = "VisitUC."
	WrapMsgInsertNewVisit  = WrapErrMsgPrefix + "InsertNewVisit"
	WrapGetPatientVisits   = WrapErrMsgPrefix + "GetPatientVisits"
	WrapUpdatePatientVisit = WrapErrMsgPrefix + "UpdatePatientVisit"
)

type VisitUC struct {
	PatientDB patientRepo.PatientDB
}

func NewVisitUC(u *VisitUC) *VisitUC {
	return u
}

func (u *VisitUC) InsertNewVisit(ctx context.Context, req model.InsertNewVisitRequest) (err error) {

	mstPatient, err := u.PatientDB.GetPatients(ctx, model.GetPatientParams{
		PatientUUIDs: []string{req.PatientUUID},
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
		err = errors.Wrap(err, WrapGetPatientVisits)
		return
	}

	return
}

func (u *VisitUC) UpdatePatientVisit(ctx context.Context, req model.UpdatePatientVisitRequest) (err error) {
	err = u.PatientDB.UpdatePatientVisit(ctx, req.TrxPatientVisit)
	if err != nil {
		return errors.Wrap(err, WrapUpdatePatientVisit)
	}

	return
}
