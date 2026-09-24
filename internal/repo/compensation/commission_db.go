package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	wrapMsgCommissionInsertGenerated = "CommissionDB.InsertGeneratedIfMissing"
	insertGeneratedChunkSize         = 500
	// worksheet_id, visit_id, staff_id, revenue_base, commission_type, commission_percent,
	// commission_flat_amount, commission_amount, sources, note, included_manually, approved_at,
	// create_time, update_time
	insertGeneratedRowSQL  = `(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NOW(), NOW())`
	insertGeneratedHeadSQL = `
		INSERT INTO mdl_trx_visit_commission (
			worksheet_id, visit_id, staff_id,
			revenue_base, commission_type, commission_percent, commission_flat_amount,
			commission_amount, sources, note, included_manually, approved_at,
			create_time, update_time
		) VALUES `
	insertGeneratedTailSQL = `
		ON CONFLICT (worksheet_id, visit_id) WHERE delete_time IS NULL
		DO NOTHING
		RETURNING id
	`
	wrapMsgCommissionList          = "CommissionDB.ListByWorksheet"
	wrapMsgCommissionListCount     = "CommissionDB.ListByWorksheetCount"
	wrapMsgCommissionSumValid      = "CommissionDB.SumValidByWorksheet"
	wrapMsgCommissionDistinctVisit = "CommissionDB.DistinctVisitIDsByWorksheet"
	wrapMsgCommissionSumRevenue    = "CommissionDB.SumRevenueByVisitIDs"
	wrapMsgCommissionWarnings      = "CommissionDB.SoftWarningAggregates"
	wrapMsgCommissionGetLiveByID   = "CommissionDB.GetLiveByID"
	wrapMsgCommissionUpdateAmounts = "CommissionDB.UpdateAssignmentAmounts"
	wrapMsgCommissionSoftDelete    = "CommissionDB.SoftDelete"
	wrapMsgCommissionSoftDeleteWS  = "CommissionDB.SoftDeleteByWorksheet"

	listByWorksheetSQL = `
		SELECT
			c.id,
			c.worksheet_id,
			w.uuid AS worksheet_uuid,
			c.visit_id,
			c.staff_id,
			c.revenue_base,
			c.commission_type,
			c.commission_percent,
			c.commission_flat_amount,
			c.commission_amount,
			c.sources,
			c.approved_at,
			COALESCE(p.name, '') AS patient_name,
			v.create_time AS visit_date
		FROM mdl_trx_visit_commission c
		INNER JOIN mdl_trx_worksheet w
			ON w.id = c.worksheet_id
			AND w.delete_time IS NULL
		LEFT JOIN mdl_trx_patient_visit v
			ON v.id = c.visit_id
			AND v.id_mst_institution = ?
			AND v.delete_time IS NULL
		LEFT JOIN mdl_mst_patient_institution p
			ON p.id = v.id_mst_patient
			AND p.delete_time IS NULL
		WHERE c.worksheet_id = ?
		  AND c.delete_time IS NULL
		ORDER BY c.visit_id ASC, c.id ASC
	`
	listByWorksheetCountSQL = `
		SELECT COUNT(*)
		FROM mdl_trx_visit_commission c
		WHERE c.worksheet_id = ?
		  AND c.delete_time IS NULL
	`
)

// CommissionConn holds the DB connection for commission operations.
type CommissionConn struct {
	DB *xormlib.DBConnect
}

// commissionWriteSession returns the active TX session if one was put on ctx
// by the usecase; otherwise returns a fresh master-engine session.
func (c *CommissionConn) commissionWriteSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

var _ compensationrepo.CommissionDB = (*CommissionConn)(nil)

// NewCommissionDB returns a CommissionDB implementation bound to the xorm connection.
func NewCommissionDB(db *xormlib.DBConnect) compensationrepo.CommissionDB {
	return &CommissionConn{DB: db}
}

func (c *CommissionConn) InsertGeneratedIfMissing(ctx context.Context, rows []model.TrxVisitCommission) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}

	inserted := 0
	for start := 0; start < len(rows); start += insertGeneratedChunkSize {
		end := start + insertGeneratedChunkSize
		if end > len(rows) {
			end = len(rows)
		}
		n, err := c.insertGeneratedChunk(ctx, rows[start:end])
		if err != nil {
			return 0, err
		}
		inserted += n
	}
	return inserted, nil
}

