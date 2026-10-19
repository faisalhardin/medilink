package compensation

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	staffrepo "github.com/faisalhardin/medilink/internal/entity/repo/staff"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	utilcommon "github.com/faisalhardin/medilink/internal/library/util/common"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

const (
	wrapWorksheetUCPrefix    = "WorksheetUC."
	wrapMsgCreateWorksheet   = wrapWorksheetUCPrefix + "CreateWorksheet"
	wrapMsgListWorksheets    = wrapWorksheetUCPrefix + "ListWorksheets"
	wrapMsgGetWorksheet      = wrapWorksheetUCPrefix + "GetWorksheet"
	wrapMsgPatchWorksheet    = wrapWorksheetUCPrefix + "PatchWorksheet"
	wrapMsgDeleteWorksheet   = wrapWorksheetUCPrefix + "DeleteWorksheet"
	wrapMsgFinalizeWorksheet = wrapWorksheetUCPrefix + "FinalizeWorksheet"

	errWorksheetNotFound        = "WORKSHEET_NOT_FOUND"
	errWorksheetOverlap         = "WORKSHEET_DATE_RANGE_OVERLAP"
	errWorksheetAlreadyOnPayday = "WORKSHEET_ALREADY_ON_PAYDAY"
	errWorksheetFinalizeOnly    = "WORKSHEET_MUST_BE_OPEN_TO_FINALIZE"
	errWorksheetDeleteOnly      = "WORKSHEET_MUST_BE_OPEN_TO_DELETE"
	errWorksheetNotEditable     = "WORKSHEET_NOT_EDITABLE"
	errWorksheetGenerateBusy    = "WORKSHEET_GENERATE_PENDING"
	errWorksheetPeriodRequired  = "COMPENSATION_PERIOD_REQUIRED"
	errPaydayPeriodNotFound     = "PAYDAY_PERIOD_NOT_FOUND"
	errPaydayPeriodFinalized    = "PAYDAY_PERIOD_FINALIZED"

	msgWorksheetNotFound        = "worksheet was not found"
	msgWorksheetOverlap         = "worksheet date range overlaps an existing worksheet for this staff"
	msgWorksheetAlreadyOnPayday = "this staff already has a worksheet on this payday"
	msgWorksheetFinalizeOnly    = "worksheet must be open to finalize"
	msgWorksheetDeleteOnly      = "worksheet must be open to delete"
	msgWorksheetNotEditable     = "worksheet must be open to edit"
	msgWorksheetGenerateBusy    = "worksheet generate is in progress"
	msgWorksheetPeriodRequired  = "compensation_period_uuid is required"
	msgPaydayPeriodNotFound     = "compensation period was not found"
	msgPaydayPeriodFinalized    = "payday period is finalized"
)

var _ compensationuc.WorksheetUC = (*WorksheetUC)(nil)

// WorksheetUC implements WorksheetUC.
type WorksheetUC struct {
	WorksheetDB          compensationrepo.WorksheetDB
	CommissionDB         compensationrepo.CommissionDB
	ContributorDB        compensationrepo.ContributorDB
	CompensationPeriodDB compensationrepo.CompensationPeriodDB
	SnapshotDB           compensationrepo.WageSnapshotDB
	StaffDB              staffrepo.StaffDB
	VisitLockDB          compensationrepo.VisitLockDB
	Transaction          xormlib.DBTransactionInterface
	now                  func() time.Time
}

func NewWorksheetUC(uc *WorksheetUC) *WorksheetUC {
	return uc
}

func (u *WorksheetUC) currentTime() time.Time {
	if u.now != nil {
		return u.now()
	}
	return time.Now().UTC()
}

func (u *WorksheetUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	return userDetail, nil
}

