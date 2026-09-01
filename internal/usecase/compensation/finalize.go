package compensation

import (
	"context"
	"database/sql"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

func (u *PeriodUC) FinalizePeriod(ctx context.Context, periodUUID string) (model.FinalizeCompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.FinalizeCompensationPeriodResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, periodUUID, wrapMsgFinalizePeriod)
	if err != nil {
		return model.FinalizeCompensationPeriodResponse{}, err
	}

	if period.Status == model.CompensationPeriodStatusFinalized {
		visitIDs, err := u.Commissions.DistinctVisitIDsByPeriod(ctx, period.ID)
		if err != nil {
			return model.FinalizeCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgFinalizePeriod)
		}
		return model.FinalizeCompensationPeriodResponse{
			Period:           period.ToResponse(0),
			LockedVisitCount: int64(len(visitIDs)),
		}, nil
	}

	if !period.Status.CanTransition(model.CompensationPeriodStatusFinalized) {
		return model.FinalizeCompensationPeriodResponse{}, commonerr.SetNewBadRequest(errIllegalTransition, msgFinalizeFromDraftOnly)
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.FinalizeCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgFinalizePeriod)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	totals, err := u.Commissions.SumByPeriod(ctx, period.ID)
	if err != nil {
		return model.FinalizeCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgFinalizePeriod)
	}
	visitIDs, err := u.Commissions.DistinctVisitIDsByPeriod(ctx, period.ID)
	if err != nil {
		return model.FinalizeCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgFinalizePeriod)
	}

	now := u.currentTime()
	applyPhase1FinalizeTotals(period, totals, now, userDetail.UUID)
	if err = u.PeriodDB.UpdateStatusAndTotals(ctx, period); err != nil {
		return model.FinalizeCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgFinalizePeriod)
	}

	lockedCount, err := u.VisitLockDB.LockVisits(ctx, period.ID, visitIDs, now)
	if err != nil {
		return model.FinalizeCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgFinalizePeriod)
	}

	return model.FinalizeCompensationPeriodResponse{
		Period:           period.ToResponse(0),
		LockedVisitCount: lockedCount,
	}, nil
}

func applyPhase1FinalizeTotals(period *model.TrxCompensationPeriod, totals compensationrepo.PeriodCommissionTotals, now time.Time, staffUUID string) {
	period.Status = model.CompensationPeriodStatusFinalized
	period.WageSnapshot = nil
	period.TotalWage = sql.NullInt64{Int64: 0, Valid: true}
	period.TotalCommission = sql.NullInt64{Int64: totals.TotalCommission, Valid: true}
	period.TotalPayout = sql.NullInt64{Int64: totals.TotalCommission, Valid: true}
	period.StaffCount = sql.NullInt64{Int64: totals.StaffCount, Valid: true}
	period.VisitCount = sql.NullInt64{Int64: totals.VisitCount, Valid: true}
	period.FinalizedAt = sql.NullTime{Time: now, Valid: true}
	if staffUUID != "" {
		period.FinalizedBy = sql.NullString{String: staffUUID, Valid: true}
	}
}
