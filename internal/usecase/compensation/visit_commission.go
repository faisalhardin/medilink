package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	permconst "github.com/faisalhardin/medilink/internal/entity/constant/permission"
	roleconst "github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

// jsonAPI matches project binding: stdlib-compatible jsoniter API.
var jsonAPI = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	wrapVisitCommissionUCPrefix  = "VisitCommissionUC."
	wrapMsgGenerate              = wrapVisitCommissionUCPrefix + "GenerateVisitCommissions"
	wrapMsgListCommissions       = wrapVisitCommissionUCPrefix + "ListVisitCommissions"
	wrapMsgPatchCommission       = wrapVisitCommissionUCPrefix + "PatchCommissionItem"
	wrapMsgArchiveCommission     = wrapVisitCommissionUCPrefix + "ArchiveVisitCommission"
	wrapMsgUpdateWorksheetTotals = wrapVisitCommissionUCPrefix + "updateWorksheetTotals"

	errWorksheetNotOpenForPatch     = "WORKSHEET_NOT_OPEN"
	msgWorksheetNotOpenForPatch     = "worksheet must be open to patch commission items"
	errWorksheetGeneratePending     = "WORKSHEET_GENERATE_PENDING"
	msgWorksheetGeneratePending     = "worksheet generate is in progress"
	errCommissionLinkedToPayday     = "COMMISSION_WORKSHEET_LINKED_TO_PAYDAY"
	msgCommissionLinkedToPayday     = "cannot modify commission while worksheet is linked to a payday period"
	errCommissionWorksheetFinalized = "COMMISSION_WORKSHEET_FINALIZED"
	msgCommissionWorksheetFinalized = "cannot archive commission on a finalized worksheet"
	errCommissionRevenueStale       = "COMMISSION_REVENUE_STALE"
	msgCommissionRevenueStale       = "visit revenue has changed; regenerate commission items before assigning"
	errForbiddenCommission          = "FORBIDDEN"
	msgForbiddenCommission          = "insufficient permissions for this staff's commissions"
)

var _ compensationuc.VisitCommissionUC = (*VisitCommissionUC)(nil)

// VisitCommissionUC implements VisitCommissionUC.
type VisitCommissionUC struct {
	WorksheetDB   compensationrepo.WorksheetDB
	CommissionDB  compensationrepo.CommissionDB
	ContributorDB compensationrepo.ContributorDB
	Transaction   xormlib.DBTransactionInterface
	now           func() time.Time
}

func NewVisitCommissionUC(uc *VisitCommissionUC) *VisitCommissionUC {
	return uc
}

func (u *VisitCommissionUC) currentTime() time.Time {
	if u.now != nil {
		return u.now()
	}
	return time.Now().UTC()
}

func (u *VisitCommissionUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	userDetail.EnsureAuthSets()
	return userDetail, nil
}

func canManageStaffCommissions(user model.UserJWTPayload, staffID string, assign bool) bool {
	if user.RolesIDSet[roleconst.Administrator] {
		return true
	}
	if assign {
		if user.PermissionsSet[permconst.CompensationAssign] {
			return true
		}
	} else if user.PermissionsSet[permconst.CompensationRead] || user.PermissionsSet[permconst.CompensationAssign] {
		return true
	}
	return user.UUID == staffID
}

