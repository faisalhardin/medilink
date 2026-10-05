package compensation

import (
	"context"
	"database/sql"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

const (
	wrapErrWagePrefix     = "WageDB."
	wrapMsgWageList       = wrapErrWagePrefix + "ListActive"
	wrapMsgWageListLive   = wrapErrWagePrefix + "ListLiveByStaff"
	wrapMsgWageClose      = wrapErrWagePrefix + "Close"
	wrapMsgWageInsert     = wrapErrWagePrefix + "Insert"
	wrapMsgWageSoftDelete = wrapErrWagePrefix + "SoftDelete"
)

// WageConn holds the DB connection for staff wage configuration.
type WageConn struct {
	DB *xormlib.DBConnect
}

func (c *WageConn) wageWriteSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

var _ compensationrepo.WageDB = (*WageConn)(nil)

// NewWageDB returns a WageDB implementation bound to the xorm connection.
func NewWageDB(db *xormlib.DBConnect) compensationrepo.WageDB {
	return &WageConn{DB: db}
}

func (c *WageConn) ListActive(ctx context.Context, institutionID int64, staffID string) ([]model.MstStaffWage, error) {
	sess := c.DB.SlaveDB.Context(ctx).
		Table(model.MstStaffWageTableName).
		Where("institution_id = ?", institutionID).
		And("delete_time IS NULL").
		And("is_active = ?", true)
	if staffID != "" {
		sess = sess.And("staff_id = ?", staffID)
	}
	rows := []model.MstStaffWage{}
	if err := sess.OrderBy("id ASC").Find(&rows); err != nil {
		return nil, errors.Wrap(err, wrapMsgWageList)
	}
	return rows, nil
}

func (c *WageConn) ListLiveByStaff(ctx context.Context, institutionID int64, staffID string) ([]model.MstStaffWage, error) {
	rows := []model.MstStaffWage{}
	err := c.wageWriteSession(ctx).
		Table(model.MstStaffWageTableName).
		Where("institution_id = ?", institutionID).
		And("staff_id = ?", staffID).
		And("delete_time IS NULL").
		OrderBy("id ASC").
		Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgWageListLive)
	}
	return rows, nil
}

func (c *WageConn) Close(ctx context.Context, id, institutionID int64, effectiveTo time.Time, updatedBy string) error {
	wage := model.MstStaffWage{
		IsActive:    false,
		EffectiveTo: sql.NullTime{Time: effectiveTo, Valid: true},
		UpdatedBy:   sql.NullString{String: updatedBy, Valid: updatedBy != ""},
	}
	affected, err := c.wageWriteSession(ctx).
		Table(model.MstStaffWageTableName).
		Where("id = ?", id).
		And("institution_id = ?", institutionID).
		And("delete_time IS NULL").
		Cols("is_active", "effective_to", "updated_by").
		Update(&wage)
	if err != nil {
		return errors.Wrap(err, wrapMsgWageClose)
	}
	if affected == 0 {
		return errors.Wrap(errors.New("wage to close was not found"), wrapMsgWageClose)
	}
	return nil
}

func (c *WageConn) Insert(ctx context.Context, w *model.MstStaffWage) error {
	if w == nil {
		return errors.Wrap(errors.New("wage is required"), wrapMsgWageInsert)
	}
	_, err := c.wageWriteSession(ctx).
		Table(model.MstStaffWageTableName).
		InsertOne(w)
	if err != nil {
		return errors.Wrap(err, wrapMsgWageInsert)
	}
	return nil
}

func (c *WageConn) SoftDelete(ctx context.Context, institutionID, id int64) (bool, error) {
	affected, err := c.wageWriteSession(ctx).
		Table(model.MstStaffWageTableName).
		Where("id = ?", id).
		And("institution_id = ?", institutionID).
		Delete(&model.MstStaffWage{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgWageSoftDelete)
	}
	return affected > 0, nil
}