func (c *CommissionConn) insertGeneratedChunk(ctx context.Context, rows []model.TrxVisitCommission) (int, error) {
	var b strings.Builder
	b.WriteString(insertGeneratedHeadSQL)
	args := make([]interface{}, 0, len(rows)*11)
	for i := range rows {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(insertGeneratedRowSQL)
		row := &rows[i]
		args = append(args,
			row.WorksheetID,
			row.VisitID,
			row.StaffID,
			row.RevenueBase,
			row.CommissionType,
			nullFloat64(row.CommissionPercent),
			nullInt64(row.CommissionFlatAmount),
			row.CommissionAmount,
			jsonbOrNil(row.Sources),
			nullString(row.Note),
			row.IncludedManually,
		)
	}
	b.WriteString(insertGeneratedTailSQL)

	res, err := c.commissionWriteSession(ctx).SQL(b.String(), args...).QueryInterface()
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgCommissionInsertGenerated)
	}
	return len(res), nil
}

func (c *CommissionConn) ListByWorksheet(ctx context.Context, params compensationrepo.ListVisitCommissionParams) ([]compensationrepo.VisitCommissionListRow, int, error) {
	var total64 int64
	_, err := c.DB.SlaveDB.Context(ctx).SQL(listByWorksheetCountSQL, params.WorksheetID).Get(&total64)
	if err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgCommissionListCount)
	}

	sqlText := listByWorksheetSQL
	args := []interface{}{params.InstitutionID, params.WorksheetID}
	if params.Limit > 0 {
		sqlText += " LIMIT ? OFFSET ?"
		args = append(args, params.Limit, params.Offset)
	}

	rows := []compensationrepo.VisitCommissionListRow{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, args...).Find(&rows); err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgCommissionList)
	}
	return rows, int(total64), nil
}

func (c *CommissionConn) ListByStaffDateRange(ctx context.Context, params compensationrepo.ListStaffDateCommissionParams) ([]compensationrepo.VisitCommissionListRow, int, error) {
	const countSQL = `
		SELECT COUNT(*)
		FROM mdl_trx_visit_commission c
		JOIN mdl_trx_patient_visit v
			ON v.id = c.visit_id
			AND v.id_mst_institution = ?
			AND v.delete_time IS NULL
		WHERE c.staff_id = ?
		  AND v.create_time >= ?
		  AND v.create_time < ?
		  AND c.delete_time IS NULL
	`
	const listSQL = `
		SELECT
			c.id,
			c.worksheet_id,
			w.uuid AS worksheet_uuid,
			c.visit_id,
			c.staff_id,
			c.revenue_base,
			c.commission_type,
			c.commission_percent,
			c.commission_flat_amount,
			c.commission_amount,
			c.sources,
			c.approved_at,
			COALESCE(p.name, '') AS patient_name,
			v.create_time AS visit_date
		FROM mdl_trx_visit_commission c
		INNER JOIN mdl_trx_worksheet w
			ON w.id = c.worksheet_id
			AND w.delete_time IS NULL
		JOIN mdl_trx_patient_visit v
			ON v.id = c.visit_id
			AND v.id_mst_institution = ?
			AND v.delete_time IS NULL
		LEFT JOIN mdl_mst_patient_institution p
			ON p.id = v.id_mst_patient
			AND p.delete_time IS NULL
		WHERE c.staff_id = ?
		  AND v.create_time >= ?
		  AND v.create_time < ?
		  AND c.delete_time IS NULL
		ORDER BY v.create_time ASC, c.visit_id ASC, c.id ASC
	`

	var total64 int64
	_, err := c.DB.SlaveDB.Context(ctx).SQL(countSQL,
		params.InstitutionID, params.StaffID, params.Start, params.EndExclusive,
	).Get(&total64)
	if err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgCommissionListCount)
	}

	sqlText := listSQL
	args := []interface{}{params.InstitutionID, params.StaffID, params.Start, params.EndExclusive}
	if params.Limit > 0 {
		sqlText += " LIMIT ? OFFSET ?"
		args = append(args, params.Limit, params.Offset)
	}

	rows := []compensationrepo.VisitCommissionListRow{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, args...).Find(&rows); err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgCommissionList)
	}
	return rows, int(total64), nil
}

