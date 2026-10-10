package compensation

import (
	"context"
	"database/sql"
	stderrors "errors"
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
	"github.com/volatiletech/null/v8"
)

const (
	wrapSnapshotUCPrefix = "WageSnapshotUC."
	wrapMsgGenerateSnap  = wrapSnapshotUCPrefix + "Generate"
	wrapMsgListSnap      = wrapSnapshotUCPrefix + "List"
	wrapMsgUpdateSnap    = wrapSnapshotUCPrefix + "Update"
	wrapMsgDeleteSnap    = wrapSnapshotUCPrefix + "Delete"

	errSnapshotPeriodRequired = "COMPENSATION_PERIOD_REQUIRED"
	errWageContractNotFound   = "WAGE_CONTRACT_NOT_FOUND"
	errSnapshotNotFound       = "STAFF_WAGE_SNAPSHOT_NOT_FOUND"

	msgSnapshotPeriodRequired = "compensation_period_uuid is required"
	msgWageContractNotFound   = "staff has no wage contract for this period"
	msgSnapshotNotFound       = "wage snapshot was not found"
)

var _ compensationuc.WageSnapshotUC = (*WageSnapshotUC)(nil)

// WageSnapshotUC generates, lists, and deletes wage snapshots.
type WageSnapshotUC struct {
	WageDB      compensationrepo.WageDB
	SnapshotDB  compensationrepo.WageSnapshotDB
	PeriodDB    compensationrepo.CompensationPeriodDB
	Transaction xormlib.DBTransactionInterface
	now         func() time.Time
}

func NewWageSnapshotUC(uc *WageSnapshotUC) *WageSnapshotUC {
	if uc.now == nil {
		uc.now = time.Now
	}
	return uc
}

func (u *WageSnapshotUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	return userDetail, nil
}

func (u *WageSnapshotUC) Generate(ctx context.Context, req model.GenerateStaffWageSnapshotsRequest) (model.StaffWageSnapshotsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.StaffWageSnapshotsResponse{}, err
	}

	periodUUID := strings.TrimSpace(req.CompensationPeriodUUID)
	if periodUUID == "" {
		return model.StaffWageSnapshotsResponse{}, commonerr.SetNewBadRequest(errSnapshotPeriodRequired, msgSnapshotPeriodRequired)
	}
	period, found, err := u.PeriodDB.GetByUUID(ctx, userDetail.InstitutionID, periodUUID)
	if err != nil {
		return model.StaffWageSnapshotsResponse{}, errors.Wrap(err, wrapMsgGenerateSnap)
	}
	if !found || period == nil {
		return model.StaffWageSnapshotsResponse{}, commonerr.SetNewBadRequest(errPaydayPeriodNotFound, msgPaydayPeriodNotFound)
	}
	start := utilcommon.DateOnly(period.PeriodStart)
	end := utilcommon.DateOnly(period.PeriodEnd)

	live, err := u.WageDB.ListLive(ctx, userDetail.InstitutionID)
	if err != nil {
		return model.StaffWageSnapshotsResponse{}, errors.Wrap(err, wrapMsgGenerateSnap)
	}
	covering := wagesCoveringPeriod(live, start, end)

	chosen := covering
	if ids := uniqueStaffIDs(req.StaffIDs); len(ids) > 0 {
		byStaff := make(map[string]model.MstStaffWage, len(covering))
		for _, row := range covering {
			byStaff[row.StaffID] = row
		}
		chosen = make([]model.MstStaffWage, 0, len(ids))
		for _, staffID := range ids {
			row, ok := byStaff[staffID]
			if !ok {
				return model.StaffWageSnapshotsResponse{}, commonerr.SetNewBadRequest(errWageContractNotFound, msgWageContractNotFound)
			}
			chosen = append(chosen, row)
		}
	}
	if len(chosen) == 0 {
		return model.StaffWageSnapshotsResponse{Snapshots: []model.StaffWageSnapshotResponse{}}, nil
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return model.StaffWageSnapshotsResponse{}, errors.Wrap(err, wrapMsgGenerateSnap)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	createdAt := u.now().UTC()
	snapshots := make([]model.StaffWageSnapshotResponse, 0, len(chosen))
	for _, wage := range chosen {
		row := &model.TrxWagePeriodSnapshot{
			StaffID:                wage.StaffID,
			InstitutionID:          userDetail.InstitutionID,
			CompensationPeriodID:   period.ID,
			CompensationPeriodUUID: sql.NullString{String: period.UUID, Valid: true},
			PeriodStart:            start,
			PeriodEnd:              end,
			WageAmount:             wage.WageAmount,
			WageCadence:            wage.WageCadence,
			CreateTime:             createdAt,
		}
		if err = u.SnapshotDB.Insert(ctx, row); err != nil {
			return model.StaffWageSnapshotsResponse{}, errors.Wrap(err, wrapMsgGenerateSnap)
		}
		snapshots = append(snapshots, row.ToResponse())
	}
	return model.StaffWageSnapshotsResponse{Snapshots: snapshots}, nil
}

