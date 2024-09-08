package patient

import (
	"context"

	"github.com/pkg/errors"

	"github.com/faisalhardin/medilink/internal/entity/model"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
)

const (
	WrapErrMsg                = "PatientUC."
	WrapMsgRegisterNewPatient = WrapErrMsg + "RegisterNewPatient"
)

type PatientUC struct {
	PatientDB patientRepo.PatientDB
}

func NewPatientUC(u *PatientUC) *PatientUC {
	return u
}

func (u *PatientUC) RegisterNewPatient(ctx context.Context, req model.RegisterNewPatientRequest) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	err = u.PatientDB.RegisterNewPatient(ctx, &model.MstPatientInstitution{
		NIK:           req.NIK,
		Name:          req.Name,
		Sex:           req.Sex,
		DateOfBirth:   req.DateOfBirth,
		PlaceOfBirth:  req.PlaceOfBirth,
		Address:       req.Address,
		Religion:      req.Religion,
		InstitutionID: userDetail.InstitutionID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}

func (u *PatientUC) GetPatients(ctx context.Context, req model.GetPatientParams) (patients []model.GetPatientResponse, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.InstitutionID = userDetail.InstitutionID
	patients, err = u.PatientDB.GetPatients(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}

func (u *PatientUC) UpdatePatient(ctx context.Context, req model.UpdatePatientRequest) (err error) {

	err = u.PatientDB.UpdatePatient(ctx, &req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}
