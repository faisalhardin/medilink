package visit

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type VisitUC interface {
	InsertNewVisit(ctx context.Context, req model.InsertNewVisitRequest) (err error)
	GetPatientVisits(ctx context.Context, req model.GetPatientVisitParams) (visits []model.TrxPatientVisit, err error)
	UpdatePatientVisit(ctx context.Context, req model.UpdatePatientVisitRequest) (err error)
}
