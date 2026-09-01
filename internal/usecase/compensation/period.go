package compensation

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	wrapPeriodUCPrefix     = "PeriodUC."
	wrapMsgCreatePeriod    = wrapPeriodUCPrefix + "CreatePeriod"
	wrapMsgListPeriods     = wrapPeriodUCPrefix + "ListPeriods"
	wrapMsgGetPeriod       = wrapPeriodUCPrefix + "GetPeriod"
	wrapMsgDraftPeriod     = wrapPeriodUCPrefix + "DraftPeriod"
	wrapMsgFinalizePeriod  = wrapPeriodUCPrefix + "FinalizePeriod"
	wrapMsgReopenPeriod    = wrapPeriodUCPrefix + "ReopenPeriod"
	wrapMsgDeletePeriod    = wrapPeriodUCPrefix + "DeletePeriod"
	defaultListLimit       = 50
	maxPeriodLabelLen      = 100
	errIllegalTransition   = "ILLEGAL_PERIOD_TRANSITION"
	errInvalidPeriodStatus = "INVALID_COMPENSATION_PERIOD_STATUS"
	errDateRangeOverlap    = "PERIOD_DATE_RANGE_OVERLAP"
	errPeriodNotFound      = "period_not_found"
	errInvalidPeriodDates  = "invalid_period_dates"
	errInvalidLabel        = "invalid_label"

	msgInvalidLabel          = "label is required and must be at most 100 characters"
	msgInvalidPeriodDates    = "period_start and period_end are required and period_end must not be before period_start"
	msgInvalidPeriodStatus   = "status must be open, draft, or finalized"
	msgDraftFromOpenOrDraft  = "period can only be drafted from open or draft status"
	msgReopenFromFinalized   = "period can only be reopened from finalized status"
	msgDeleteOpenOnly        = "only an open period can be deleted"
	msgPeriodNotFound        = "period was not found"
	msgDateRangeOverlap      = "period date range overlaps an existing period"
	msgFinalizeFromDraftOnly = "period can only be finalized from draft status"
)

var _ compensationuc.PeriodUC = (*PeriodUC)(nil)

type PeriodUC struct {
	PeriodDB    compensationrepo.PeriodDB
	Commissions compensationrepo.CommissionAggregator
	VisitLockDB compensationrepo.VisitLockDB
	Transaction xormlib.DBTransactionInterface
	now         func() time.Time
}

func NewPeriodUC(uc *PeriodUC) *PeriodUC {
	return uc
}

func (u *PeriodUC) currentTime() time.Time {
	if u.now != nil {
		return u.now()
	}
	return time.Now().UTC()
}

func (u *PeriodUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	return userDetail, nil
}

func (u *PeriodUC) CreatePeriod(ctx context.Context, req model.CreateCompensationPeriodRequest) (model.CompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	label := strings.TrimSpace(req.Label)
	if label == "" || len(label) > maxPeriodLabelLen {
		return model.CompensationPeriodResponse{}, commonerr.SetNewBadRequest(errInvalidLabel, msgInvalidLabel)
	}

	start := dateOnly(req.PeriodStart.Time())
	end := dateOnly(req.PeriodEnd.Time())
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return model.CompensationPeriodResponse{}, commonerr.SetNewBadRequest(errInvalidPeriodDates, msgInvalidPeriodDates)
	}

	if err := u.rejectIfOverlapping(ctx, userDetail.InstitutionID, start, end, ""); err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	period := &model.TrxCompensationPeriod{
		InstitutionID: userDetail.InstitutionID,
		Label:         label,
		PeriodStart:   start,
		PeriodEnd:     end,
	}
	if err := u.PeriodDB.Create(ctx, period); err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgCreatePeriod)
	}

	return period.ToResponse(0), nil
}

func (u *PeriodUC) ListPeriods(ctx context.Context, req model.ListCompensationPeriodsRequest) (model.ListCompensationPeriodsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListCompensationPeriodsResponse{}, err
	}

	if req.Status != "" && !req.Status.IsValid() {
		return model.ListCompensationPeriodsResponse{}, commonerr.SetNewBadRequest(errInvalidPeriodStatus, msgInvalidPeriodStatus)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	rows, total, err := u.PeriodDB.List(ctx, model.ListCompensationPeriodParams{
		InstitutionID: userDetail.InstitutionID,
		Status:        req.Status,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return model.ListCompensationPeriodsResponse{}, errors.Wrap(err, wrapMsgListPeriods)
	}

	periods := make([]model.CompensationPeriodResponse, 0, len(rows))
	for _, row := range rows {
		periods = append(periods, row.ToResponse(0))
	}
	return model.ListCompensationPeriodsResponse{
		Periods: periods,
		Total:   total,
	}, nil
}

func (u *PeriodUC) GetPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, periodUUID, wrapMsgGetPeriod)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}
	return period.ToResponse(0), nil
}