// GenerateVisitCommissions triggers the async generate worker and returns 202.
func (u *VisitCommissionUC) GenerateVisitCommissions(ctx context.Context, req model.GenerateVisitCommissionsRequest) (model.GenerateVisitCommissionsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.GenerateVisitCommissionsResponse{}, err
	}

	if strings.TrimSpace(req.WorksheetID) == "" {
		return model.GenerateVisitCommissionsResponse{}, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}

	worksheet, found, err := u.WorksheetDB.GetByUUID(ctx, userDetail.InstitutionID, req.WorksheetID)
	if err != nil {
		return model.GenerateVisitCommissionsResponse{}, errors.Wrap(err, wrapMsgGenerate)
	}
	if !found || worksheet == nil {
		return model.GenerateVisitCommissionsResponse{}, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}

	if !canManageStaffCommissions(userDetail, worksheet.StaffID, true) {
		return model.GenerateVisitCommissionsResponse{}, commonerr.SetNewError(http.StatusForbidden, errForbiddenCommission, msgForbiddenCommission)
	}

	if worksheet.Status == model.WorksheetStatusPending {
		return model.GenerateVisitCommissionsResponse{}, commonerr.SetNewError(http.StatusConflict, errWorksheetGeneratePending, msgWorksheetGeneratePending)
	}
	if worksheet.Status != model.WorksheetStatusOpen {
		return model.GenerateVisitCommissionsResponse{}, commonerr.SetNewBadRequest(errWorksheetNotOpenForPatch, msgWorksheetNotOpenForPatch)
	}
	if worksheet.CompensationPeriodID.Valid {
		return model.GenerateVisitCommissionsResponse{}, commonerr.SetNewBadRequest(errCommissionLinkedToPayday, msgCommissionLinkedToPayday)
	}

	if !req.IncludeGenerate {
		return model.GenerateVisitCommissionsResponse{
			UUID:           worksheet.UUID,
			Status:         worksheet.Status,
			GenerateStatus: worksheet.GenerateStatus,
		}, nil
	}

	if err := u.WorksheetDB.MarkGeneratePending(ctx, worksheet.ID); err != nil {
		return model.GenerateVisitCommissionsResponse{}, errors.Wrap(err, wrapMsgGenerate)
	}

	go u.runGenerateWorker(worksheet)

	return model.GenerateVisitCommissionsResponse{
		UUID:           worksheet.UUID,
		Status:         model.WorksheetStatusPending,
		GenerateStatus: model.WorksheetGenerateStatusRunning,
	}, nil
}

// runGenerateWorker runs in a goroutine and must use a fresh background context.
// On any failure, defer marks the worksheet generate_status=failed (outside any TX).
func (u *VisitCommissionUC) runGenerateWorker(w *model.TrxWorksheet) {
	bgCtx := context.Background()
	var err error
	defer func() {
		if err == nil {
			return
		}
		_ = u.WorksheetDB.MarkGenerateFinished(bgCtx, w.ID, model.WorksheetGenerateStatusFailed, err.Error())
	}()

	periodEndExclusive := w.PeriodEnd.AddDate(0, 0, 1)

	var detections []compensationrepo.DetectedAttribution
	detections, err = u.ContributorDB.DetectForPeriodStaff(
		bgCtx, w.InstitutionID, w.StaffID, w.PeriodStart, periodEndExclusive,
	)
	if err != nil {
		return
	}

	visitIDs := visitIDsFromDetections(detections)
	var revenues map[int64]int64
	revenues, err = u.CommissionDB.SumRevenueByVisitIDs(bgCtx, visitIDs)
	if err != nil {
		return
	}

	var rows []model.TrxVisitCommission
	rows, err = buildCommissionRows(w.ID, w.StaffID, detections, revenues)
	if err != nil {
		return
	}

	session, err := u.Transaction.Begin(bgCtx)
	if err != nil {
		return
	}
	defer u.Transaction.Finish(session, &err)
	txCtx := xormlib.SetDBSession(bgCtx, session)

	if _, err = u.CommissionDB.InsertGeneratedIfMissing(txCtx, rows); err != nil {
		return
	}

	// Update totals before marking succeeded (valid = approved only; seeds stay 0).
	if _, err = u.updateWorksheetTotals(txCtx, w.ID); err != nil {
		return
	}
	if err = u.WorksheetDB.MarkGenerateFinished(txCtx, w.ID, model.WorksheetGenerateStatusSucceeded, ""); err != nil {
		return
	}
}

func buildCommissionRows(worksheetID int64, staffID string, detections []compensationrepo.DetectedAttribution, revenues map[int64]int64) ([]model.TrxVisitCommission, error) {
	sourcesByVisit, manualByVisit := sourcesAndManualForStaff(staffID, detections)
	visitIDs := visitIDsFromDetections(detections)

	rows := make([]model.TrxVisitCommission, 0, len(visitIDs))
	for _, visitID := range visitIDs {
		sources := sourcesByVisit[visitID]
		if sources == nil {
			sources = []model.ContributionSource{}
		}
		sourcesJSON, err := jsonAPI.Marshal(sources)
		if err != nil {
			return nil, err
		}
		revenueBase := int64(0)
		if revenues != nil {
			revenueBase = revenues[visitID]
		}
		rows = append(rows, model.TrxVisitCommission{
			WorksheetID:          worksheetID,
			VisitID:              visitID,
			StaffID:              staffID,
			RevenueBase:          revenueBase,
			CommissionType:       model.CommissionTypeFlat,
			CommissionFlatAmount: sql.NullInt64{Int64: 0, Valid: true},
			CommissionAmount:     0,
			Sources:              sourcesJSON,
			IncludedManually:     manualByVisit[visitID],
		})
	}
	return rows, nil
}

