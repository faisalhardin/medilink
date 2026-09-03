package compensation

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	wrapMsgCommissionUpsert        = "CommissionDB.Upsert"
	wrapMsgCommissionList          = "CommissionDB.ListByPeriodStaff"
	wrapMsgCommissionListCount     = "CommissionDB.ListByPeriodStaffCount"
	wrapMsgCommissionSumByPeriod   = "CommissionDB.SumByPeriod"
	wrapMsgCommissionDistinctVisit = "CommissionDB.DistinctVisitIDsByPeriod"
	wrapMsgCommissionSumByStaff    = "CommissionDB.SumByStaff"
	wrapMsgCommissionSumRevenue    = "CommissionDB.SumRevenueByVisitIDs"
	wrapMsgCommissionWarnings      = "CommissionDB.SoftWarningAggregates"
)

var _ compensationrepo.CommissionDB = (*Conn)(nil)

// NewCommissionDB returns a CommissionDB implementation bound to the xorm connection.
func NewCommissionDB(db *xormlib.DBConnect) compensationrepo.CommissionDB {
	return &Conn{DB: db}
}

func (c *Conn) Upsert(ctx context.Context, row *model.TrxVisitCommission) error {
	if row == nil {
		return errors.Wrap(errors.New("commission is required"), wrapMsgCommissionUpsert)
	}

	const sqlText = `
		INSERT INTO mdl_trx_visit_commission (
			period_id, visit_id, staff_id,
			revenue_base, commission_type, commission_percent, commission_flat_amount,
			commission_amount, sources, note, included_manually,
			create_time, update_time
		) VALUES (
			?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			NOW(), NOW()
		)
		ON CONFLICT (period_id, visit_id, staff_id) WHERE delete_time IS NULL
		DO UPDATE SET
			revenue_base = EXCLUDED.revenue_base,
			commission_type = EXCLUDED.commission_type,
			commission_percent = EXCLUDED.commission_percent,
			commission_flat_amount = EXCLUDED.commission_flat_amount,
			commission_amount = EXCLUDED.commission_amount,
			sources = EXCLUDED.sources,
			note = EXCLUDED.note,
			included_manually = EXCLUDED.included_manually,
			update_time = NOW()
		RETURNING id, create_time, update_time
	`

	res, err := c.writeSession(ctx).SQL(sqlText,
		row.PeriodID,
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
	).QueryInterface()
	if err != nil {
		return errors.Wrap(err, wrapMsgCommissionUpsert)
	}
	if len(res) == 0 {
		return errors.Wrap(errors.New("upsert returned no row"), wrapMsgCommissionUpsert)
	}

	r := res[0]
	row.ID = toInt64(r["id"])
	row.CreateTime = xormlib.ToTime(r["create_time"])
	row.UpdateTime = xormlib.ToTime(r["update_time"])
	return nil
}

func (c *Conn) ListByPeriodStaff(ctx context.Context, params compensationrepo.ListVisitCommissionParams) ([]model.TrxVisitCommission, int, error) {
	countSess := c.commissionListSession(ctx, params)
	total64, err := countSess.Count(&model.TrxVisitCommission{})
	if err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgCommissionListCount)
	}

	sess := c.commissionListSession(ctx, params).OrderBy("visit_id ASC, id ASC")
	if params.Limit > 0 {
		sess = sess.Limit(params.Limit, params.Offset)
	}

	rows := []model.TrxVisitCommission{}
	if err := sess.Find(&rows); err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgCommissionList)
	}
	return rows, int(total64), nil
}

func (c *Conn) commissionListSession(ctx context.Context, params compensationrepo.ListVisitCommissionParams) *xorm.Session {
	return c.DB.SlaveDB.Context(ctx).
		Table(model.TrxVisitCommissionTableName).
		Where("period_id = ?", params.PeriodID).
		And("staff_id = ?", params.StaffID)
}

func (c *Conn) SumByPeriod(ctx context.Context, periodID int64) (compensationrepo.PeriodCommissionTotals, error) {
	const sqlText = `
		SELECT
			COALESCE(SUM(commission_amount), 0) AS total_commission,
			COUNT(DISTINCT staff_id) AS staff_count,
			COUNT(DISTINCT visit_id) AS visit_count
		FROM mdl_trx_visit_commission
		WHERE period_id = ?
		  AND delete_time IS NULL
	`

	totals := compensationrepo.PeriodCommissionTotals{}
	_, err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, periodID).Get(&totals)
	if err != nil {
		return compensationrepo.PeriodCommissionTotals{}, errors.Wrap(err, wrapMsgCommissionSumByPeriod)
	}
	return totals, nil
}

func (c *Conn) DistinctVisitIDsByPeriod(ctx context.Context, periodID int64) ([]int64, error) {
	const sqlText = `
		SELECT DISTINCT visit_id
		FROM mdl_trx_visit_commission
		WHERE period_id = ?
		  AND delete_time IS NULL
		ORDER BY visit_id
	`

	type visitIDRow struct {
		VisitID int64 `xorm:"visit_id"`
	}
	rows := []visitIDRow{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, periodID).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgCommissionDistinctVisit)
	}

	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.VisitID)
	}
	return ids, nil
}

func (c *Conn) SumByStaff(ctx context.Context, periodID int64) ([]compensationrepo.StaffCommissionTotals, error) {
	const sqlText = `
		SELECT
			staff_id,
			COALESCE(SUM(commission_amount), 0) AS total_commission,
			COUNT(DISTINCT visit_id) AS visit_count
		FROM mdl_trx_visit_commission
		WHERE period_id = ?
		  AND delete_time IS NULL
		GROUP BY staff_id
		ORDER BY staff_id
	`

	rows := []compensationrepo.StaffCommissionTotals{}
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, periodID).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgCommissionSumByStaff)
	}
	return rows, nil
}

func (c *Conn) SumRevenueByVisitIDs(ctx context.Context, visitIDs []int64) (map[int64]int64, error) {
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

func (c *Conn) SoftWarningAggregates(ctx context.Context, periodID int64) ([]compensationrepo.VisitCommissionWarning, error) {
	const sqlText = `
		WITH commission_agg AS (
			SELECT
				visit_id,
				COALESCE(SUM(commission_percent) FILTER (WHERE commission_type = 'percent'), 0)::float8 AS percent_sum,
				COALESCE(SUM(commission_amount), 0) AS commission_idr
			FROM mdl_trx_visit_commission
			WHERE period_id = ?
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
	if err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, periodID).Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgCommissionWarnings)
	}
	return rows, nil
}

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
