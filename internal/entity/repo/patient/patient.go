package patient

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type PatientDB interface {
	RegisterNewPatient(ctx context.Context, patient *model.MstPatientInstitution) (err error)
	RecordPatientVisit(ctx context.Context, request *model.TrxPatientVisit) (err error)
	GetPatientVisitsRecordByPatientID(ctx context.Context, patientID int64) (mstPatientVisits []model.TrxPatientVisit, err error)
	GetPatients(ctx context.Context, params model.GetPatientParams) (patients []model.MstPatientInstitution, err error)
	UpdatePatient(ctx context.Context, request *model.UpdatePatientRequest) (err error)
	GetPatientVisits(ctx context.Context, params model.GetPatientVisitParams) (trxPatientVisit []model.TrxPatientVisit, err error)
	UpdatePatientVisit(ctx context.Context, trxVisit model.TrxPatientVisit) (err error)
}
