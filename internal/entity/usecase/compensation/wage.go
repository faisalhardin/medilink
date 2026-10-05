package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// WageUC is the staff wage configuration contract.
type WageUC interface {
	ListWages(ctx context.Context, req model.ListStaffWagesRequest) (model.ListStaffWagesResponse, error)
	UpsertWage(ctx context.Context, req model.UpsertStaffWageRequest) (model.UpsertStaffWageResponse, error)
	DeleteWage(ctx context.Context, wageID int64) (model.DeleteStaffWageResponse, error)
}