func (u *WageSnapshotUC) List(ctx context.Context) (model.StaffWageSnapshotsResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.StaffWageSnapshotsResponse{}, err
	}
	rows, err := u.SnapshotDB.List(ctx, userDetail.InstitutionID)
	if err != nil {
		return model.StaffWageSnapshotsResponse{}, errors.Wrap(err, wrapMsgListSnap)
	}
	snapshots := make([]model.StaffWageSnapshotResponse, 0, len(rows))
	for _, row := range rows {
		snapshots = append(snapshots, row.ToResponse())
	}
	return model.StaffWageSnapshotsResponse{Snapshots: snapshots}, nil
}

func sqlNullInt64(v null.Int64) sql.NullInt64 {
	if !v.Valid {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v.Int64, Valid: true}
}

func (u *WageSnapshotUC) Update(ctx context.Context, id int64, req model.UpdateStaffWageSnapshotRequest) (model.StaffWageSnapshotResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.StaffWageSnapshotResponse{}, err
	}
	row, found, err := u.SnapshotDB.Get(ctx, userDetail.InstitutionID, id)
	if err != nil {
		return model.StaffWageSnapshotResponse{}, errors.Wrap(err, wrapMsgUpdateSnap)
	}
	if !found || row == nil {
		return model.StaffWageSnapshotResponse{}, commonerr.SetNewError(http.StatusNotFound, errSnapshotNotFound, msgSnapshotNotFound)
	}

	mandatory := sqlNullInt64(req.MandatoryWorkingDays)
	staffDays := sqlNullInt64(req.StaffWorkingDays)
	finalWage := sqlNullInt64(req.FinalWage)
	total, err := model.ApplySnapshotWage(row.WageAmount, mandatory, staffDays, finalWage)
	if err != nil {
		var inputErr *model.SnapshotWageInputError
		if stderrors.As(err, &inputErr) {
			return model.StaffWageSnapshotResponse{}, commonerr.SetNewBadRequest(inputErr.Code, inputErr.Message)
		}
		return model.StaffWageSnapshotResponse{}, errors.Wrap(err, wrapMsgUpdateSnap)
	}

	row.MandatoryWorkingDays = mandatory
	row.StaffWorkingDays = staffDays
	row.FinalWage = finalWage
	row.TotalWage = total
	found, err = u.SnapshotDB.UpdateWageInputs(ctx, row)
	if err != nil {
		return model.StaffWageSnapshotResponse{}, errors.Wrap(err, wrapMsgUpdateSnap)
	}
	if !found {
		return model.StaffWageSnapshotResponse{}, commonerr.SetNewError(http.StatusNotFound, errSnapshotNotFound, msgSnapshotNotFound)
	}
	return row.ToResponse(), nil
}

func (u *WageSnapshotUC) Delete(ctx context.Context, id int64) (model.DeleteStaffWageSnapshotResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.DeleteStaffWageSnapshotResponse{}, err
	}
	found, err := u.SnapshotDB.SoftDelete(ctx, userDetail.InstitutionID, id)
	if err != nil {
		return model.DeleteStaffWageSnapshotResponse{}, errors.Wrap(err, wrapMsgDeleteSnap)
	}
	if !found {
		return model.DeleteStaffWageSnapshotResponse{}, commonerr.SetNewError(http.StatusNotFound, errSnapshotNotFound, msgSnapshotNotFound)
	}
	return model.DeleteStaffWageSnapshotResponse{Success: true}, nil
}

// wagesCoveringPeriod keeps one wage per staff whose dates overlap start..end.
// When two contracts overlap the period, the later start is used for the whole period.
func wagesCoveringPeriod(rows []model.MstStaffWage, start, end time.Time) []model.MstStaffWage {
	best := make(map[string]model.MstStaffWage, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		if !utilcommon.PeriodsOverlap(utilcommon.DateOnly(row.EffectiveFrom), wageRangeEnd(row.EffectiveTo), start, end) {
			continue
		}
		prev, seen := best[row.StaffID]
		if !seen {
			order = append(order, row.StaffID)
			best[row.StaffID] = row
			continue
		}
		if laterWage(row, prev) {
			best[row.StaffID] = row
		}
	}
	out := make([]model.MstStaffWage, 0, len(order))
	for _, staffID := range order {
		out = append(out, best[staffID])
	}
	return out
}

func laterWage(a, b model.MstStaffWage) bool {
	af := utilcommon.DateOnly(a.EffectiveFrom)
	bf := utilcommon.DateOnly(b.EffectiveFrom)
	if af.Equal(bf) {
		return a.ID > b.ID
	}
	return af.After(bf)
}

func uniqueStaffIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
