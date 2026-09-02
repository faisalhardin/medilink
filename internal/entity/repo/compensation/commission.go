package compensation

import "context"

// PeriodCommissionTotals is the institution-level snapshot computed from commission rows.
type PeriodCommissionTotals struct {
	TotalCommission int64
	StaffCount      int64
	VisitCount      int64
}

// CommissionAggregator reads stored visit-commission rows for a payday period.
// Implemented by the commission repository card; the period usecase depends on this contract.
type CommissionAggregator interface {
	SumByPeriod(ctx context.Context, periodID int64) (PeriodCommissionTotals, error)
	DistinctVisitIDsByPeriod(ctx context.Context, periodID int64) ([]int64, error)
}
