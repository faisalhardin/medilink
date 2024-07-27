package patient

import (
	"context"
	"database/sql"

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

	param := &model.MstPatientInstitution{
		DateOfBirth: req.DateOfBirth,
	}
	if len(req.Name) > 0 {
		param.Name = sql.NullString{
			String: req.Name,
			Valid:  true,
		}
	}

	err = u.PatientDB.RegisterNewPatient(ctx, param) // &model.MstPatientInstitution{
	// NIK:          req.NIK,
	// Name:         req.Name,
	// DateOfBirth:  req.DateOfBirth,
	// PlaceOfBirth: req.PlaceOfBirth,
	// Address:      req.Address,
	// Religion:     req.Religion,
	// }

	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}