func (c *CommissionConn) SumValidByWorksheet(ctx context.Context, worksheetID int64) (int64, error) {
	const sqlText = `
		SELECT COALESCE(SUM(commission_amount), 0)
		FROM mdl_trx_visit_commission
		WHERE worksheet_id = ?
		  AND approved_at IS NOT NULL
		  AND delete_time IS NULL
	`
	var total int64
	_, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, worksheetID).Get(&total)
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgCommissionSumValid)
	}
	return total, nil
}

func (c *CommissionConn) CountVisitsByWorksheet(ctx context.Context, worksheetID int64) (int64, error) {
	const sqlText = `
		SELECT COUNT(DISTINCT visit_id)
		FROM mdl_trx_visit_commission
		WHERE worksheet_id = ?
		  AND delete_time IS NULL
	`
	var total int64
	_, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, worksheetID).Get(&total)
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgCommissionSumValid)
	}
	return total, nil
}

func (c *CommissionConn) DistinctVisitIDsByWorksheet(ctx context.Context, worksheetID int64) ([]int64, error) {
	const sqlText = `
		SELECT DISTINCT visit_id
		FROM mdl_trx_visit_commission
		WHERE worksheet_id = ?
		  AND delete_time IS NULL
		ORDER BY visit_id
	`

	type visitIDRow struct {
		VisitID int64 `xorm:"visit_id"`
	}
	rows := []visitIDRow{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, worksheetID).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgCommissionDistinctVisit)
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.VisitID)
	}
	return ids, nil
}

func (c *CommissionConn) GetLiveByID(ctx context.Context, institutionID, id int64) (*model.TrxVisitCommission, bool, error) {
	const sqlText = `
		SELECT
			c.id,
			c.worksheet_id,
			c.visit_id,
			c.staff_id,
			c.revenue_base,
			c.commission_type,
			c.commission_percent,
			c.commission_flat_amount,
			c.commission_amount,
			c.sources,
			c.note,
			c.included_manually,
			c.approved_at,
			c.create_time,
			c.update_time,
			c.delete_time
		FROM mdl_trx_visit_commission c
		INNER JOIN mdl_trx_worksheet w
			ON w.id = c.worksheet_id
			AND w.delete_time IS NULL
		WHERE c.id = ?
		  AND c.delete_time IS NULL
		  AND w.institution_id = ?
	`

	row := &model.TrxVisitCommission{}
	found, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, id, institutionID).Get(row)
	if err != nil {
		return nil, false, errors.Wrap(err, wrapMsgCommissionGetLiveByID)
	}
	if !found {
		return nil, false, nil
	}
	return row, true, nil
}

func (c *CommissionConn) UpdateAssignmentAmounts(ctx context.Context, row *model.TrxVisitCommission) error {
	if row == nil {
		return errors.Wrap(errors.New("commission is required"), wrapMsgCommissionUpdateAmounts)
	}

	const sqlText = `
		UPDATE mdl_trx_visit_commission
		SET
			revenue_base = ?,
			commission_type = ?,
			commission_percent = ?,
			commission_flat_amount = ?,
			commission_amount = ?,
			note = ?,
			approved_at = ?,
			update_time = NOW()
		WHERE id = ?
		  AND delete_time IS NULL
		RETURNING update_time
	`

	res, err := c.commissionWriteSession(ctx).SQL(sqlText,
		row.RevenueBase,
		row.CommissionType,
		nullFloat64(row.CommissionPercent),
		nullInt64(row.CommissionFlatAmount),
		row.CommissionAmount,
		nullString(row.Note),
		nullTime(row.ApprovedAt),
		row.ID,
	).QueryInterface()
	if err != nil {
		return errors.Wrap(err, wrapMsgCommissionUpdateAmounts)
	}
	if len(res) == 0 {
		return errors.Wrap(errors.New("commission was not updated"), wrapMsgCommissionUpdateAmounts)
	}
	row.UpdateTime = xormlib.ToTime(res[0]["update_time"])
	return nil
}

