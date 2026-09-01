package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// PeriodDB is the data-access contract for mdl_trx_compensation_period.
// Mutating methods honour an active xorm session from the request context
// (see internal/library/db/xorm.GetDBSession) so the usecase can run draft,
// finalize, and reopen inside a DBTransaction.
type PeriodDB interface {
	// Create inserts a payday period. UUID is generated when empty; status
	// defaults to open when empty. Soft-delete is not applied on insert.
	Create(ctx context.Context, period *model.TrxCompensationPeriod) error

	// GetByUUID loads a non-deleted period scoped to the institution.
	// found is false when the row is missing or already soft-deleted.
	GetByUUID(ctx context.Context, institutionID int64, uuid string) (*model.TrxCompensationPeriod, bool, error)

	// List returns non-deleted periods for the institution, optionally filtered
	// by status, ordered by period_start DESC, period_end DESC. total is the
	// unpaginated match count. Limit/offset apply only when Limit > 0.
	List(ctx context.Context, params model.ListCompensationPeriodParams) ([]model.TrxCompensationPeriod, int, error)

	// UpdateStatusAndTotals writes status, wage snapshot, institution totals,
	// staff/visit counts, and drafted/finalized audit columns. Scoped by
	// uuid + institution_id. Does not enforce lifecycle transitions.
	UpdateStatusAndTotals(ctx context.Context, period *model.TrxCompensationPeriod) error

	// SoftDelete marks the period deleted. found is false when the row is
	// missing or already deleted. Open-only deletion is enforced in usecase.
	SoftDelete(ctx context.Context, institutionID int64, uuid string) (found bool, err error)
}
