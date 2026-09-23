package compensation

import (
	"context"
	"database/sql"
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
	wrapCompensationPeriodUCPrefix = "CompensationPeriodUC."
	wrapMsgCreatePeriod            = wrapCompensationPeriodUCPrefix + "CreatePeriod"
	wrapMsgListPeriods             = wrapCompensationPeriodUCPrefix + "ListPeriods"
	wrapMsgGetPeriod               = wrapCompensationPeriodUCPrefix + "GetPeriod"
	wrapMsgDraftPeriod             = wrapCompensationPeriodUCPrefix + "DraftPeriod"
	wrapMsgFinalizePeriod          = wrapCompensationPeriodUCPrefix + "FinalizePeriod"
	wrapMsgReopenPeriod            = wrapCompensationPeriodUCPrefix + "ReopenPeriod"
	wrapMsgDeletePeriod            = wrapCompensationPeriodUCPrefix + "DeletePeriod"
	wrapMsgListPeriodStaff         = wrapCompensationPeriodUCPrefix + "ListPeriodStaff"
	wrapMsgGetPeriodStaff          = wrapCompensationPeriodUCPrefix + "GetPeriodStaff"
	defaultListLimit               = 50
	maxPeriodLabelLen              = 100
	errIllegalTransition           = "ILLEGAL_PERIOD_TRANSITION"
	errInvalidPeriodStatus         = "INVALID_COMPENSATION_PERIOD_STATUS"
	errDateRangeOverlap            = "PERIOD_DATE_RANGE_OVERLAP"
	errPeriodNotFound              = "period_not_found"
	errInvalidPeriodDates          = "invalid_period_dates"
	errInvalidLabel                = "invalid_label"
	errInvalidCommissionType       = "INVALID_COMMISSION_TYPE"
	errCommissionNotFound          = "COMMISSION_NOT_FOUND"
	errInvalidCommissionPercent    = "INVALID_COMMISSION_PERCENT"
	errInvalidCommissionFlatAmount = "INVALID_COMMISSION_FLAT_AMOUNT"
	labelSourceProductName         = "product_name"
	labelSourceICD10Display        = "icd10_display"

	msgInvalidLabel                = "label is required and must be at most 100 characters"
	msgInvalidPeriodDates          = "period_start and period_end are required and period_end must not be before period_start"
	msgInvalidPeriodStatus         = "status must be open, draft, or finalized"
	msgDraftFromOpenOrDraft        = "period can only be drafted from open or draft status"
	msgReopenFromFinalized         = "period can only be reopened from finalized status"
	msgDeleteOpenOnly              = "only an open period can be deleted"
	msgPeriodNotFound              = "period was not found"
	msgDateRangeOverlap            = "period date range overlaps an existing period"
	msgFinalizeFromDraftOnly       = "period can only be finalized from draft status"
	msgInvalidCommissionType       = "commission_type must be percent or flat"
	msgCommissionNotFound          = "commission item was not found"
	msgInvalidCommissionPercent    = "commission_percent is required and must be greater than or equal to 0"
	msgInvalidCommissionFlatAmount = "commission_flat_amount is required and must be greater than or equal to 0"
)

var _ compensationuc.CompensationPeriodUC = (*CompensationPeriodUC)(nil)

type CompensationPeriodUC struct {
	CompensationPeriodDB compensationrepo.CompensationPeriodDB
	WorksheetDB          compensationrepo.WorksheetDB
	ContributorDB        compensationrepo.ContributorDB
	StaffDB              staffrepo.StaffDB
	Transaction          xormlib.DBTransactionInterface
	now                  func() time.Time
}

func NewCompensationPeriodUC(uc *CompensationPeriodUC) *CompensationPeriodUC {
	return uc
}

func (u *CompensationPeriodUC) currentTime() time.Time {
	if u.now != nil {
		return u.now()
	}
	return time.Now().UTC()
}

func (u *CompensationPeriodUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	return userDetail, nil
}