func (u *WorksheetUC) CreateWorksheet(ctx context.Context, req model.CreateWorksheetRequest) (model.WorksheetResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.WorksheetResponse{}, err
	}

	label := strings.TrimSpace(req.Label)
	if label == "" || len(label) > maxPeriodLabelLen {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errInvalidLabel, msgInvalidLabel)
	}

	if req.StaffID == "" {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest("STAFF_REQUIRED", "staff_id is required")
	}

	periodUUID := strings.TrimSpace(req.CompensationPeriodUUID)
	if periodUUID == "" {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetPeriodRequired, msgWorksheetPeriodRequired)
	}
	period, found, err := u.CompensationPeriodDB.GetByUUID(ctx, userDetail.InstitutionID, periodUUID)
	if err != nil {
		return model.WorksheetResponse{}, errors.Wrap(err, wrapMsgCreateWorksheet)
	}
	if !found || period == nil {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errPaydayPeriodNotFound, msgPaydayPeriodNotFound)
	}
	if err = rejectIfPaydayFinalized(period); err != nil {
		return model.WorksheetResponse{}, err
	}

	start := utilcommon.DateOnly(period.PeriodStart)
	end := utilcommon.DateOnly(period.PeriodEnd)
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errInvalidPeriodDates, msgInvalidPeriodDates)
	}

	taken, err := u.WorksheetDB.ExistsForStaffPeriod(ctx, userDetail.InstitutionID, req.StaffID, period.ID)
	if err != nil {
		return model.WorksheetResponse{}, errors.Wrap(err, wrapMsgCreateWorksheet)
	}
	if taken {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetAlreadyOnPayday, msgWorksheetAlreadyOnPayday)
	}

	overlaps, err := u.WorksheetDB.ExistsOverlapping(ctx, userDetail.InstitutionID, req.StaffID, start, end, 0)
	if err != nil {
		return model.WorksheetResponse{}, errors.Wrap(err, wrapMsgCreateWorksheet)
	}
	if overlaps {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetOverlap, msgWorksheetOverlap)
	}

	w := &model.TrxWorksheet{
		InstitutionID:          userDetail.InstitutionID,
		StaffID:                req.StaffID,
		Label:                  label,
		PeriodStart:            start,
		PeriodEnd:              end,
		Status:                 model.WorksheetStatusOpen,
		CompensationPeriodID:   sql.NullInt64{Int64: period.ID, Valid: true},
		CompensationPeriodUUID: sql.NullString{String: period.UUID, Valid: true},
		CreatedBy:              sql.NullString{String: userDetail.UUID, Valid: true},
	}
	if err := u.WorksheetDB.Create(ctx, w); err != nil {
		return model.WorksheetResponse{}, errors.Wrap(err, wrapMsgCreateWorksheet)
	}
	return w.ToResponse(), nil
}

func (u *WorksheetUC) ListWorksheets(ctx context.Context, req model.ListWorksheetsRequest) (model.ListWorksheetsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListWorksheetsResponse{}, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	req.Limit = limit + 1 // fetch one extra to detect next page
	req.InstitutionID = userDetail.InstitutionID
	req.CompensationPeriodUUID = strings.TrimSpace(req.CompensationPeriodUUID)
	if err := normalizeWorksheetListDates(&req); err != nil {
		return model.ListWorksheetsResponse{}, err
	}

	rows, err := u.WorksheetDB.List(ctx, req)
	if err != nil {
		return model.ListWorksheetsResponse{}, errors.Wrap(err, wrapMsgListWorksheets)
	}

	var nextCursor null.String
	if len(rows) > limit {
		last := rows[limit-1]
		nextCursor = null.StringFrom(strconv.FormatInt(last.ID, 10))
		rows = rows[:limit]
	}

	out := make([]model.WorksheetResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToResponse())
	}
	return model.ListWorksheetsResponse{
		Worksheets: out,
		NextCursor: nextCursor,
	}, nil
}

func normalizeWorksheetListDates(req *model.ListWorksheetsRequest) error {
	startZero := req.PeriodStart.Time().IsZero()
	endZero := req.PeriodEnd.Time().IsZero()
	if startZero && endZero {
		return nil
	}
	if startZero || endZero {
		return commonerr.SetNewBadRequest(errInvalidPeriodDates, msgInvalidPeriodDates)
	}
	start := utilcommon.DateOnly(req.PeriodStart.Time())
	end := utilcommon.DateOnly(req.PeriodEnd.Time())
	if end.Before(start) {
		return commonerr.SetNewBadRequest(errInvalidPeriodDates, msgInvalidPeriodDates)
	}
	req.PeriodStart = model.Time(start)
	req.PeriodEnd = model.Time(end)
	return nil
}

func (u *WorksheetUC) GetWorksheet(ctx context.Context, worksheetUUID string) (model.WorksheetResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.WorksheetResponse{}, err
	}

	w, err := u.loadWorksheet(ctx, userDetail.InstitutionID, worksheetUUID, wrapMsgGetWorksheet)
	if err != nil {
		return model.WorksheetResponse{}, err
	}
	return w.ToResponse(), nil
}

