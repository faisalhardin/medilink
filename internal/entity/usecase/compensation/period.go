package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// CompensationPeriodUC is the payday-period lifecycle contract (create, draft, finalize, reopen, delete).
type CompensationPeriodUC interface {
	CreatePeriod(ctx context.Context, req model.CreateCompensationPeriodRequest) (model.CompensationPeriodResponse, error)
	ListPeriods(ctx context.Context, req model.ListCompensationPeriodsRequest) (model.ListCompensationPeriodsResponse, error)
	GetPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error)
	DraftPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error)
	FinalizePeriod(ctx context.Context, periodUUID string) (model.FinalizeCompensationPeriodResponse, error)
	ReopenPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error)
	DeletePeriod(ctx context.Context, periodUUID string) (model.DeleteCompensationPeriodResponse, error)
	ListPeriodStaff(ctx context.Context, periodUUID string, req model.ListCompensationPeriodStaffRequest) (model.ListCompensationPeriodStaffResponse, error)
}
