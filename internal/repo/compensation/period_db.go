package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	wrapErrMsgPrefix    = "PeriodDB."
	wrapMsgCreate       = wrapErrMsgPrefix + "Create"
	wrapMsgGetByUUID    = wrapErrMsgPrefix + "GetByUUID"
	wrapMsgList         = wrapErrMsgPrefix + "List"
	wrapMsgListCount    = wrapErrMsgPrefix + "ListCount"
	wrapMsgUpdate       = wrapErrMsgPrefix + "UpdateStatusAndTotals"
	wrapMsgSoftDelete   = wrapErrMsgPrefix + "SoftDelete"
	wrapMsgGenerateUUID = wrapErrMsgPrefix + "GenerateUUID"
)

var _ compensationrepo.PeriodDB = (*Conn)(nil)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewPeriodDB returns a PeriodDB implementation bound to the xorm connection.
func NewPeriodDB(db *xormlib.DBConnect) compensationrepo.PeriodDB {
	return &Conn{DB: db}
}

// writeSession returns the active TX session if one was put on ctx by the
// usecase; otherwise returns a fresh master-engine session.
func (c *Conn) writeSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

func (c *Conn) Create(ctx context.Context, period *model.TrxCompensationPeriod) error {
	if period == nil {
		return errors.Wrap(errors.New("period is required"), wrapMsgCreate)
	}

	if period.UUID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return errors.Wrap(err, wrapMsgGenerateUUID)
		}
		period.UUID = id.String()
	}
	if period.Status == "" {
		period.Status = model.CompensationPeriodStatusOpen
	}

	_, err := c.writeSession(ctx).
		Table(model.TrxCompensationPeriodTableName).
		InsertOne(period)
	if err != nil {
		return errors.Wrap(err, wrapMsgCreate)
	}
	return nil
}

func (c *Conn) GetByUUID(ctx context.Context, institutionID int64, uuid string) (*model.TrxCompensationPeriod, bool, error) {
	row := &model.TrxCompensationPeriod{}
	ok, err := c.DB.SlaveDB.Context(ctx).
		Table(model.TrxCompensationPeriodTableName).
		Where("uuid = ?", uuid).
		And("institution_id = ?", institutionID).
		Get(row)
	if err != nil {
		return nil, false, errors.Wrap(err, wrapMsgGetByUUID)
	}
	if !ok {
		return nil, false, nil
	}
	return row, true, nil
}

func (c *Conn) List(ctx context.Context, params model.ListCompensationPeriodParams) ([]model.TrxCompensationPeriod, int, error) {
	countSess := c.listSession(ctx, params)
	total64, err := countSess.Count(&model.TrxCompensationPeriod{})
	if err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgListCount)
	}

	sess := c.listSession(ctx, params).OrderBy("period_start DESC, period_end DESC")
	if params.Limit > 0 {
		sess = sess.Limit(params.Limit, params.Offset)
	}

	rows := []model.TrxCompensationPeriod{}
	if err := sess.Find(&rows); err != nil {
		return nil, 0, errors.Wrap(err, wrapMsgList)
	}
	return rows, int(total64), nil
}

func (c *Conn) listSession(ctx context.Context, params model.ListCompensationPeriodParams) *xorm.Session {
	sess := c.DB.SlaveDB.Context(ctx).
		Table(model.TrxCompensationPeriodTableName).
		Where("institution_id = ?", params.InstitutionID)
	if params.Status != "" {
		sess = sess.And("status = ?", params.Status)
	}
	return sess
}

func (c *Conn) UpdateStatusAndTotals(ctx context.Context, period *model.TrxCompensationPeriod) error {
	if period == nil {
		return errors.Wrap(errors.New("period is required"), wrapMsgUpdate)
	}

	_, err := c.writeSession(ctx).
		Table(model.TrxCompensationPeriodTableName).
		Where("uuid = ?", period.UUID).
		And("institution_id = ?", period.InstitutionID).
		Cols(
			"status",
			"wage_snapshot",
			"total_wage",
			"total_commission",
			"total_payout",
			"staff_count",
			"visit_count",
			"drafted_at",
			"drafted_by",
			"finalized_at",
			"finalized_by",
		).
		Update(period)
	if err != nil {
		return errors.Wrap(err, wrapMsgUpdate)
	}
	return nil
}

func (c *Conn) SoftDelete(ctx context.Context, institutionID int64, uuid string) (bool, error) {
	affected, err := c.writeSession(ctx).
		Table(model.TrxCompensationPeriodTableName).
		Where("uuid = ?", uuid).
		And("institution_id = ?", institutionID).
		Delete(&model.TrxCompensationPeriod{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgSoftDelete)
	}
	return affected > 0, nil
}
