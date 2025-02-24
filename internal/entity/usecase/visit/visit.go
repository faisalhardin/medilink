package visit

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type VisitUC interface {
	InsertNewVisit(ctx context.Context, req model.InsertNewVisitRequest) (err error)
	GetPatientVisitDetail(ctx context.Context, req model.GetPatientVisitParams) (visitDetail model.GetPatientVisitDetailResponse, err error)
	ListPatientVisits(ctx context.Context, req model.GetPatientVisitParams) (visitResponse []model.ListPatientVisitBoards, err error)
	UpdatePatientVisit(ctx context.Context, req model.UpdatePatientVisitRequest) (err error)

	InsertVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (err error)
	UpdateVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (err error)
	GetVisitTouchpoint(ctx context.Context, req model.DtlPatientVisitRequest) (dtlVisit []model.DtlPatientVisit, err error)

	InsertVisitProduct(ctx context.Context, req model.InsertTrxVisitProductRequest) (err error)
}