func (u *WorksheetUC) PatchWorksheet(ctx context.Context, req model.PatchWorksheetRequest) (model.WorksheetResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.WorksheetResponse{}, err
	}

	w, err := u.loadWorksheet(ctx, userDetail.InstitutionID, req.UUID, wrapMsgPatchWorksheet)
	if err != nil {
		return model.WorksheetResponse{}, err
	}

	if worksheetGenerating(w) {
		return model.WorksheetResponse{}, commonerr.SetNewError(http.StatusConflict, errWorksheetGenerateBusy, msgWorksheetGenerateBusy)
	}
	if w.Status != model.WorksheetStatusOpen {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetNotEditable, msgWorksheetNotEditable)
	}

	if req.Label.Valid {
		label := strings.TrimSpace(req.Label.String)
		if label == "" || len(label) > maxPeriodLabelLen {
			return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errInvalidLabel, msgInvalidLabel)
		}
		w.Label = label
	}

	if err := u.WorksheetDB.Update(ctx, w); err != nil {
		return model.WorksheetResponse{}, errors.Wrap(err, wrapMsgPatchWorksheet)
	}
	return w.ToResponse(), nil
}

func (u *WorksheetUC) DeleteWorksheet(ctx context.Context, worksheetUUID string) (model.DeleteWorksheetResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.DeleteWorksheetResponse{}, err
	}

	w, err := u.loadWorksheet(ctx, userDetail.InstitutionID, worksheetUUID, wrapMsgDeleteWorksheet)
	if err != nil {
		return model.DeleteWorksheetResponse{}, err
	}
	if w.Status == model.WorksheetStatusFinalized {
		return model.DeleteWorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetDeleteOnly, msgWorksheetDeleteOnly)
	}
	if worksheetGenerating(w) {
		return model.DeleteWorksheetResponse{}, commonerr.SetNewError(http.StatusConflict, errWorksheetGenerateBusy, msgWorksheetGenerateBusy)
	}
	if err = u.rejectIfPaydayFinalized(ctx, w, wrapMsgDeleteWorksheet); err != nil {
		return model.DeleteWorksheetResponse{}, err
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.DeleteWorksheetResponse{}, errors.Wrap(err, wrapMsgDeleteWorksheet)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	if _, err = u.CommissionDB.SoftDeleteByWorksheet(ctx, w.ID); err != nil {
		return model.DeleteWorksheetResponse{}, errors.Wrap(err, wrapMsgDeleteWorksheet)
	}
	if err = u.removeWorksheetFromPayday(ctx, userDetail.InstitutionID, w); err != nil {
		return model.DeleteWorksheetResponse{}, err
	}

	found, err := u.WorksheetDB.SoftDelete(ctx, userDetail.InstitutionID, worksheetUUID)
	if err != nil {
		return model.DeleteWorksheetResponse{}, errors.Wrap(err, wrapMsgDeleteWorksheet)
	}
	if !found {
		return model.DeleteWorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}
	return model.DeleteWorksheetResponse{Success: true}, nil
}

func (u *WorksheetUC) FinalizeWorksheet(ctx context.Context, worksheetUUID string) (model.FinalizeWorksheetResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.FinalizeWorksheetResponse{}, err
	}

	w, err := u.loadWorksheet(ctx, userDetail.InstitutionID, worksheetUUID, wrapMsgFinalizeWorksheet)
	if err != nil {
		return model.FinalizeWorksheetResponse{}, err
	}

	if w.Status == model.WorksheetStatusFinalized {
		visitIDs, err := u.CommissionDB.DistinctVisitIDsByWorksheet(ctx, w.ID)
		if err != nil {
			return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
		}
		return model.FinalizeWorksheetResponse{
			Worksheet:        w.ToResponse(),
			LockedVisitCount: int64(len(visitIDs)),
		}, nil
	}

	if worksheetGenerating(w) {
		return model.FinalizeWorksheetResponse{}, commonerr.SetNewError(http.StatusConflict, errWorksheetGenerateBusy, msgWorksheetGenerateBusy)
	}
	if w.Status != model.WorksheetStatusOpen {
		return model.FinalizeWorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetFinalizeOnly, msgWorksheetFinalizeOnly)
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	visitIDs, err := u.CommissionDB.DistinctVisitIDsByWorksheet(ctx, w.ID)
	if err != nil {
		return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
	}

	// Sum approved commissions for total.
	totalCommission, err := u.CommissionDB.SumValidByWorksheet(ctx, w.ID)
	if err != nil {
		return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
	}

	now := u.currentTime()
	oldCommission := w.TotalCommission
	if w.CompensationPeriodID.Valid {
		period, found, pErr := u.CompensationPeriodDB.GetByID(ctx, userDetail.InstitutionID, w.CompensationPeriodID.Int64)
		if pErr != nil {
			return model.FinalizeWorksheetResponse{}, errors.Wrap(pErr, wrapMsgFinalizeWorksheet)
		}
		if found && period != nil && period.Status == model.CompensationPeriodStatusFinalized && totalCommission != oldCommission {
			return model.FinalizeWorksheetResponse{}, commonerr.SetNewBadRequest(errPaydayPeriodFinalized, msgPaydayPeriodFinalized)
		}
		if (period == nil || period.Status != model.CompensationPeriodStatusFinalized) && totalCommission != oldCommission {
			if err = u.CompensationPeriodDB.AddCommissionDelta(ctx, userDetail.InstitutionID, w.CompensationPeriodID.Int64, totalCommission-oldCommission); err != nil {
				return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
			}
		}
	}

	// Update worksheet status to finalized.
	w.Status = model.WorksheetStatusFinalized
	w.TotalCommission = totalCommission
	w.VisitCount = int64(len(visitIDs))
	w.FinalizedAt = sql.NullTime{Time: now, Valid: true}
	w.FinalizedBy = sql.NullString{String: userDetail.UUID, Valid: true}
	if err = u.WorksheetDB.Update(ctx, w); err != nil {
		return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
	}

	// Lock the visits.
	lockedCount, err := u.VisitLockDB.LockVisits(ctx, w.ID, visitIDs, now)
	if err != nil {
		return model.FinalizeWorksheetResponse{}, errors.Wrap(err, wrapMsgFinalizeWorksheet)
	}

	return model.FinalizeWorksheetResponse{
		Worksheet:        w.ToResponse(),
		LockedVisitCount: lockedCount,
	}, nil
}

