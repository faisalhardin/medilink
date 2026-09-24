package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// StaffCommissionTotals is the per-staff commission subtotal within a period or worksheet.
type StaffCommissionTotals struct {
	StaffID         string `xorm:"staff_id"`
	TotalCommission int64  `xorm:"total_commission"`
	VisitCount      int64  `xorm:"visit_count"`
}

// VisitCommissionWarning is a visit that exceeds a soft-warning threshold.
type VisitCommissionWarning struct {
	VisitID       int64   `xorm:"visit_id"`
	PercentSum    float64 `xorm:"percent_sum"`
	CommissionIDR int64   `xorm:"commission_idr"`
	RevenueBase   int64   `xorm:"revenue_base"`
}

// ListVisitCommissionParams filters paginated commission reads for one worksheet.
// InstitutionID is used for visit/patient joins.
// Limit is applied only when greater than 0.
type ListVisitCommissionParams struct {
	InstitutionID int64
	WorksheetID   int64
	Limit         int
	Offset        int
}

// ListStaffDateCommissionParams filters commissions by staff and visit date range.
type ListStaffDateCommissionParams struct {
	InstitutionID int64
	StaffID       string
	Start         time.Time
	EndExclusive  time.Time
	Limit         int
	Offset        int
}

// VisitCommissionListRow is one live commission row with visit header fields.
type VisitCommissionListRow struct {
	ID                   int64                `xorm:"id"`
	WorksheetID          int64                `xorm:"worksheet_id"`
	WorksheetUUID        string               `xorm:"worksheet_uuid"`
	VisitID              int64                `xorm:"visit_id"`
	StaffID              string               `xorm:"staff_id"`
	RevenueBase          int64                `xorm:"revenue_base"`
	CommissionType       model.CommissionType `xorm:"commission_type"`
	CommissionPercent    sql.NullFloat64      `xorm:"commission_percent"`
	CommissionFlatAmount sql.NullInt64        `xorm:"commission_flat_amount"`
	CommissionAmount     int64                `xorm:"commission_amount"`
	Sources              json.RawMessage      `xorm:"sources"`
	ApprovedAt           sql.NullTime         `xorm:"approved_at"`
	PatientName          string               `xorm:"patient_name"`
	VisitDate            time.Time            `xorm:"visit_date"`
}

// CommissionDB is the data-access contract for mdl_trx_visit_commission.
// All methods are worksheet-scoped (worksheet_id replaces period_id).
// Mutating methods honour an active xorm session from the request context.
type CommissionDB interface {
	// InsertGeneratedIfMissing inserts generate-default live rows keyed by
	// (worksheet_id, visit_id). Existing live rows are left unchanged.
	// Empty rows returns 0 without querying.
	InsertGeneratedIfMissing(ctx context.Context, rows []model.TrxVisitCommission) (inserted int, err error)

	// ListByWorksheet returns non-deleted commission rows for the worksheet,
	// joined with visit/patient for patient_name and visit_date.
	// Ordered visit_id ASC, id ASC. total is the unpaginated count.
	ListByWorksheet(ctx context.Context, params ListVisitCommissionParams) ([]VisitCommissionListRow, int, error)

	// ListByStaffDateRange returns live commissions for staff whose visit create_time
	// is in [start, endExclusive), scoped to institution.
	ListByStaffDateRange(ctx context.Context, params ListStaffDateCommissionParams) ([]VisitCommissionListRow, int, error)

	// SumValidByWorksheet sums commission_amount for rows where approved_at IS NOT NULL.
	SumValidByWorksheet(ctx context.Context, worksheetID int64) (int64, error)

	// CountVisitsByWorksheet counts distinct visit_id among all live (non-deleted) rows.
	CountVisitsByWorksheet(ctx context.Context, worksheetID int64) (int64, error)

	// DistinctVisitIDsByWorksheet returns the deduplicated visit IDs for the worksheet.
	DistinctVisitIDsByWorksheet(ctx context.Context, worksheetID int64) ([]int64, error)

	// GetLiveByID returns the live commission row for id when its worksheet belongs
	// to institutionID (via join). found is false when missing or out of scope.
	GetLiveByID(ctx context.Context, institutionID, id int64) (*model.TrxVisitCommission, bool, error)

	// UpdateAssignmentAmounts updates revenue_base, commission_type, commission_percent,
	// commission_flat_amount, commission_amount, note, approved_at on the live row.
	UpdateAssignmentAmounts(ctx context.Context, row *model.TrxVisitCommission) error

	// SoftDelete soft-deletes the commission row by id.
	SoftDelete(ctx context.Context, id int64) (found bool, err error)

	// SoftDeleteByWorksheet soft-deletes all live commission rows for the worksheet.
	SoftDeleteByWorksheet(ctx context.Context, worksheetID int64) (int64, error)

	// SumRevenueByVisitIDs returns live visit-product cart sums keyed by visit id.
	SumRevenueByVisitIDs(ctx context.Context, visitIDs []int64) (map[int64]int64, error)

	// SoftWarningAggregates returns visits in the worksheet that violate soft warnings.
	SoftWarningAggregates(ctx context.Context, worksheetID int64) ([]VisitCommissionWarning, error)
}
