package patient

import (
	"context"

	"github.com/pkg/errors"

	"github.com/faisalhardin/medilink/internal/entity/model"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
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

	err = u.PatientDB.RegisterNewPatient(ctx, &model.MstPatientInstitution{
		NIK:          req.NIK,
		Name:         req.Name,
		DateOfBirth:  req.DateOfBirth,
		PlaceOfBirth: req.PlaceOfBirth,
		Address:      req.Address,
		Religion:     req.Religion,
	})

	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}