func worksheetGenerating(w *model.TrxWorksheet) bool {
	return w != nil && w.GenerateStatus == model.WorksheetGenerateStatusRunning
}

func rejectIfPaydayFinalized(period *model.TrxCompensationPeriod) error {
	if period != nil && period.Status == model.CompensationPeriodStatusFinalized {
		return commonerr.SetNewBadRequest(errPaydayPeriodFinalized, msgPaydayPeriodFinalized)
	}
	return nil
}

func (u *WorksheetUC) rejectIfPaydayFinalized(ctx context.Context, w *model.TrxWorksheet, wrap string) error {
	if w == nil || !w.CompensationPeriodID.Valid {
		return nil
	}
	period, found, err := u.CompensationPeriodDB.GetByID(ctx, w.InstitutionID, w.CompensationPeriodID.Int64)
	if err != nil {
		return errors.Wrap(err, wrap)
	}
	if !found {
		return nil
	}
	return rejectIfPaydayFinalized(period)
}

func (u *WorksheetUC) removeWorksheetFromPayday(ctx context.Context, institutionID int64, w *model.TrxWorksheet) error {
	if w == nil || !w.CompensationPeriodID.Valid {
		return nil
	}
	if w.TotalCommission != 0 {
		if err := u.CompensationPeriodDB.AddCommissionDelta(ctx, institutionID, w.CompensationPeriodID.Int64, -w.TotalCommission); err != nil {
			return errors.Wrap(err, wrapMsgDeleteWorksheet)
		}
	}
	if u.SnapshotDB == nil {
		return nil
	}
	snap, found, err := u.SnapshotDB.GetByStaffPeriod(ctx, institutionID, w.StaffID, w.CompensationPeriodID.Int64)
	if err != nil {
		return errors.Wrap(err, wrapMsgDeleteWorksheet)
	}
	if !found || snap == nil {
		return nil
	}
	deleted, err := u.SnapshotDB.SoftDelete(ctx, institutionID, snap.ID)
	if err != nil {
		return errors.Wrap(err, wrapMsgDeleteWorksheet)
	}
	if deleted && snap.TotalWage.Valid && snap.TotalWage.Int64 != 0 {
		if err = u.CompensationPeriodDB.AddWageDelta(ctx, institutionID, w.CompensationPeriodID.Int64, -snap.TotalWage.Int64); err != nil {
			return errors.Wrap(err, wrapMsgDeleteWorksheet)
		}
	}
	return nil
}

func (u *WorksheetUC) loadWorksheet(ctx context.Context, institutionID int64, worksheetUUID, wrap string) (*model.TrxWorksheet, error) {
	if worksheetUUID == "" {
		return nil, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}
	w, found, err := u.WorksheetDB.GetByUUID(ctx, institutionID, worksheetUUID)
	if err != nil {
		return nil, errors.Wrap(err, wrap)
	}
	if !found || w == nil {
		return nil, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}
	return w, nil
}
