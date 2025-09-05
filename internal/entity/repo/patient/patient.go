package patient

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type PatientDB interface {
	RegisterNewPatient(ctx context.Context, patient *model.MstPatientInstitution) (err error)
	RecordPatientVisit(ctx context.Context, request *model.TrxPatientVisit) (err error)
	GetPatientVisitsRecordByPatientID(ctx context.Context, patientID int64) (mstPatientVisits []model.TrxPatientVisit, err error)
	GetPatientVisitsByID(ctx context.Context, visitID int64) (mstPatientVisits model.TrxPatientVisit, err error)
	GetPatientByID(ctx context.Context, patientID int64) (patient model.MstPatientInstitution, err error)
	GetPatients(ctx context.Context, params model.GetPatientParams) (patients []model.MstPatientInstitution, err error)
	UpdatePatient(ctx context.Context, request *model.UpdatePatientRequest) (err error)
	GetPatientVisits(ctx context.Context, params model.GetPatientVisitParams) (trxPatientVisit []model.GetPatientVisitResponse, err error)
	UpdatePatientVisit(ctx context.Context, updateRequest model.UpdatePatientVisitRequest) (trxVisit model.TrxPatientVisit, err error)
	DeletePatientVisit(ctx context.Context, request *model.TrxPatientVisit) (err error)
	InsertDtlPatientVisit(ctx context.Context, request *model.DtlPatientVisit) (err error)
	UpdateDtlPatientVisit(ctx context.Context, request *model.DtlPatientVisit) (err error)
	GetDtlPatientVisit(ctx context.Context, params model.GetDtlPatientVisitParams) (dtlPatientVisit []model.DtlPatientVisit, err error)
	GetDtlPatientVisitByID(ctx context.Context, id int64) (dtlPatientVisit model.DtlPatientVisit, err error)
	InsertTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	UpdateTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	UpsertTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	DeleteTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error)
	GetTrxVisitProduct(ctx context.Context, params model.GetVisitProductRequest) (trxVisitProduct []model.TrxVisitProduct, err error)
}
