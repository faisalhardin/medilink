package compensation

import (
	"context"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// WageDB is the data-access contract for mdl_mst_staff_wage.
// Mutating methods honour an active xorm session from the request context.
type WageDB interface {
	// ListActive returns non-deleted active wages for the institution.
	// staffID empty returns every active wage; otherwise only that staff.
	ListActive(ctx context.Context, institutionID int64, staffID string) ([]model.MstStaffWage, error)

	// ListLiveByStaff returns every non-deleted wage for one staff, active or not.
	// Callers use it to enforce effective-date overlap.
	ListLiveByStaff(ctx context.Context, institutionID int64, staffID string) ([]model.MstStaffWage, error)

	// Close marks a wage inactive and sets effective_to.
	Close(ctx context.Context, id, institutionID int64, effectiveTo time.Time, updatedBy string) error

	// Insert inserts a wage. The database assigns id.
	Insert(ctx context.Context, w *model.MstStaffWage) error

	// SoftDelete marks the wage deleted. found is false when it is missing,
	// already deleted, or belongs to another institution.
	SoftDelete(ctx context.Context, institutionID, id int64) (found bool, err error)
}