// ListVisitCommissions returns paginated commission rows for a worksheet or staff+date range.
func (u *VisitCommissionUC) ListVisitCommissions(ctx context.Context, req model.ListVisitCommissionsRequest) (model.ListVisitCommissionsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListVisitCommissionsResponse{}, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListLimit
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	var (
		commissionRows []compensationrepo.VisitCommissionListRow
		total          int
	)

	if strings.TrimSpace(req.WorksheetUUID) != "" {
		w, found, err := u.WorksheetDB.GetByUUID(ctx, userDetail.InstitutionID, req.WorksheetUUID)
		if err != nil {
			return model.ListVisitCommissionsResponse{}, errors.Wrap(err, wrapMsgListCommissions)
		}
		if !found || w == nil {
			return model.ListVisitCommissionsResponse{}, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
		}
		if !canManageStaffCommissions(userDetail, w.StaffID, false) {
			return model.ListVisitCommissionsResponse{}, commonerr.SetNewError(http.StatusForbidden, errForbiddenCommission, msgForbiddenCommission)
		}
		commissionRows, total, err = u.CommissionDB.ListByWorksheet(ctx, compensationrepo.ListVisitCommissionParams{
			InstitutionID: userDetail.InstitutionID,
			WorksheetID:   w.ID,
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			return model.ListVisitCommissionsResponse{}, errors.Wrap(err, wrapMsgListCommissions)
		}
	} else {
		if strings.TrimSpace(req.StaffID) == "" || time.Time(req.Start).IsZero() || time.Time(req.End).IsZero() {
			return model.ListVisitCommissionsResponse{}, commonerr.SetNewBadRequest("INVALID_LIST_FILTER", "staff_id, start, and end are required")
		}
		if !canManageStaffCommissions(userDetail, req.StaffID, false) {
			return model.ListVisitCommissionsResponse{}, commonerr.SetNewError(http.StatusForbidden, errForbiddenCommission, msgForbiddenCommission)
		}
		endExclusive := req.End.Time().AddDate(0, 0, 1)
		commissionRows, total, err = u.CommissionDB.ListByStaffDateRange(ctx, compensationrepo.ListStaffDateCommissionParams{
			InstitutionID: userDetail.InstitutionID,
			StaffID:       req.StaffID,
			Start:         req.Start.Time(),
			EndExclusive:  endExclusive,
			Limit:         limit,
			Offset:        offset,
		})
		if err != nil {
			return model.ListVisitCommissionsResponse{}, errors.Wrap(err, wrapMsgListCommissions)
		}
	}

	out := make([]model.VisitCommissionResponse, 0, len(commissionRows))
	for _, row := range commissionRows {
		out = append(out, commissionRowToResponse(row))
	}
	return model.ListVisitCommissionsResponse{
		Commissions: out,
		Total:       total,
	}, nil
}

// PatchCommissionItem updates commission amounts and re-sums worksheet totals if open.
func (u *VisitCommissionUC) PatchCommissionItem(ctx context.Context, req model.PatchCommissionItemRequest) (model.PatchCommissionItemResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.PatchCommissionItemResponse{}, err
	}

	if req.ID <= 0 {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errCommissionNotFound, msgCommissionNotFound)
	}

	row, found, err := u.CommissionDB.GetLiveByID(ctx, userDetail.InstitutionID, req.ID)
	if err != nil {
		return model.PatchCommissionItemResponse{}, errors.Wrap(err, wrapMsgPatchCommission)
	}
	if !found || row == nil {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errCommissionNotFound, msgCommissionNotFound)
	}

	if !req.CommissionType.IsValid() {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errInvalidCommissionType, msgInvalidCommissionType)
	}

	w, wFound, wErr := u.WorksheetDB.GetByID(ctx, row.WorksheetID)
	if wErr != nil {
		return model.PatchCommissionItemResponse{}, errors.Wrap(wErr, wrapMsgPatchCommission)
	}
	if !wFound || w == nil {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}
	if !canManageStaffCommissions(userDetail, w.StaffID, true) {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewError(http.StatusForbidden, errForbiddenCommission, msgForbiddenCommission)
	}
	if w.Status == model.WorksheetStatusPending {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewError(http.StatusConflict, errWorksheetGeneratePending, msgWorksheetGeneratePending)
	}
	if w.Status != model.WorksheetStatusOpen {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errWorksheetNotOpenForPatch, msgWorksheetNotOpenForPatch)
	}
	if w.CompensationPeriodID.Valid {
		return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errCommissionLinkedToPayday, msgCommissionLinkedToPayday)
	}

	// Option A: compare stored revenue_base to live full visit product cart.
	// Seed rows (never assigned) must be regenerated when the cart changed.
	// Already-assigned rows refresh revenue_base at assignment time.
	revenues, err := u.CommissionDB.SumRevenueByVisitIDs(ctx, []int64{row.VisitID})
	if err != nil {
		return model.PatchCommissionItemResponse{}, errors.Wrap(err, wrapMsgPatchCommission)
	}
	liveBase := int64(0)
	if base, ok := revenues[row.VisitID]; ok {
		liveBase = base
	}
	if liveBase != row.RevenueBase {
		if !row.ApprovedAt.Valid {
			return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errCommissionRevenueStale, msgCommissionRevenueStale)
		}
		row.RevenueBase = liveBase
	}

	switch req.CommissionType {
	case model.CommissionTypePercent:
		if !req.CommissionPercent.Valid || req.CommissionPercent.Float64 < 0 {
			return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errInvalidCommissionPercent, msgInvalidCommissionPercent)
		}
		row.CommissionType = model.CommissionTypePercent
		row.CommissionPercent = sql.NullFloat64{Float64: req.CommissionPercent.Float64, Valid: true}
		row.CommissionFlatAmount = sql.NullInt64{}
		row.CommissionAmount = int64(math.Round(float64(row.RevenueBase) * req.CommissionPercent.Float64 / 100))
	case model.CommissionTypeFlat:
		if !req.CommissionFlatAmount.Valid || req.CommissionFlatAmount.Int64 < 0 {
			return model.PatchCommissionItemResponse{}, commonerr.SetNewBadRequest(errInvalidCommissionFlatAmount, msgInvalidCommissionFlatAmount)
		}
		row.CommissionType = model.CommissionTypeFlat
		row.CommissionFlatAmount = sql.NullInt64{Int64: req.CommissionFlatAmount.Int64, Valid: true}
		row.CommissionPercent = sql.NullFloat64{}
		row.CommissionAmount = req.CommissionFlatAmount.Int64
	}

	if req.Note.Valid {
		row.Note = sql.NullString{String: req.Note.String, Valid: true}
	} else {
		row.Note = sql.NullString{}
	}

	now := u.currentTime()
	row.ApprovedAt = sql.NullTime{Time: now, Valid: true}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.PatchCommissionItemResponse{}, errors.Wrap(err, wrapMsgPatchCommission)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	if err = u.CommissionDB.UpdateAssignmentAmounts(ctx, row); err != nil {
		return model.PatchCommissionItemResponse{}, errors.Wrap(err, wrapMsgPatchCommission)
	}

	// Live re-sum: update worksheet total_commission when it's open.
	subtotal := int64(0)
	if w.Status == model.WorksheetStatusOpen {
		subtotal, err = u.updateWorksheetTotals(ctx, row.WorksheetID)
		if err != nil {
			return model.PatchCommissionItemResponse{}, errors.Wrap(err, wrapMsgPatchCommission)
		}
	}

	return model.PatchCommissionItemResponse{
		UpdatedCount:       1,
		CommissionSubtotal: subtotal,
	}, nil
}

