package patient

import (
	"context"
	"database/sql"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	wrapMsgLockVisits           = WrapErrMsgPrefix + "LockVisits"
	wrapMsgCountOpenVisitsByRoom = WrapErrMsgPrefix + "CountOpenVisitsByRoom"
)

func (c *Conn) CountOpenVisitsByRoom(ctx context.Context, clinicID, roomID int64) (int, error) {
	if clinicID <= 0 || roomID <= 0 {
		return 0, nil
	}
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}
	n, err := session.
		Table(model.TrxPatientVisitTableName).
		Where("id_mst_service_point = ? AND id_mst_institution = ?", roomID, clinicID).
		And("delete_time IS NULL").
		Count(&model.TrxPatientVisit{})
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgCountOpenVisitsByRoom)
	}
	return int(n), nil
}

func (c *Conn) LockVisits(ctx context.Context, worksheetID int64, visitIDs []int64, lockedAt time.Time) (int64, error) {
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
		Cols("worksheet_id", "compensation_locked_at").
		Update(&model.TrxPatientVisit{
			WorksheetID:         sql.NullInt64{Int64: worksheetID, Valid: true},
			CompensationLockedAt: sql.NullTime{Time: lockedAt, Valid: true},
		})
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgLockVisits)
	}
	return affected, nil
}
