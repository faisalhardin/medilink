package patient

import (
	"context"
	"database/sql"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const wrapMsgLockVisits = WrapErrMsgPrefix + "LockVisits"

func (c *Conn) LockVisits(ctx context.Context, periodID int64, visitIDs []int64, lockedAt time.Time) (int64, error) {
	if len(visitIDs) == 0 {
		return 0, nil
	}

	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	affected, err := session.
		Table(model.TrxPatientVisitTableName).
		In("id", visitIDs).
		Cols("compensation_period_id", "compensation_locked_at").
		Update(&model.TrxPatientVisit{
			CompensationPeriodID: sql.NullInt64{Int64: periodID, Valid: true},
			CompensationLockedAt: sql.NullTime{Time: lockedAt, Valid: true},
		})
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgLockVisits)
	}
	return affected, nil
}