func (u *PeriodUC) DraftPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, periodUUID, wrapMsgDraftPeriod)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}
	if period.Status != model.CompensationPeriodStatusOpen && period.Status != model.CompensationPeriodStatusDraft {
		return model.CompensationPeriodResponse{}, commonerr.SetNewBadRequest(errIllegalTransition, msgDraftFromOpenOrDraft)
	}

	if err := u.rejectIfOverlapping(ctx, userDetail.InstitutionID, period.PeriodStart, period.PeriodEnd, period.UUID); err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	totals, err := u.Commissions.SumByPeriod(ctx, period.ID)
	if err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgDraftPeriod)
	}

	applyPhase1DraftTotals(period, totals, u.currentTime(), userDetail.UUID)
	if err := u.PeriodDB.UpdateStatusAndTotals(ctx, period); err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgDraftPeriod)
	}
	return period.ToResponse(0), nil
}

func (u *PeriodUC) ReopenPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, periodUUID, wrapMsgReopenPeriod)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}
	if period.Status != model.CompensationPeriodStatusFinalized {
		return model.CompensationPeriodResponse{}, commonerr.SetNewBadRequest(errIllegalTransition, msgReopenFromFinalized)
	}

	period.Status = model.CompensationPeriodStatusDraft
	if err := u.PeriodDB.UpdateStatusAndTotals(ctx, period); err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgReopenPeriod)
	}
	return period.ToResponse(0), nil
}

func (u *PeriodUC) DeletePeriod(ctx context.Context, periodUUID string) (model.DeleteCompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.DeleteCompensationPeriodResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, periodUUID, wrapMsgDeletePeriod)
	if err != nil {
		return model.DeleteCompensationPeriodResponse{}, err
	}
	if period.Status != model.CompensationPeriodStatusOpen {
		return model.DeleteCompensationPeriodResponse{}, commonerr.SetNewBadRequest(errIllegalTransition, msgDeleteOpenOnly)
	}

	found, err := u.PeriodDB.SoftDelete(ctx, userDetail.InstitutionID, periodUUID)
	if err != nil {
		return model.DeleteCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgDeletePeriod)
	}
	if !found {
		return model.DeleteCompensationPeriodResponse{}, commonerr.SetNewBadRequest(errPeriodNotFound, msgPeriodNotFound)
	}
	return model.DeleteCompensationPeriodResponse{Success: true}, nil
}

func (u *PeriodUC) loadPeriod(ctx context.Context, institutionID int64, periodUUID, wrap string) (*model.TrxCompensationPeriod, error) {
	if periodUUID == "" {
		return nil, commonerr.SetNewBadRequest(errPeriodNotFound, msgPeriodNotFound)
	}
	period, found, err := u.PeriodDB.GetByUUID(ctx, institutionID, periodUUID)
	if err != nil {
		return nil, errors.Wrap(err, wrap)
	}
	if !found || period == nil {
		return nil, commonerr.SetNewBadRequest(errPeriodNotFound, msgPeriodNotFound)
	}
	return period, nil
}

func (u *PeriodUC) rejectIfOverlapping(ctx context.Context, institutionID int64, start, end time.Time, excludeUUID string) error {
	rows, _, err := u.PeriodDB.List(ctx, model.ListCompensationPeriodParams{
		InstitutionID: institutionID,
	})
	if err != nil {
		return errors.Wrap(err, wrapPeriodUCPrefix+"rejectIfOverlapping")
	}
	for _, row := range rows {
		if excludeUUID != "" && row.UUID == excludeUUID {
			continue
		}
		if periodsOverlap(start, end, row.PeriodStart, row.PeriodEnd) {
			return commonerr.SetNewBadRequest(errDateRangeOverlap, msgDateRangeOverlap)
		}
	}
	return nil
}

func applyPhase1DraftTotals(period *model.TrxCompensationPeriod, totals compensationrepo.PeriodCommissionTotals, now time.Time, staffUUID string) {
	period.Status = model.CompensationPeriodStatusDraft
	period.WageSnapshot = nil
	period.TotalWage = sql.NullInt64{Int64: 0, Valid: true}
	period.TotalCommission = sql.NullInt64{Int64: totals.TotalCommission, Valid: true}
	period.TotalPayout = sql.NullInt64{Int64: totals.TotalCommission, Valid: true}
	period.StaffCount = sql.NullInt64{Int64: totals.StaffCount, Valid: true}
	period.VisitCount = sql.NullInt64{Int64: totals.VisitCount, Valid: true}
	period.DraftedAt = sql.NullTime{Time: now, Valid: true}
	if staffUUID != "" {
		period.DraftedBy = sql.NullString{String: staffUUID, Valid: true}
	}
}

func dateOnly(t time.Time) time.Time {
	if t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func periodsOverlap(startA, endA, startB, endB time.Time) bool {
	aStart, aEnd := dateOnly(startA), dateOnly(endA)
	bStart, bEnd := dateOnly(startB), dateOnly(endB)
	return !aStart.After(bEnd) && !aEnd.Before(bStart)
}