func (u *CompensationPeriodUC) CreatePeriod(ctx context.Context, req model.CreateCompensationPeriodRequest) (model.CompensationPeriodResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.CompensationPeriodResponse{}, err
	}

	label := strings.TrimSpace(req.Label)
	if label == "" || len(label) > maxPeriodLabelLen {
		return model.CompensationPeriodResponse{}, commonerr.SetNewBadRequest(errInvalidLabel, msgInvalidLabel)
	}

	start := utilcommon.DateOnly(req.PeriodStart.Time())
	end := utilcommon.DateOnly(req.PeriodEnd.Time())
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
	if err := u.CompensationPeriodDB.Create(ctx, period); err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgCreatePeriod)
	}

	return period.ToResponse(0), nil
}

func (u *CompensationPeriodUC) ListPeriods(ctx context.Context, req model.ListCompensationPeriodsRequest) (model.ListCompensationPeriodsResponse, error) {
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

	rows, total, err := u.CompensationPeriodDB.List(ctx, model.ListCompensationPeriodParams{
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

func (u *CompensationPeriodUC) GetPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error) {
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

func (u *CompensationPeriodUC) ListPeriodStaff(ctx context.Context, periodUUID string, req model.ListCompensationPeriodStaffRequest) (model.ListCompensationPeriodStaffResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListCompensationPeriodStaffResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, periodUUID, wrapMsgListPeriodStaff)
	if err != nil {
		return model.ListCompensationPeriodStaffResponse{}, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	detections, err := u.ContributorDB.DetectStaffForPeriod(ctx, userDetail.InstitutionID, period.PeriodStart, period.PeriodEnd.AddDate(0, 0, 1))
	if err != nil {
		return model.ListCompensationPeriodStaffResponse{}, errors.Wrap(err, wrapMsgListPeriodStaff)
	}

	// Get per-staff commission rollup from worksheets attached to this period.
	worksheetTotals, err := u.WorksheetDB.SumByStaffForCompensationPeriod(ctx, period.ID)
	if err != nil {
		return model.ListCompensationPeriodStaffResponse{}, errors.Wrap(err, wrapMsgListPeriodStaff)
	}
	commissionsByStaff := make(map[string]compensationrepo.StaffCommissionTotals, len(worksheetTotals))
	for _, row := range worksheetTotals {
		commissionsByStaff[row.StaffID] = row
	}

	staff := make([]model.CompensationPeriodStaffRow, 0, len(detections))
	for _, detected := range detections {
		roles := detected.Roles
		if roles == nil {
			roles = []string{}
		}
		var commissionSubtotal, commissionedVisits int64
		if totals, ok := commissionsByStaff[detected.StaffID]; ok {
			commissionSubtotal = totals.TotalCommission
			commissionedVisits = totals.VisitCount
		}
		staff = append(staff, model.CompensationPeriodStaffRow{
			StaffID:            detected.StaffID,
			Name:               detected.Name,
			Roles:              roles,
			VisitCount:         detected.VisitCount,
			Wage:               0,
			CommissionSubtotal: commissionSubtotal,
			PayTotal:           commissionSubtotal,
			AssignmentStatus:   assignmentStatus(detected.VisitCount, commissionedVisits),
		})
	}

	total := len(staff)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	return model.ListCompensationPeriodStaffResponse{
		Staff: staff[start:end],
		Total: total,
	}, nil
}

func (u *CompensationPeriodUC) GetPeriodStaff(ctx context.Context, req model.GetCompensationPeriodStaffRequest) (model.GetCompensationPeriodStaffResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.GetCompensationPeriodStaffResponse{}, err
	}

	if _, err := u.loadPeriod(ctx, userDetail.InstitutionID, req.PeriodUUID, wrapMsgGetPeriodStaff); err != nil {
		return model.GetCompensationPeriodStaffResponse{}, err
	}

	staff, err := u.StaffDB.GetStaffByUUID(ctx, userDetail.InstitutionID, req.StaffID, true)
	if err != nil {
		return model.GetCompensationPeriodStaffResponse{}, err
	}

	return model.GetCompensationPeriodStaffResponse{
		StaffInfo:    staffInfoFrom(staff),
		ComputedWage: 0,
		WageOverride: null.Int64{},
	}, nil
}

func (u *CompensationPeriodUC) DraftPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error) {
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

	// Rollup totals from worksheets attached to this period.
	totals, err := u.WorksheetDB.SumByCompensationPeriod(ctx, period.ID)
	if err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgDraftPeriod)
	}

	applyPhase1DraftTotals(period, totals, u.currentTime(), userDetail.UUID)
	if err := u.CompensationPeriodDB.UpdateStatusAndTotals(ctx, period); err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgDraftPeriod)
	}
	return period.ToResponse(0), nil
}