// ArchiveVisitCommission soft-deletes a commission row when worksheet allows it.
func (u *VisitCommissionUC) ArchiveVisitCommission(ctx context.Context, id int64) (model.ArchiveVisitCommissionResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ArchiveVisitCommissionResponse{}, err
	}

	if id <= 0 {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewBadRequest(errCommissionNotFound, msgCommissionNotFound)
	}

	row, found, err := u.CommissionDB.GetLiveByID(ctx, userDetail.InstitutionID, id)
	if err != nil {
		return model.ArchiveVisitCommissionResponse{}, errors.Wrap(err, wrapMsgArchiveCommission)
	}
	if !found || row == nil {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewBadRequest(errCommissionNotFound, msgCommissionNotFound)
	}

	w, wFound, wErr := u.WorksheetDB.GetByID(ctx, row.WorksheetID)
	if wErr != nil {
		return model.ArchiveVisitCommissionResponse{}, errors.Wrap(wErr, wrapMsgArchiveCommission)
	}
	if !wFound || w == nil {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewBadRequest(errWorksheetNotFound, msgWorksheetNotFound)
	}
	if !canManageStaffCommissions(userDetail, w.StaffID, true) {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewError(http.StatusForbidden, errForbiddenCommission, msgForbiddenCommission)
	}
	if w.Status == model.WorksheetStatusPending {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewError(http.StatusConflict, errWorksheetGeneratePending, msgWorksheetGeneratePending)
	}
	if w.Status == model.WorksheetStatusFinalized {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewBadRequest(errCommissionWorksheetFinalized, msgCommissionWorksheetFinalized)
	}
	if w.CompensationPeriodID.Valid {
		return model.ArchiveVisitCommissionResponse{}, commonerr.SetNewBadRequest(errCommissionLinkedToPayday, msgCommissionLinkedToPayday)
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.ArchiveVisitCommissionResponse{}, errors.Wrap(err, wrapMsgArchiveCommission)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	ok, err := u.CommissionDB.SoftDelete(ctx, id)
	if err != nil {
		return model.ArchiveVisitCommissionResponse{}, errors.Wrap(err, wrapMsgArchiveCommission)
	}
	if ok && w.Status == model.WorksheetStatusOpen {
		if _, err = u.updateWorksheetTotals(ctx, w.ID); err != nil {
			return model.ArchiveVisitCommissionResponse{}, errors.Wrap(err, wrapMsgArchiveCommission)
		}
	}
	return model.ArchiveVisitCommissionResponse{Success: ok}, nil
}

