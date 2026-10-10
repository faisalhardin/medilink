package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

const (
	wrapErrSnapshotPrefix     = "WageSnapshotDB."
	wrapMsgSnapshotList       = wrapErrSnapshotPrefix + "List"
	wrapMsgSnapshotInsert     = wrapErrSnapshotPrefix + "Insert"
	wrapMsgSnapshotGet        = wrapErrSnapshotPrefix + "Get"
	wrapMsgSnapshotUpdate     = wrapErrSnapshotPrefix + "UpdateWageInputs"
	wrapMsgSnapshotSoftDelete = wrapErrSnapshotPrefix + "SoftDelete"
)

// WageSnapshotConn holds the DB connection for wage snapshots.
type WageSnapshotConn struct {
	DB *xormlib.DBConnect
}

func (c *WageSnapshotConn) writeSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

var _ compensationrepo.WageSnapshotDB = (*WageSnapshotConn)(nil)

// NewWageSnapshotDB returns a WageSnapshotDB implementation bound to the xorm connection.
func NewWageSnapshotDB(db *xormlib.DBConnect) compensationrepo.WageSnapshotDB {
	return &WageSnapshotConn{DB: db}
}

func (c *WageSnapshotConn) List(ctx context.Context, institutionID int64) ([]model.TrxWagePeriodSnapshot, error) {
	const sqlText = `
		SELECT s.*,
		       p.uuid AS compensation_period_uuid
		FROM mdl_trx_wage_period_snapshot s
		LEFT JOIN mdl_trx_compensation_period p
		  ON p.id = s.compensation_period_id
		 AND p.institution_id = s.institution_id
		 AND p.delete_time IS NULL
		WHERE s.institution_id = ?
		  AND s.delete_time IS NULL
		ORDER BY s.id ASC
	`
	rows := []model.TrxWagePeriodSnapshot{}
	err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, institutionID).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgSnapshotList)
	}
	return rows, nil
}

func (c *WageSnapshotConn) Insert(ctx context.Context, row *model.TrxWagePeriodSnapshot) error {
	if row == nil {
		return errors.Wrap(errors.New("wage snapshot is required"), wrapMsgSnapshotInsert)
	}
	_, err := c.writeSession(ctx).
		Table(model.TrxWagePeriodSnapshotTableName).
		InsertOne(row)
	if err != nil {
		return errors.Wrap(err, wrapMsgSnapshotInsert)
	}
	return nil
}

func (c *WageSnapshotConn) Get(ctx context.Context, institutionID, id int64) (*model.TrxWagePeriodSnapshot, bool, error) {
	const sqlText = `
		SELECT s.*,
		       p.uuid AS compensation_period_uuid
		FROM mdl_trx_wage_period_snapshot s
		LEFT JOIN mdl_trx_compensation_period p
		  ON p.id = s.compensation_period_id
		 AND p.institution_id = s.institution_id
		 AND p.delete_time IS NULL
		WHERE s.institution_id = ?
		  AND s.id = ?
		  AND s.delete_time IS NULL
	`
	row := model.TrxWagePeriodSnapshot{}
	has, err := c.writeSession(ctx).SQL(sqlText, institutionID, id).Get(&row)
	if err != nil {
		return nil, false, errors.Wrap(err, wrapMsgSnapshotGet)
	}
	if !has {
		return nil, false, nil
	}
	return &row, true, nil
}

func (c *WageSnapshotConn) UpdateWageInputs(ctx context.Context, row *model.TrxWagePeriodSnapshot) (bool, error) {
	if row == nil {
		return false, errors.Wrap(errors.New("wage snapshot is required"), wrapMsgSnapshotUpdate)
	}
	const sqlText = `
		UPDATE mdl_trx_wage_period_snapshot
		SET mandatory_working_days = ?,
		    staff_working_days = ?,
		    final_wage = ?,
		    total_wage = ?
		WHERE id = ?
		  AND institution_id = ?
		  AND delete_time IS NULL
	`
	res, err := c.writeSession(ctx).Exec(
		sqlText,
		nullInt64(row.MandatoryWorkingDays),
		nullInt64(row.StaffWorkingDays),
		nullInt64(row.FinalWage),
		nullInt64(row.TotalWage),
		row.ID,
		row.InstitutionID,
	)
	if err != nil {
		return false, errors.Wrap(err, wrapMsgSnapshotUpdate)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, errors.Wrap(err, wrapMsgSnapshotUpdate)
	}
	return affected > 0, nil
}

func (c *WageSnapshotConn) SoftDelete(ctx context.Context, institutionID, id int64) (bool, error) {
	affected, err := c.writeSession(ctx).
		Table(model.TrxWagePeriodSnapshotTableName).
		Where("id = ?", id).
		And("institution_id = ?", institutionID).
		Delete(&model.TrxWagePeriodSnapshot{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgSnapshotSoftDelete)
	}
	return affected > 0, nil
}