func (u *CompensationPeriodUC) ReopenPeriod(ctx context.Context, periodUUID string) (model.CompensationPeriodResponse, error) {
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
	if err := u.CompensationPeriodDB.UpdateStatusAndTotals(ctx, period); err != nil {
		return model.CompensationPeriodResponse{}, errors.Wrap(err, wrapMsgReopenPeriod)
	}
	return period.ToResponse(0), nil
}

func (u *CompensationPeriodUC) DeletePeriod(ctx context.Context, periodUUID string) (model.DeleteCompensationPeriodResponse, error) {
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

	found, err := u.CompensationPeriodDB.SoftDelete(ctx, userDetail.InstitutionID, periodUUID)
	if err != nil {
		return model.DeleteCompensationPeriodResponse{}, errors.Wrap(err, wrapMsgDeletePeriod)
	}
	if !found {
		return model.DeleteCompensationPeriodResponse{}, commonerr.SetNewBadRequest(errPeriodNotFound, msgPeriodNotFound)
	}
	return model.DeleteCompensationPeriodResponse{Success: true}, nil
}

func (u *CompensationPeriodUC) loadPeriod(ctx context.Context, institutionID int64, periodUUID, wrap string) (*model.TrxCompensationPeriod, error) {
	if periodUUID == "" {
		return nil, commonerr.SetNewBadRequest(errPeriodNotFound, msgPeriodNotFound)
	}
	period, found, err := u.CompensationPeriodDB.GetByUUID(ctx, institutionID, periodUUID)
	if err != nil {
		return nil, errors.Wrap(err, wrap)
	}
	if !found || period == nil {
		return nil, commonerr.SetNewBadRequest(errPeriodNotFound, msgPeriodNotFound)
	}
	return period, nil
}

func (u *CompensationPeriodUC) rejectIfOverlapping(ctx context.Context, institutionID int64, start, end time.Time, excludeUUID string) error {
	rows, _, err := u.CompensationPeriodDB.List(ctx, model.ListCompensationPeriodParams{
		InstitutionID: institutionID,
	})
	if err != nil {
		return errors.Wrap(err, wrapCompensationPeriodUCPrefix+"rejectIfOverlapping")
	}
	for _, row := range rows {
		if excludeUUID != "" && row.UUID == excludeUUID {
			continue
		}
		if utilcommon.PeriodsOverlap(start, end, row.PeriodStart, row.PeriodEnd) {
			return commonerr.SetNewBadRequest(errDateRangeOverlap, msgDateRangeOverlap)
		}
	}
	return nil
}

func applyPhase1DraftTotals(period *model.TrxCompensationPeriod, totals compensationrepo.WorksheetPeriodTotals, now time.Time, staffUUID string) {
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

func staffInfoFrom(staff model.StaffWithRolesResponse) model.CompensationPeriodStaffInfo {
	roles := make([]string, 0, len(staff.Roles))
	for _, role := range staff.Roles {
		if role.Name != "" {
			roles = append(roles, role.Name)
		}
	}
	return model.CompensationPeriodStaffInfo{
		StaffID: staff.UUID,
		Name:    staff.Name,
		Roles:   roles,
	}
}

func assignmentStatus(visitCount, commissionedVisitCount int64) model.CompensationAssignmentStatus {
	if commissionedVisitCount <= 0 {
		return model.CompensationAssignmentStatusUnassigned
	}
	if commissionedVisitCount < visitCount {
		return model.CompensationAssignmentStatusPartial
	}
	return model.CompensationAssignmentStatusComplete
}

