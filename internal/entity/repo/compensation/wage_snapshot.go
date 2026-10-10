package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// WageSnapshotDB is the data-access contract for mdl_trx_wage_period_snapshot.
// Mutating methods honour an active xorm session from the request context.
type WageSnapshotDB interface {
	// List returns non-deleted snapshots for the institution, ordered by id ascending.
	List(ctx context.Context, institutionID int64) ([]model.TrxWagePeriodSnapshot, error)

	// Insert inserts a snapshot. The database assigns id.
	Insert(ctx context.Context, row *model.TrxWagePeriodSnapshot) error

	// Get returns one non-deleted snapshot for the institution.
	// CompensationPeriodUUID is filled from the payday period when that row still exists.
	Get(ctx context.Context, institutionID, id int64) (*model.TrxWagePeriodSnapshot, bool, error)

	// UpdateWageInputs writes the day counts, flat wage, and computed total.
	// NULL is stored when a value was not entered. found is false when the row is missing.
	UpdateWageInputs(ctx context.Context, row *model.TrxWagePeriodSnapshot) (found bool, err error)

	// SoftDelete marks the snapshot deleted. found is false when it is missing,
	// already deleted, or belongs to another institution.
	SoftDelete(ctx context.Context, institutionID, id int64) (found bool, err error)
}
