package patient

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type PatientDB interface {
	RegisterNewPatient(ctx context.Context, patient *model.MstPatientInstitution) (err error)
	RecordPatientVisit(ctx context.Context, request *model.MstPatientVisit) (err error)
	GetPatientVisitsRecordByPatientID(ctx context.Context, patientID int64) (mstPatientVisits []model.MstPatientVisit, err error)
	GetPatients(ctx context.Context, params model.GetPatientParams) (patients []model.GetPatientResponse, err error)
	UpdatePatient(ctx context.Context, request *model.UpdatePatientRequest) (err error)
}
