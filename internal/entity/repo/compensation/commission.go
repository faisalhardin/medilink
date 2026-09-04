package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// PeriodCommissionTotals is the institution-level snapshot computed from commission rows.
type PeriodCommissionTotals struct {
	TotalCommission int64 `xorm:"total_commission"`
	StaffCount      int64 `xorm:"staff_count"`
	VisitCount      int64 `xorm:"visit_count"`
}

// StaffCommissionTotals is the per-staff commission subtotal within a payday period.
type StaffCommissionTotals struct {
	StaffID         string `xorm:"staff_id"`
	TotalCommission int64  `xorm:"total_commission"`
	VisitCount      int64  `xorm:"visit_count"`
}

// VisitCommissionWarning is a visit that exceeds a soft-warning threshold.
// PercentSum is the sum of percent-type commission_percent values only.
// CommissionIDR is the sum of all resolved commission_amount values.
// RevenueBase is the live visit-product cart sum (missing cart = 0).
type VisitCommissionWarning struct {
	VisitID       int64   `xorm:"visit_id"`
	PercentSum    float64 `xorm:"percent_sum"`
	CommissionIDR int64   `xorm:"commission_idr"`
	RevenueBase   int64   `xorm:"revenue_base"`
}

// ListVisitCommissionParams filters paginated commission reads for one staff
// in a payday period. Limit is applied only when greater than 0.
type ListVisitCommissionParams struct {
	PeriodID int64
	StaffID  string
	Limit    int
	Offset   int
}

// CommissionAggregator reads stored visit-commission rows for a payday period.
// Implemented by the commission repository card; the period usecase depends on this contract.
type CommissionAggregator interface {
	SumByPeriod(ctx context.Context, periodID int64) (PeriodCommissionTotals, error)
	DistinctVisitIDsByPeriod(ctx context.Context, periodID int64) ([]int64, error)
}

// CommissionDB is the data-access contract for mdl_trx_visit_commission.
// Mutating methods honour an active xorm session from the request context
// (see internal/library/db/xorm.GetDBSession).
type CommissionDB interface {
	CommissionAggregator

	// Upsert inserts or overwrites the live row keyed by
	// (period_id, visit_id, staff_id). Soft-deleted rows are not resurrected;
	// a new live row is inserted instead. Nil row returns an error.
	Upsert(ctx context.Context, row *model.TrxVisitCommission) error

	// ListByPeriodStaff returns non-deleted commission rows for the period and
	// staff, ordered by visit_id ASC, id ASC. total is the unpaginated match
	// count. Limit/offset apply only when Limit > 0.
	// Named to avoid colliding with PeriodDB.List on the shared Conn.
	ListByPeriodStaff(ctx context.Context, params ListVisitCommissionParams) ([]model.TrxVisitCommission, int, error)

	// SumByStaff returns per-staff commission subtotals for live rows in the
	// period, ordered by staff_id. Staff with no rows are omitted.
	SumByStaff(ctx context.Context, periodID int64) ([]StaffCommissionTotals, error)

	// SumRevenueByVisitIDs returns live visit-product cart sums keyed by
	// visit id: ROUND(SUM(COALESCE(adjusted_price, total_price))). Empty or
	// nil visitIDs returns an empty map without querying. Visits with no
	// product lines are omitted (caller treats missing as 0).
	SumRevenueByVisitIDs(ctx context.Context, visitIDs []int64) (map[int64]int64, error)

	// SoftWarningAggregates returns only visits in the period that violate a
	// soft warning: combined percent rows > 100, or total resolved IDR greater
	// than the live visit-product revenue base.
	SoftWarningAggregates(ctx context.Context, periodID int64) ([]VisitCommissionWarning, error)
}
