package compensation

import (
	"context"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	wrapErrWorksheetPrefix       = "WorksheetDB."
	wrapMsgWorksheetCreate       = wrapErrWorksheetPrefix + "Create"
	wrapMsgWorksheetGetByUUID    = wrapErrWorksheetPrefix + "GetByUUID"
	wrapMsgWorksheetGetByID      = wrapErrWorksheetPrefix + "GetByID"
	wrapMsgWorksheetList         = wrapErrWorksheetPrefix + "List"
	wrapMsgWorksheetUpdate       = wrapErrWorksheetPrefix + "Update"
	wrapMsgWorksheetSoftDelete   = wrapErrWorksheetPrefix + "SoftDelete"
	wrapMsgWorksheetMarkPending  = wrapErrWorksheetPrefix + "MarkGeneratePending"
	wrapMsgWorksheetMarkFinished = wrapErrWorksheetPrefix + "MarkGenerateFinished"
	wrapMsgWorksheetUpdateTotals = wrapErrWorksheetPrefix + "UpdateTotals"
	wrapMsgWorksheetOverlap      = wrapErrWorksheetPrefix + "ExistsOverlapping"
	wrapMsgWorksheetSumPeriod    = wrapErrWorksheetPrefix + "SumByCompensationPeriod"
	wrapMsgWorksheetSumStaff     = wrapErrWorksheetPrefix + "SumByStaffForCompensationPeriod"
)

// WorksheetConn holds the DB connection for worksheet operations.
type WorksheetConn struct {
	DB *xormlib.DBConnect
}

func (c *WorksheetConn) worksheetWriteSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

var _ compensationrepo.WorksheetDB = (*WorksheetConn)(nil)

// NewWorksheetDB returns a WorksheetDB implementation bound to the xorm connection.
func NewWorksheetDB(db *xormlib.DBConnect) compensationrepo.WorksheetDB {
	return &WorksheetConn{DB: db}
}

func (c *WorksheetConn) Create(ctx context.Context, w *model.TrxWorksheet) error {
	if w == nil {
		return errors.Wrap(errors.New("worksheet is required"), wrapMsgWorksheetCreate)
	}
	if w.UUID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return errors.Wrap(err, wrapMsgWorksheetCreate)
		}
		w.UUID = id.String()
	}
	if w.Status == "" {
		w.Status = model.WorksheetStatusOpen
	}
	if w.GenerateStatus == "" {
		w.GenerateStatus = model.WorksheetGenerateStatusIdle
	}
	_, err := c.worksheetWriteSession(ctx).
		Table(model.TrxWorksheetTableName).
		InsertOne(w)
	if err != nil {
		return errors.Wrap(err, wrapMsgWorksheetCreate)
	}
	return nil
}

func (c *WorksheetConn) GetByUUID(ctx context.Context, institutionID int64, uuid string) (*model.TrxWorksheet, bool, error) {
	const sqlText = `
		SELECT w.*,
		       p.uuid AS compensation_period_uuid
		FROM mdl_trx_worksheet w
		LEFT JOIN mdl_trx_compensation_period p
		  ON p.id = w.compensation_period_id
		 AND p.institution_id = w.institution_id
		 AND p.delete_time IS NULL
		WHERE w.uuid = ?
		  AND w.institution_id = ?
		  AND w.delete_time IS NULL
	`
	row := &model.TrxWorksheet{}
	ok, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, uuid, institutionID).Get(row)
	if err != nil {
		return nil, false, errors.Wrap(err, wrapMsgWorksheetGetByUUID)
	}
	if !ok {
		return nil, false, nil
	}
	return row, true, nil
}

func (c *WorksheetConn) GetByID(ctx context.Context, id int64) (*model.TrxWorksheet, bool, error) {
	row := &model.TrxWorksheet{}
	ok, err := c.DB.SlaveDB.Context(ctx).
		Table(model.TrxWorksheetTableName).
		Where("id = ?", id).
		Get(row)
	if err != nil {
		return nil, false, errors.Wrap(err, wrapMsgWorksheetGetByID)
	}
	if !ok {
		return nil, false, nil
	}
	return row, true, nil
}

func (c *WorksheetConn) List(ctx context.Context, params model.ListWorksheetsRequest) ([]model.TrxWorksheet, error) {
	args := []interface{}{params.InstitutionID}
	var b strings.Builder
	b.WriteString(`
		SELECT w.*,
		       p.uuid AS compensation_period_uuid
		FROM mdl_trx_worksheet w
		LEFT JOIN mdl_trx_compensation_period p
		  ON p.id = w.compensation_period_id
		 AND p.institution_id = w.institution_id
		 AND p.delete_time IS NULL
		WHERE w.institution_id = ?
		  AND w.delete_time IS NULL
	`)
	if params.StaffID != "" {
		b.WriteString(` AND w.staff_id = ?`)
		args = append(args, params.StaffID)
	}
	if params.Status != "" {
		b.WriteString(` AND w.status = ?`)
		args = append(args, string(params.Status))
	}
	if params.Cursor != "" {
		b.WriteString(` AND w.id < ?`)
		args = append(args, params.Cursor)
	}
	b.WriteString(` ORDER BY w.id DESC`)
	if params.Limit > 0 {
		b.WriteString(` LIMIT ?`)
		args = append(args, params.Limit)
	}

	rows := []model.TrxWorksheet{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(b.String(), args...).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgWorksheetList)
	}
	return rows, nil
}

