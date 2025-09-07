package patient

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type PatientUC interface {
	RegisterNewPatient(ctx context.Context, req model.RegisterNewPatientRequest) (newPatientResponse model.GetPatientResponse, err error)
	GetPatients(ctx context.Context, patientUUID string) (patient model.GetPatientResponse, err error)
	ListPatients(ctx context.Context, req model.GetPatientParams) (patients []model.GetPatientResponse, err error)
	UpdatePatient(ctx context.Context, req model.UpdatePatientRequest) (err error)
}
