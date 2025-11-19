package patient

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/faisalhardin/medilink/internal/entity/model"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
)

const (
	WrapErrMsg                = "PatientUC."
	WrapMsgRegisterNewPatient = WrapErrMsg + "RegisterNewPatient"

	defaultLimit = 10
)

type PatientUC struct {
	PatientDB patientRepo.PatientDB
}

func NewPatientUC(u *PatientUC) *PatientUC {
	return u
}

func (u *PatientUC) RegisterNewPatient(ctx context.Context, req model.RegisterNewPatientRequest) (newPatientResponse model.GetPatientResponse, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	newPatient := model.MstPatientInstitution{
		NIK:           req.NIK,
		Name:          req.Name,
		Sex:           req.Sex,
		DateOfBirth:   req.DateOfBirth.Time(),
		PlaceOfBirth:  req.PlaceOfBirth,
		Address:       req.Address,
		Religion:      req.Religion,
		InstitutionID: userDetail.InstitutionID,
	}

	err = u.PatientDB.RegisterNewPatient(ctx, &newPatient)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	newPatientResponse = model.GetPatientResponse{
		UUID:         newPatient.UUID,
		NIK:          newPatient.NIK,
		Name:         newPatient.Name,
		PlaceOfBirth: newPatient.PlaceOfBirth,
		DateOfBirth:  newPatient.DateOfBirth,
		Address:      newPatient.Address,
		Religion:     newPatient.Religion,
		PhoneNumber:  newPatient.PhoneNumber,
		Sex:          newPatient.Sex,
	}

	return newPatientResponse, nil
}

func (u *PatientUC) GetPatients(ctx context.Context, patientUUID string) (patient model.GetPatientResponse, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstPatients, err := u.PatientDB.GetPatients(ctx, model.GetPatientParams{
		PatientUUIDs:  []string{patientUUID},
		InstitutionID: userDetail.InstitutionID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	if len(mstPatients) == 0 {
		err = commonerr.SetNewBadRequest("patient not found", fmt.Sprintf("there is no patient with registered with uuid = %v", patientUUID))
		return
	}

	patient = model.GetPatientResponse{
		UUID:         mstPatients[0].UUID,
		NIK:          mstPatients[0].NIK,
		Name:         mstPatients[0].Name,
		PlaceOfBirth: mstPatients[0].PlaceOfBirth,
		DateOfBirth:  mstPatients[0].DateOfBirth,
		Address:      mstPatients[0].Address,
		Religion:     mstPatients[0].Religion,
		Sex:          mstPatients[0].Sex,
	}

	return
}

func (u *PatientUC) ListPatients(ctx context.Context, req model.GetPatientParams) (patients []model.GetPatientResponse, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	if req.CommonRequestPayload.Limit == 0 {
		req.CommonRequestPayload.Limit = defaultLimit
	}

	req.InstitutionID = userDetail.InstitutionID
	mstPatients, err := u.PatientDB.GetPatients(ctx, req)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	for _, patient := range mstPatients {
		patients = append(patients, model.GetPatientResponse{
			UUID:         patient.UUID,
			NIK:          patient.NIK,
			Name:         patient.Name,
			PlaceOfBirth: patient.PlaceOfBirth,
			DateOfBirth:  patient.DateOfBirth,
			Address:      patient.Address,
			Religion:     patient.Religion,
			PhoneNumber:  patient.PhoneNumber,
			Sex:          patient.Sex,
		})
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