func (c *WorksheetConn) Update(ctx context.Context, w *model.TrxWorksheet) error {
	if w == nil {
		return errors.Wrap(errors.New("worksheet is required"), wrapMsgWorksheetUpdate)
	}
	_, err := c.worksheetWriteSession(ctx).
		Table(model.TrxWorksheetTableName).
		Where("uuid = ?", w.UUID).
		And("institution_id = ?", w.InstitutionID).
		Cols(
			"label", "period_start", "period_end", "compensation_period_id",
			"status", "total_commission", "visit_count",
			"finalized_at", "finalized_by",
			"update_time",
		).
		Update(w)
	if err != nil {
		return errors.Wrap(err, wrapMsgWorksheetUpdate)
	}
	return nil
}

func (c *WorksheetConn) SoftDelete(ctx context.Context, institutionID int64, uuid string) (bool, error) {
	affected, err := c.worksheetWriteSession(ctx).
		Table(model.TrxWorksheetTableName).
		Where("uuid = ?", uuid).
		And("institution_id = ?", institutionID).
		Delete(&model.TrxWorksheet{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgWorksheetSoftDelete)
	}
	return affected > 0, nil
}

func (c *WorksheetConn) MarkGeneratePending(ctx context.Context, worksheetID int64) error {
	const sqlText = `
		UPDATE mdl_trx_worksheet
		SET status = 'pending',
		    generate_status = 'running',
		    generate_started_at = NOW(),
		    generate_error = NULL,
		    update_time = NOW()
		WHERE id = ?
		  AND delete_time IS NULL
		  AND status IN ('open')
	`
	res, err := c.worksheetWriteSession(ctx).Exec(sqlText, worksheetID)
	if err != nil {
		return errors.Wrap(err, wrapMsgWorksheetMarkPending)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return errors.Wrap(errors.New("worksheet not open for generate"), wrapMsgWorksheetMarkPending)
	}
	return nil
}

func (c *WorksheetConn) MarkGenerateFinished(ctx context.Context, worksheetID int64, status model.WorksheetGenerateStatus, errMsg string) error {
	var errMsgArg interface{}
	if status == model.WorksheetGenerateStatusFailed && errMsg != "" {
		errMsgArg = errMsg
	}

	const sqlText = `
		UPDATE mdl_trx_worksheet
		SET generate_status = ?,
		    status = 'open',
		    generate_finished_at = NOW(),
		    generate_error = ?,
		    update_time = NOW()
		WHERE id = ?
		  AND delete_time IS NULL
	`
	_, err := c.worksheetWriteSession(ctx).Exec(sqlText, string(status), errMsgArg, worksheetID)
	if err != nil {
		return errors.Wrap(err, wrapMsgWorksheetMarkFinished)
	}
	return nil
}

func (c *WorksheetConn) UpdateTotals(ctx context.Context, worksheetID int64, totalCommission int64, visitCount int64) error {
	const sqlText = `
		UPDATE mdl_trx_worksheet
		SET total_commission = ?,
		    visit_count = ?,
		    update_time = NOW()
		WHERE id = ?
		  AND delete_time IS NULL
	`
	_, err := c.worksheetWriteSession(ctx).Exec(sqlText, totalCommission, visitCount, worksheetID)
	if err != nil {
		return errors.Wrap(err, wrapMsgWorksheetUpdateTotals)
	}
	return nil
}

func (c *WorksheetConn) ExistsOverlapping(ctx context.Context, institutionID int64, staffID string, start, end time.Time, excludeID int64) (bool, error) {
	const sqlText = `
		SELECT COUNT(*) > 0
		FROM mdl_trx_worksheet
		WHERE institution_id = ?
		  AND staff_id = ?
		  AND delete_time IS NULL
		  AND id != ?
		  AND period_start <= ?
		  AND period_end >= ?
	`
	var exists bool
	_, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, institutionID, staffID, excludeID, end, start).Get(&exists)
	if err != nil {
		return false, errors.Wrap(err, wrapMsgWorksheetOverlap)
	}
	return exists, nil
}

func (c *WorksheetConn) SumByCompensationPeriod(ctx context.Context, compensationPeriodID int64) (compensationrepo.WorksheetPeriodTotals, error) {
	const sqlText = `
		SELECT
			COALESCE(SUM(total_commission), 0) AS total_commission,
			COUNT(DISTINCT staff_id)           AS staff_count,
			COALESCE(SUM(visit_count), 0)      AS visit_count
		FROM mdl_trx_worksheet
		WHERE compensation_period_id = ?
		  AND delete_time IS NULL
	`
	totals := compensationrepo.WorksheetPeriodTotals{}
	_, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, compensationPeriodID).Get(&totals)
	if err != nil {
		return compensationrepo.WorksheetPeriodTotals{}, errors.Wrap(err, wrapMsgWorksheetSumPeriod)
	}
	return totals, nil
}

func (c *WorksheetConn) SumByStaffForCompensationPeriod(ctx context.Context, compensationPeriodID int64) ([]compensationrepo.StaffCommissionTotals, error) {
	const sqlText = `
		SELECT
			staff_id,
			COALESCE(SUM(total_commission), 0) AS total_commission,
			COALESCE(SUM(visit_count), 0)      AS visit_count
		FROM mdl_trx_worksheet
		WHERE compensation_period_id = ?
		  AND delete_time IS NULL
		GROUP BY staff_id
		ORDER BY staff_id
	`
	rows := []compensationrepo.StaffCommissionTotals{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, compensationPeriodID).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgWorksheetSumStaff)
	}
	return rows, nil
}
