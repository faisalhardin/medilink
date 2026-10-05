package compensation

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	utilcommon "github.com/faisalhardin/medilink/internal/library/util/common"
	"github.com/pkg/errors"
)

const (
	wrapWageUCPrefix  = "WageUC."
	wrapMsgListWages  = wrapWageUCPrefix + "ListWages"
	wrapMsgUpsertWage = wrapWageUCPrefix + "UpsertWage"
	wrapMsgDeleteWage = wrapWageUCPrefix + "DeleteWage"

	errInvalidWageCadence = "INVALID_WAGE_CADENCE"
	errWageRangeInvalid   = "WAGE_EFFECTIVE_RANGE_INVALID"
	errWageOverlap        = "WAGE_EFFECTIVE_RANGE_OVERLAP"
	errWageMultipleActive = "WAGE_MULTIPLE_ACTIVE"
	errWageNotFound       = "WAGE_NOT_FOUND"

	msgInvalidWageCadence = "wage_cadence must be monthly, weekly, or daily"
	msgWageStaffRequired  = "staff_id and effective_from are required"
	msgWageRangeOrder     = "effective_to must be on or after effective_from"
	msgWageOverlap        = "wage effective dates overlap another wage for this staff"
	msgWageMultipleActive = "more than one active wage exists for this staff"
	msgWageNotFound       = "wage was not found"
)

// openEndedWage is the overlap stand-in for a null effective_to.
var openEndedWage = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)

var _ compensationuc.WageUC = (*WageUC)(nil)

// WageUC implements wage configuration.
type WageUC struct {
	WageDB      compensationrepo.WageDB
	Transaction xormlib.DBTransactionInterface
}

func NewWageUC(uc *WageUC) *WageUC {
	return uc
}

func (u *WageUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	return userDetail, nil
}

func (u *WageUC) ListWages(ctx context.Context, req model.ListStaffWagesRequest) (model.ListStaffWagesResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListStaffWagesResponse{}, err
	}

	rows, err := u.WageDB.ListActive(ctx, userDetail.InstitutionID, strings.TrimSpace(req.StaffID))
	if err != nil {
		return model.ListStaffWagesResponse{}, errors.Wrap(err, wrapMsgListWages)
	}

	wages := make([]model.StaffWageResponse, 0, len(rows))
	for _, row := range rows {
		wages = append(wages, row.ToResponse())
	}
	return model.ListStaffWagesResponse{Wages: wages}, nil
}

