package compensation

import (
	"context"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// WorksheetPeriodTotals is the institution-level snapshot computed from worksheet rows.
type WorksheetPeriodTotals struct {
	TotalCommission int64 `xorm:"total_commission"`
	StaffCount      int64 `xorm:"staff_count"`
	VisitCount      int64 `xorm:"visit_count"`
}

// WorksheetDB is the data-access contract for mdl_trx_worksheet.
// Mutating methods honour an active xorm session from the request context.
type WorksheetDB interface {
	// Create inserts a new worksheet. UUID is generated when empty; status
	// defaults to pending; generate_status defaults to idle.
	Create(ctx context.Context, w *model.TrxWorksheet) error

	// GetByUUID loads a non-deleted worksheet scoped to the institution.
	// found is false when missing or already soft-deleted.
	GetByUUID(ctx context.Context, institutionID int64, uuid string) (*model.TrxWorksheet, bool, error)

	// GetByID loads a worksheet by its primary key (used in background workers).
	GetByID(ctx context.Context, id int64) (*model.TrxWorksheet, bool, error)

	// List returns non-deleted worksheets for the institution, optionally
	// filtered by staff_id and status, ordered by id DESC. Cursor is the last
	// seen id (exclusive); empty means first page. Returns at most Limit rows
	// (Limit+1 fetched internally by the usecase to detect a next page).
	// No total count.
	List(ctx context.Context, params model.ListWorksheetsRequest) ([]model.TrxWorksheet, error)

	// Update writes label, period_start, period_end, compensation_period_id
	// on the worksheet. Scoped by uuid + institution_id.
	Update(ctx context.Context, w *model.TrxWorksheet) error

	// SoftDelete marks the worksheet deleted.
	// found is false when missing or already deleted.
	SoftDelete(ctx context.Context, institutionID int64, uuid string) (found bool, err error)

	// MarkGeneratePending sets status=pending, generate_status=running,
	// generate_started_at=NOW(). Scoped by id.
	MarkGeneratePending(ctx context.Context, worksheetID int64) error

	// MarkGenerateFinished sets generate_status=succeeded/failed,
	// generate_finished_at=NOW(), generate_error (on failure),
	// and status=open on success. Scoped by id.
	MarkGenerateFinished(ctx context.Context, worksheetID int64, status model.WorksheetGenerateStatus, errMsg string) error

	// UpdateTotals writes total_commission and visit_count. Scoped by id.
	UpdateTotals(ctx context.Context, worksheetID int64, totalCommission int64, visitCount int64) error

	// ExistsOverlapping reports whether another non-deleted worksheet for the
	// same institution+staff overlaps [start, end]. excludeID is excluded (0 = no exclusion).
	ExistsOverlapping(ctx context.Context, institutionID int64, staffID string, start, end time.Time, excludeID int64) (bool, error)

	// SumByCompensationPeriod returns aggregate totals from all non-deleted
	// worksheets where compensation_period_id = compensationPeriodID.
	SumByCompensationPeriod(ctx context.Context, compensationPeriodID int64) (WorksheetPeriodTotals, error)

	// SumByStaffForCompensationPeriod returns per-staff commission totals from
	// worksheets attached to the period, ordered by staff_id.
	SumByStaffForCompensationPeriod(ctx context.Context, compensationPeriodID int64) ([]StaffCommissionTotals, error)
}
