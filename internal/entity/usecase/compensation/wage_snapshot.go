package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// WageSnapshotUC is the wage snapshot contract.
type WageSnapshotUC interface {
	Generate(ctx context.Context, req model.GenerateStaffWageSnapshotsRequest) (model.StaffWageSnapshotsResponse, error)
	List(ctx context.Context) (model.StaffWageSnapshotsResponse, error)
	Update(ctx context.Context, id int64, req model.UpdateStaffWageSnapshotRequest) (model.StaffWageSnapshotResponse, error)
	Delete(ctx context.Context, id int64) (model.DeleteStaffWageSnapshotResponse, error)
}