func (u *WageUC) UpsertWage(ctx context.Context, req model.UpsertStaffWageRequest) (model.UpsertStaffWageResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.UpsertStaffWageResponse{}, err
	}

	staffID := strings.TrimSpace(req.StaffID)
	if staffID == "" || req.EffectiveFrom.Time().IsZero() {
		return model.UpsertStaffWageResponse{}, commonerr.SetNewBadRequest(errWageRangeInvalid, msgWageStaffRequired)
	}
	if !req.WageCadence.IsValid() {
		return model.UpsertStaffWageResponse{}, commonerr.SetNewBadRequest(errInvalidWageCadence, msgInvalidWageCadence)
	}

	from := utilcommon.DateOnly(req.EffectiveFrom.Time())
	var to sql.NullTime
	if req.EffectiveTo != nil {
		if req.EffectiveTo.Time().IsZero() {
			return model.UpsertStaffWageResponse{}, commonerr.SetNewBadRequest(errWageRangeInvalid, msgWageRangeOrder)
		}
		toDate := utilcommon.DateOnly(req.EffectiveTo.Time())
		if toDate.Before(from) {
			return model.UpsertStaffWageResponse{}, commonerr.SetNewBadRequest(errWageRangeInvalid, msgWageRangeOrder)
		}
		to = sql.NullTime{Time: toDate, Valid: true}
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.UpsertStaffWageResponse{}, errors.Wrap(err, wrapMsgUpsertWage)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	live, err := u.WageDB.ListLiveByStaff(ctx, userDetail.InstitutionID, staffID)
	if err != nil {
		return model.UpsertStaffWageResponse{}, errors.Wrap(err, wrapMsgUpsertWage)
	}

	active := make([]model.MstStaffWage, 0, 1)
	for _, row := range live {
		if row.IsActive {
			active = append(active, row)
		}
	}
	if len(active) > 1 {
		err = commonerr.SetNewBadRequest(errWageMultipleActive, msgWageMultipleActive)
		return model.UpsertStaffWageResponse{}, err
	}

	var excludeID int64
	if len(active) == 1 {
		prev := active[0]
		excludeID = prev.ID
		closeTo := from.AddDate(0, 0, -1)
		prevFrom := utilcommon.DateOnly(prev.EffectiveFrom)
		if closeTo.Before(prevFrom) {
			err = commonerr.SetNewBadRequest(errWageOverlap, msgWageOverlap)
			return model.UpsertStaffWageResponse{}, err
		}
		closedTo := sql.NullTime{Time: closeTo, Valid: true}
		if wageOverlapsAny(prevFrom, closedTo, live, excludeID) {
			err = commonerr.SetNewBadRequest(errWageOverlap, msgWageOverlap)
			return model.UpsertStaffWageResponse{}, err
		}
		if wageOverlapsAny(from, to, live, excludeID) {
			err = commonerr.SetNewBadRequest(errWageOverlap, msgWageOverlap)
			return model.UpsertStaffWageResponse{}, err
		}
		if err = u.WageDB.Close(ctx, prev.ID, userDetail.InstitutionID, closeTo, userDetail.UUID); err != nil {
			return model.UpsertStaffWageResponse{}, errors.Wrap(err, wrapMsgUpsertWage)
		}
	} else if wageOverlapsAny(from, to, live, 0) {
		err = commonerr.SetNewBadRequest(errWageOverlap, msgWageOverlap)
		return model.UpsertStaffWageResponse{}, err
	}

	row := &model.MstStaffWage{
		StaffID:       staffID,
		InstitutionID: userDetail.InstitutionID,
		WageAmount:    req.WageAmount,
		WageCadence:   req.WageCadence,
		IsActive:      true,
		EffectiveFrom: from,
		EffectiveTo:   to,
		CreatedBy:     sql.NullString{String: userDetail.UUID, Valid: userDetail.UUID != ""},
		UpdatedBy:     sql.NullString{String: userDetail.UUID, Valid: userDetail.UUID != ""},
	}
	if err = u.WageDB.Insert(ctx, row); err != nil {
		return model.UpsertStaffWageResponse{}, errors.Wrap(err, wrapMsgUpsertWage)
	}
	return model.UpsertStaffWageResponse{Wage: row.ToResponse()}, nil
}

func (u *WageUC) DeleteWage(ctx context.Context, wageID int64) (model.DeleteStaffWageResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.DeleteStaffWageResponse{}, err
	}

	found, err := u.WageDB.SoftDelete(ctx, userDetail.InstitutionID, wageID)
	if err != nil {
		return model.DeleteStaffWageResponse{}, errors.Wrap(err, wrapMsgDeleteWage)
	}
	if !found {
		return model.DeleteStaffWageResponse{}, commonerr.SetNewError(http.StatusNotFound, errWageNotFound, msgWageNotFound)
	}
	return model.DeleteStaffWageResponse{Success: true}, nil
}

func wageOverlapsAny(from time.Time, to sql.NullTime, rows []model.MstStaffWage, excludeID int64) bool {
	for _, row := range rows {
		if row.ID == excludeID {
			continue
		}
		if utilcommon.PeriodsOverlap(from, wageRangeEnd(to), utilcommon.DateOnly(row.EffectiveFrom), wageRangeEnd(row.EffectiveTo)) {
			return true
		}
	}
	return false
}

func wageRangeEnd(to sql.NullTime) time.Time {
	if !to.Valid || to.Time.IsZero() {
		return openEndedWage
	}
	return utilcommon.DateOnly(to.Time)
}
