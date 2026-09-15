package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	staffrepo "github.com/faisalhardin/medilink/internal/entity/repo/staff"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

const (
	wrapCompensationPeriodUCPrefix   = "CompensationPeriodUC."
	wrapMsgCreatePeriod              = wrapCompensationPeriodUCPrefix + "CreatePeriod"
	wrapMsgListPeriods               = wrapCompensationPeriodUCPrefix + "ListPeriods"
	wrapMsgGetPeriod                 = wrapCompensationPeriodUCPrefix + "GetPeriod"
	wrapMsgDraftPeriod               = wrapCompensationPeriodUCPrefix + "DraftPeriod"
	wrapMsgFinalizePeriod            = wrapCompensationPeriodUCPrefix + "FinalizePeriod"
	wrapMsgReopenPeriod              = wrapCompensationPeriodUCPrefix + "ReopenPeriod"
	wrapMsgDeletePeriod              = wrapCompensationPeriodUCPrefix + "DeletePeriod"
	wrapMsgListPeriodStaff           = wrapCompensationPeriodUCPrefix + "ListPeriodStaff"
	wrapMsgGetPeriodStaff            = wrapCompensationPeriodUCPrefix + "GetPeriodStaff"
	wrapMsgListPeriodStaffVisits     = wrapCompensationPeriodUCPrefix + "ListPeriodStaffVisits"
	wrapMsgGeneratePeriodStaffVisits = wrapCompensationPeriodUCPrefix + "GeneratePeriodStaffVisits"
	defaultListLimit                 = 50
	maxPeriodLabelLen                = 100
	errIllegalTransition             = "ILLEGAL_PERIOD_TRANSITION"
	errInvalidPeriodStatus           = "INVALID_COMPENSATION_PERIOD_STATUS"
	errDateRangeOverlap              = "PERIOD_DATE_RANGE_OVERLAP"
	errPeriodNotFound                = "period_not_found"
	errInvalidPeriodDates            = "invalid_period_dates"
	errInvalidLabel                  = "invalid_label"
	labelSourceProductName           = "product_name"
	labelSourceICD10Display          = "icd10_display"

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

var _ compensationuc.CompensationPeriodUC = (*CompensationPeriodUC)(nil)

type CompensationPeriodUC struct {
	CompensationPeriodDB compensationrepo.CompensationPeriodDB
	Commissions          compensationrepo.CommissionDB
	ContributorDB        compensationrepo.ContributorDB
	StaffDB              staffrepo.StaffDB
	VisitLockDB          compensationrepo.VisitLockDB
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

	commissionRows, err := u.Commissions.SumByStaff(ctx, period.ID)
	if err != nil {
		return model.ListCompensationPeriodStaffResponse{}, errors.Wrap(err, wrapMsgListPeriodStaff)
	}
	commissionsByStaff := make(map[string]compensationrepo.StaffCommissionTotals, len(commissionRows))
	for _, row := range commissionRows {
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

func (u *CompensationPeriodUC) ListPeriodStaffVisits(ctx context.Context, req model.ListCompensationPeriodStaffVisitsRequest) (model.ListCompensationPeriodStaffVisitsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListCompensationPeriodStaffVisitsResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, req.PeriodUUID, wrapMsgListPeriodStaffVisits)
	if err != nil {
		return model.ListCompensationPeriodStaffVisitsResponse{}, err
	}

	if _, err := u.StaffDB.GetStaffByUUID(ctx, userDetail.InstitutionID, req.StaffID, true); err != nil {
		return model.ListCompensationPeriodStaffVisitsResponse{}, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	commissionRows, total, err := u.Commissions.ListByPeriodStaff(ctx, compensationrepo.ListVisitCommissionParams{
		InstitutionID: userDetail.InstitutionID,
		PeriodID:      period.ID,
		StaffID:       req.StaffID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		return model.ListCompensationPeriodStaffVisitsResponse{}, errors.Wrap(err, wrapMsgListPeriodStaffVisits)
	}

	visits := make([]model.CompensationPeriodStaffVisitRow, 0, len(commissionRows))
	for _, commission := range commissionRows {
		visits = append(visits, staffVisitRowFromList(commission))
	}

	return model.ListCompensationPeriodStaffVisitsResponse{
		Visits: visits,
		Total:  total,
	}, nil
}

func (u *CompensationPeriodUC) GeneratePeriodStaffVisits(ctx context.Context, req model.GenerateCompensationPeriodStaffVisitsRequest) (model.GenerateCompensationPeriodStaffVisitsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.GenerateCompensationPeriodStaffVisitsResponse{}, err
	}

	period, err := u.loadPeriod(ctx, userDetail.InstitutionID, req.PeriodUUID, wrapMsgGeneratePeriodStaffVisits)
	if err != nil {
		return model.GenerateCompensationPeriodStaffVisitsResponse{}, err
	}

	if _, err := u.StaffDB.GetStaffByUUID(ctx, userDetail.InstitutionID, req.StaffID, true); err != nil {
		return model.GenerateCompensationPeriodStaffVisitsResponse{}, err
	}

	periodEndExclusive := period.PeriodEnd.AddDate(0, 0, 1)
	detections, err := u.ContributorDB.DetectForPeriodStaff(
		ctx, userDetail.InstitutionID, req.StaffID, period.PeriodStart, periodEndExclusive,
	)
	if err != nil {
		return model.GenerateCompensationPeriodStaffVisitsResponse{}, errors.Wrap(err, wrapMsgGeneratePeriodStaffVisits)
	}
	if len(detections) == 0 {
		return model.GenerateCompensationPeriodStaffVisitsResponse{GeneratedCount: 0}, nil
	}

	sourcesByVisit, manualByVisit := sourcesAndManualForStaff(req.StaffID, detections)
	visitIDs := visitIDsFromDetections(detections)

	rows := make([]model.TrxVisitCommission, 0, len(visitIDs))
	for _, visitID := range visitIDs {
		sources := sourcesByVisit[visitID]
		if sources == nil {
			sources = []model.ContributionSource{}
		}
		var sourcesJSON []byte
		sourcesJSON, err = json.Marshal(sources)
		if err != nil {
			return model.GenerateCompensationPeriodStaffVisitsResponse{}, errors.Wrap(err, wrapMsgGeneratePeriodStaffVisits)
		}
		rows = append(rows, model.TrxVisitCommission{
			PeriodID:             period.ID,
			VisitID:              visitID,
			StaffID:              req.StaffID,
			RevenueBase:          0,
			CommissionType:       model.CommissionTypeFlat,
			CommissionFlatAmount: sql.NullInt64{Int64: 0, Valid: true},
			CommissionAmount:     0,
			Sources:              sourcesJSON,
			IncludedManually:     manualByVisit[visitID],
		})
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.GenerateCompensationPeriodStaffVisitsResponse{}, errors.Wrap(err, wrapMsgGeneratePeriodStaffVisits)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	generated, err := u.Commissions.InsertGeneratedIfMissing(ctx, rows)
	if err != nil {
		return model.GenerateCompensationPeriodStaffVisitsResponse{}, errors.Wrap(err, wrapMsgGeneratePeriodStaffVisits)
	}

	return model.GenerateCompensationPeriodStaffVisitsResponse{GeneratedCount: generated}, nil
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

func staffVisitRowFromList(commission compensationrepo.VisitCommissionListRow) model.CompensationPeriodStaffVisitRow {
	visitDate := ""
	if !commission.VisitDate.IsZero() {
		visitDate = commission.VisitDate.Format("2006-01-02")
	}
	row := model.CompensationPeriodStaffVisitRow{
		VisitID:     commission.VisitID,
		PatientName: commission.PatientName,
		VisitDate:   visitDate,
		Sources:     sourcesFromJSON(commission.Sources),
		RevenueBase: commission.RevenueBase,
	}
	row.HasContributors = len(row.Sources) > 0
	if commission.ApprovedAt.Valid {
		ct := commission.CommissionType
		row.CommissionType = &ct
		if commission.CommissionPercent.Valid {
			row.CommissionPercent = null.Float64{Float64: commission.CommissionPercent.Float64, Valid: true}
		}
		if commission.CommissionFlatAmount.Valid {
			row.CommissionFlatAmount = null.Int64{Int64: commission.CommissionFlatAmount.Int64, Valid: true}
		}
		row.CommissionAmount = null.Int64{Int64: commission.CommissionAmount, Valid: true}
	}
	return row
}

func sourcesFromJSON(raw json.RawMessage) []model.ContributionSource {
	out := make([]model.ContributionSource, 0)
	if len(raw) == 0 {
		return out
	}
	if err := json.Unmarshal(raw, &out); err != nil || out == nil {
		return []model.ContributionSource{}
	}
	return out
}

func visitIDsFromDetections(detections []compensationrepo.DetectedAttribution) []int64 {
	seen := make(map[int64]struct{}, len(detections))
	ids := make([]int64, 0, len(detections))
	for _, attr := range detections {
		if _, ok := seen[attr.VisitID]; ok {
			continue
		}
		seen[attr.VisitID] = struct{}{}
		ids = append(ids, attr.VisitID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func sourcesAndManualForStaff(staffID string, detections []compensationrepo.DetectedAttribution) (map[int64][]model.ContributionSource, map[int64]bool) {
	sourcesByVisit := make(map[int64][]model.ContributionSource)
	manualByVisit := make(map[int64]bool)
	for _, attr := range detections {
		if attr.StaffID != staffID {
			continue
		}
		sourcesByVisit[attr.VisitID] = append(sourcesByVisit[attr.VisitID], contributionSourceFrom(attr))
		if attr.Type == model.ContributionSourceTypeManual {
			manualByVisit[attr.VisitID] = true
		}
	}
	return sourcesByVisit, manualByVisit
}

func contributionSourceFrom(attr compensationrepo.DetectedAttribution) model.ContributionSource {
	src := model.ContributionSource{Type: attr.Type}
	switch attr.Type {
	case model.ContributionSourceTypeProcedure:
		if attr.ProcedureID != 0 {
			src.ProcedureID = null.Int64{Int64: attr.ProcedureID, Valid: true}
		}
		if attr.ProductID.Valid {
			src.ProductID = null.Int64{Int64: attr.ProductID.Int64, Valid: true}
		}
		if attr.Label.Valid {
			src.Label = null.String{String: attr.Label.String, Valid: true}
			src.LabelSource = null.String{String: labelSourceProductName, Valid: true}
		}
	case model.ContributionSourceTypeDiagnosis:
		if attr.DiagnosisID != 0 {
			src.DiagnosisID = null.Int64{Int64: attr.DiagnosisID, Valid: true}
		}
		if attr.Label.Valid {
			src.Label = null.String{String: attr.Label.String, Valid: true}
			src.LabelSource = null.String{String: labelSourceICD10Display, Valid: true}
		}
	}
	return src
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

	totals, err := u.Commissions.SumByPeriod(ctx, period.ID)
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