func (c *CommissionConn) SoftDelete(ctx context.Context, id int64) (bool, error) {
	const sqlText = `
		UPDATE mdl_trx_visit_commission
		SET delete_time = NOW()
		WHERE id = ?
		  AND delete_time IS NULL
		RETURNING id
	`
	res, err := c.commissionWriteSession(ctx).SQL(sqlText, id).QueryInterface()
	if err != nil {
		return false, errors.Wrap(err, wrapMsgCommissionSoftDelete)
	}
	return len(res) > 0, nil
}

func (c *CommissionConn) SoftDeleteByWorksheet(ctx context.Context, worksheetID int64) (int64, error) {
	const sqlText = `
		UPDATE mdl_trx_visit_commission
		SET delete_time = NOW()
		WHERE worksheet_id = ?
		  AND delete_time IS NULL
	`
	res, err := c.commissionWriteSession(ctx).Exec(sqlText, worksheetID)
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgCommissionSoftDeleteWS)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (c *CommissionConn) SumRevenueByVisitIDs(ctx context.Context, visitIDs []int64) (map[int64]int64, error) {
	if len(visitIDs) == 0 {
		return map[int64]int64{}, nil
	}

	const sqlText = `
		SELECT
			id_trx_patient_visit AS visit_id,
			ROUND(COALESCE(SUM(COALESCE(adjusted_price, total_price)), 0))::bigint AS revenue_base
		FROM mdl_trx_visit_product
		WHERE delete_time IS NULL
		  AND id_trx_patient_visit = ANY(?)
		GROUP BY id_trx_patient_visit
	`

	type visitRevenueRow struct {
		VisitID int64 `xorm:"visit_id"`
		Revenue int64 `xorm:"revenue_base"`
	}
	rows := []visitRevenueRow{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, pq.Array(visitIDs)).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgCommissionSumRevenue)
	}

	out := make(map[int64]int64, len(rows))
	for _, row := range rows {
		out[row.VisitID] = row.Revenue
	}
	return out, nil
}

func (c *CommissionConn) SoftWarningAggregates(ctx context.Context, worksheetID int64) ([]compensationrepo.VisitCommissionWarning, error) {
	const sqlText = `
		WITH commission_agg AS (
			SELECT
				visit_id,
				COALESCE(SUM(commission_percent) FILTER (WHERE commission_type = 'percent'), 0)::float8 AS percent_sum,
				COALESCE(SUM(commission_amount), 0) AS commission_idr
			FROM mdl_trx_visit_commission
			WHERE worksheet_id = ?
			  AND delete_time IS NULL
			GROUP BY visit_id
		),
		revenue AS (
			SELECT
				vp.id_trx_patient_visit AS visit_id,
				ROUND(COALESCE(SUM(COALESCE(vp.adjusted_price, vp.total_price)), 0))::bigint AS revenue_base
			FROM mdl_trx_visit_product vp
			WHERE vp.delete_time IS NULL
			  AND vp.id_trx_patient_visit IN (SELECT visit_id FROM commission_agg)
			GROUP BY vp.id_trx_patient_visit
		)
		SELECT
			c.visit_id,
			c.percent_sum,
			c.commission_idr,
			COALESCE(r.revenue_base, 0) AS revenue_base
		FROM commission_agg c
		LEFT JOIN revenue r ON r.visit_id = c.visit_id
		WHERE c.percent_sum > 100
		   OR c.commission_idr > COALESCE(r.revenue_base, 0)
		ORDER BY c.visit_id
	`

	rows := []compensationrepo.VisitCommissionWarning{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, worksheetID).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgCommissionWarnings)
	}
	return rows, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func nullFloat64(v sql.NullFloat64) interface{} {
	if !v.Valid {
		return nil
	}
	return v.Float64
}

func nullInt64(v sql.NullInt64) interface{} {
	if !v.Valid {
		return nil
	}
	return v.Int64
}

func nullString(v sql.NullString) interface{} {
	if !v.Valid {
		return nil
	}
	return v.String
}

func nullTime(v sql.NullTime) interface{} {
	if !v.Valid {
		return nil
	}
	return v.Time
}

func jsonbOrNil(raw json.RawMessage) interface{} {
	if len(raw) == 0 {
		return nil
	}
	return []byte(raw)
}

func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	default:
		return 0
	}
}