// updateWorksheetTotals re-sums approved commissions and writes worksheet totals.
// visit_count is all live distinct visits; total_commission is approved-only.
func (u *VisitCommissionUC) updateWorksheetTotals(ctx context.Context, worksheetID int64) (int64, error) {
	totalCommission, err := u.CommissionDB.SumValidByWorksheet(ctx, worksheetID)
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgUpdateWorksheetTotals)
	}
	visitCount, err := u.CommissionDB.CountVisitsByWorksheet(ctx, worksheetID)
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgUpdateWorksheetTotals)
	}
	if err := u.WorksheetDB.UpdateTotals(ctx, worksheetID, totalCommission, visitCount); err != nil {
		return 0, errors.Wrap(err, wrapMsgUpdateWorksheetTotals)
	}
	return totalCommission, nil
}

// ─── helpers shared with worksheet generate ──────────────────────────────────

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

func sourcesFromJSON(raw json.RawMessage) []model.ContributionSource {
	out := make([]model.ContributionSource, 0)
	if len(raw) == 0 {
		return out
	}
	if err := jsonAPI.Unmarshal(raw, &out); err != nil || out == nil {
		return []model.ContributionSource{}
	}
	return out
}

func commissionRowToResponse(commission compensationrepo.VisitCommissionListRow) model.VisitCommissionResponse {
	visitDate := ""
	if !commission.VisitDate.IsZero() {
		visitDate = commission.VisitDate.Format("2006-01-02")
	}
	sources := sourcesFromJSON(commission.Sources)
	row := model.VisitCommissionResponse{
		ID:              commission.ID,
		WorksheetUUID:   commission.WorksheetUUID,
		VisitID:         commission.VisitID,
		PatientName:     commission.PatientName,
		VisitDate:       visitDate,
		Sources:         sources,
		RevenueBase:     commission.RevenueBase,
		HasContributors: len(sources) > 0,
	}

	if commission.ApprovedAt.Valid {
		row.ApprovedAt = null.TimeFrom(commission.ApprovedAt.Time)
		row.CommissionType = null.StringFrom(string(commission.CommissionType))
		if commission.CommissionPercent.Valid {
			row.CommissionPercent = null.Float64From(commission.CommissionPercent.Float64)
		}
		if commission.CommissionFlatAmount.Valid {
			row.CommissionFlatAmount = null.Int64From(commission.CommissionFlatAmount.Int64)
		}
		row.CommissionAmount = null.Int64From(commission.CommissionAmount)
	}
	return row
}
