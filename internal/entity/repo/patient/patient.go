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
	InsertDtlPatientVisit(ctx context.Context, request *model.DtlPatientVisit) (err error)
	UpdateDtlPatientVisit(ctx context.Context, request *model.DtlPatientVisit) (err error)
	GetDtlPatientVisit(ctx context.Context, params model.DtlPatientVisit) (dtlPatientVisit []model.DtlPatientVisit, err error)
	InsertTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	UpdateTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	DeleteTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	GetTrxVisitProduct(ctx context.Context, params model.TrxVisitProduct) (trxVisitProduct []model.TrxVisitProduct, err error)
}
