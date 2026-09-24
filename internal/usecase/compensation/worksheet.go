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

	errWorksheetNotFound       = "WORKSHEET_NOT_FOUND"
	errWorksheetOverlap        = "WORKSHEET_DATE_RANGE_OVERLAP"
	errWorksheetFinalizeOnly   = "WORKSHEET_MUST_BE_OPEN_TO_FINALIZE"
	errWorksheetDeleteOnly     = "WORKSHEET_MUST_BE_PENDING_OR_OPEN_TO_DELETE"
	errWorksheetNotEditable    = "WORKSHEET_NOT_EDITABLE"
	errWorksheetLinkedToPayday = "WORKSHEET_LINKED_TO_PAYDAY"
	errWorksheetGenerateBusy   = "WORKSHEET_GENERATE_PENDING"
	errPaydayPeriodNotFound    = "PAYDAY_PERIOD_NOT_FOUND"
	errPaydayPeriodFinalized   = "PAYDAY_PERIOD_FINALIZED"

	msgWorksheetNotFound       = "worksheet was not found"
	msgWorksheetOverlap        = "worksheet date range overlaps an existing worksheet for this staff"
	msgWorksheetFinalizeOnly   = "worksheet must be open to finalize"
	msgWorksheetDeleteOnly     = "worksheet must be pending or open to delete"
	msgWorksheetNotEditable    = "worksheet must be open to edit"
	msgWorksheetLinkedToPayday = "cannot delete worksheet while linked to a payday period"
	msgWorksheetGenerateBusy   = "worksheet generate is in progress"
	msgPaydayPeriodNotFound    = "compensation period was not found"
	msgPaydayPeriodFinalized   = "cannot attach worksheet to a finalized payday period"
)

var _ compensationuc.WorksheetUC = (*WorksheetUC)(nil)

// WorksheetUC implements WorksheetUC.
type WorksheetUC struct {
	WorksheetDB          compensationrepo.WorksheetDB
	CommissionDB         compensationrepo.CommissionDB
	ContributorDB        compensationrepo.ContributorDB
	CompensationPeriodDB compensationrepo.CompensationPeriodDB
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

	start := utilcommon.DateOnly(req.PeriodStart.Time())
	end := utilcommon.DateOnly(req.PeriodEnd.Time())
	if start.IsZero() || end.IsZero() || end.Before(start) {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errInvalidPeriodDates, msgInvalidPeriodDates)
	}

	overlaps, err := u.WorksheetDB.ExistsOverlapping(ctx, userDetail.InstitutionID, req.StaffID, start, end, 0)
	if err != nil {
		return model.WorksheetResponse{}, errors.Wrap(err, wrapMsgCreateWorksheet)
	}
	if overlaps {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetOverlap, msgWorksheetOverlap)
	}

	w := &model.TrxWorksheet{
		InstitutionID: userDetail.InstitutionID,
		StaffID:       req.StaffID,
		Label:         label,
		PeriodStart:   start,
		PeriodEnd:     end,
		CreatedBy:     sql.NullString{String: userDetail.UUID, Valid: true},
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

	if w.Status == model.WorksheetStatusPending {
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
	if req.PeriodStart != nil {
		w.PeriodStart = utilcommon.DateOnly(req.PeriodStart.Time())
	}
	if req.PeriodEnd != nil {
		w.PeriodEnd = utilcommon.DateOnly(req.PeriodEnd.Time())
	}
	if w.PeriodStart.IsZero() || w.PeriodEnd.IsZero() || w.PeriodEnd.Before(w.PeriodStart) {
		return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errInvalidPeriodDates, msgInvalidPeriodDates)
	}
	if req.PeriodStart != nil || req.PeriodEnd != nil {
		overlaps, overlapErr := u.WorksheetDB.ExistsOverlapping(ctx, userDetail.InstitutionID, w.StaffID, w.PeriodStart, w.PeriodEnd, w.ID)
		if overlapErr != nil {
			return model.WorksheetResponse{}, errors.Wrap(overlapErr, wrapMsgPatchWorksheet)
		}
		if overlaps {
			return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetOverlap, msgWorksheetOverlap)
		}
	}

	if req.CompensationPeriodUUID != nil {
		if !req.CompensationPeriodUUID.Valid || strings.TrimSpace(req.CompensationPeriodUUID.String) == "" {
			// JSON null or empty → detach payday link.
			w.CompensationPeriodID = sql.NullInt64{}
			w.CompensationPeriodUUID = sql.NullString{}
		} else {
			periodUUID := strings.TrimSpace(req.CompensationPeriodUUID.String)
			period, found, pErr := u.CompensationPeriodDB.GetByUUID(ctx, userDetail.InstitutionID, periodUUID)
			if pErr != nil {
				return model.WorksheetResponse{}, errors.Wrap(pErr, wrapMsgPatchWorksheet)
			}
			if !found || period == nil {
				return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errPaydayPeriodNotFound, msgPaydayPeriodNotFound)
			}
			if period.Status == model.CompensationPeriodStatusFinalized {
				return model.WorksheetResponse{}, commonerr.SetNewBadRequest(errPaydayPeriodFinalized, msgPaydayPeriodFinalized)
			}
			w.CompensationPeriodID = sql.NullInt64{Int64: period.ID, Valid: true}
			w.CompensationPeriodUUID = sql.NullString{String: period.UUID, Valid: true}
		}
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
	if w.CompensationPeriodID.Valid {
		return model.DeleteWorksheetResponse{}, commonerr.SetNewBadRequest(errWorksheetLinkedToPayday, msgWorksheetLinkedToPayday)
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
