package compensation

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-xorm/xorm"
)

const (
	testInstitutionID int64 = 42
	testStaffUUID           = "staff-uuid-1"
)

func testCtx() context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		InstitutionID: testInstitutionID,
		UUID:          testStaffUUID,
	})
}

func parseDate(s string) model.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return model.Time(t)
}

func createReq(label, start, end string) model.CreateCompensationPeriodRequest {
	return model.CreateCompensationPeriodRequest{
		Label:       label,
		PeriodStart: parseDate(start),
		PeriodEnd:   parseDate(end),
	}
}

func errorName(t *testing.T, err error) string {
	t.Helper()
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) {
		t.Fatalf("error type %T: %v", err, err)
	}
	if len(em.ErrorList) == 0 {
		t.Fatal("empty error_list")
	}
	return em.ErrorList[0].ErrorName
}

type fakePeriodDB struct {
	periods         []*model.TrxCompensationPeriod
	lastList        model.ListCompensationPeriodParams
	createErr       error
	getErr          error
	listErr         error
	updateErr       error
	deleteErr       error
	updateCalls     int
	softDeleteCalls int
}

func (f *fakePeriodDB) Create(_ context.Context, period *model.TrxCompensationPeriod) error {
	if f.createErr != nil {
		return f.createErr
	}
	if period.UUID == "" {
		period.UUID = "generated-uuid"
	}
	if period.Status == "" {
		period.Status = model.CompensationPeriodStatusOpen
	}
	period.ID = int64(len(f.periods) + 1)
	cp := *period
	f.periods = append(f.periods, &cp)
	return nil
}

func (f *fakePeriodDB) GetByUUID(_ context.Context, institutionID int64, uuid string) (*model.TrxCompensationPeriod, bool, error) {
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	for _, p := range f.periods {
		if p.UUID == uuid && p.InstitutionID == institutionID {
			cp := *p
			return &cp, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakePeriodDB) List(_ context.Context, params model.ListCompensationPeriodParams) ([]model.TrxCompensationPeriod, int, error) {
	f.lastList = params
	if f.listErr != nil {
		return nil, 0, f.listErr
	}
	matched := make([]model.TrxCompensationPeriod, 0, len(f.periods))
	for _, p := range f.periods {
		if p.InstitutionID != params.InstitutionID {
			continue
		}
		if params.Status != "" && p.Status != params.Status {
			continue
		}
		matched = append(matched, *p)
	}
	total := len(matched)
	if params.Limit > 0 {
		start := params.Offset
		if start > total {
			start = total
		}
		end := start + params.Limit
		if end > total {
			end = total
		}
		matched = matched[start:end]
	}
	return matched, total, nil
}

func (f *fakePeriodDB) UpdateStatusAndTotals(_ context.Context, period *model.TrxCompensationPeriod) error {
	f.updateCalls++
	if f.updateErr != nil {
		return f.updateErr
	}
	for i, p := range f.periods {
		if p.UUID == period.UUID && p.InstitutionID == period.InstitutionID {
			cp := *period
			f.periods[i] = &cp
			return nil
		}
	}
	return errors.New("period not in fake store")
}

func (f *fakePeriodDB) SoftDelete(_ context.Context, institutionID int64, uuid string) (bool, error) {
	f.softDeleteCalls++
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	for i, p := range f.periods {
		if p.UUID == uuid && p.InstitutionID == institutionID {
			f.periods = append(f.periods[:i], f.periods[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

type fakeCommissions struct {
	totals   map[int64]compensationrepo.PeriodCommissionTotals
	visits   map[int64][]int64
	sumErr   error
	visitErr error
}

func (f *fakeCommissions) SumByPeriod(_ context.Context, periodID int64) (compensationrepo.PeriodCommissionTotals, error) {
	if f.sumErr != nil {
		return compensationrepo.PeriodCommissionTotals{}, f.sumErr
	}
	if f.totals == nil {
		return compensationrepo.PeriodCommissionTotals{}, nil
	}
	return f.totals[periodID], nil
}

func (f *fakeCommissions) DistinctVisitIDsByPeriod(_ context.Context, periodID int64) ([]int64, error) {
	if f.visitErr != nil {
		return nil, f.visitErr
	}
	if f.visits == nil {
		return nil, nil
	}
	return append([]int64(nil), f.visits[periodID]...), nil
}

type fakeVisitLock struct {
	locked map[int64][]int64
	calls  int
	err    error
}

func (f *fakeVisitLock) LockVisits(_ context.Context, periodID int64, visitIDs []int64, _ time.Time) (int64, error) {
	f.calls++
	if f.err != nil {
		return 0, f.err
	}
	if f.locked == nil {
		f.locked = map[int64][]int64{}
	}
	f.locked[periodID] = append([]int64(nil), visitIDs...)
	return int64(len(visitIDs)), nil
}

type fakeTx struct {
	beginErr   error
	began      bool
	finished   bool
	rolledBack bool
}

func (t *fakeTx) Begin(_ context.Context) (*xorm.Session, error) {
	if t.beginErr != nil {
		return nil, t.beginErr
	}
	t.began = true
	return nil, nil
}

func (t *fakeTx) Finish(_ *xorm.Session, err *error) {
	t.finished = true
	if err != nil && *err != nil {
		t.rolledBack = true
	}
}

func newUC(db *fakePeriodDB, commissions *fakeCommissions, locks *fakeVisitLock, tx *fakeTx) *PeriodUC {
	if db == nil {
		db = &fakePeriodDB{}
	}
	if commissions == nil {
		commissions = &fakeCommissions{}
	}
	if locks == nil {
		locks = &fakeVisitLock{}
	}
	if tx == nil {
		tx = &fakeTx{}
	}
	return &PeriodUC{
		PeriodDB:    db,
		Commissions: commissions,
		VisitLockDB: locks,
		Transaction: tx,
		now: func() time.Time {
			return time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
		},
	}
}

func TestCreatePeriod(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := &fakePeriodDB{}
		uc := newUC(db, nil, nil, nil)
		got, err := uc.CreatePeriod(testCtx(), createReq("Aug 2026", "2026-08-01", "2026-08-31"))
		if err != nil {
			t.Fatalf("CreatePeriod: %v", err)
		}
		if got.UUID != "generated-uuid" || got.Status != model.CompensationPeriodStatusOpen {
			t.Fatalf("unexpected period: %+v", got)
		}
		if got.PeriodStart != "2026-08-01" || got.PeriodEnd != "2026-08-31" || got.TotalWage != 0 {
			t.Fatalf("unexpected mapped fields: %+v", got)
		}
	})

	t.Run("overlap", func(t *testing.T) {
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{{
			UUID:          "existing",
			InstitutionID: testInstitutionID,
			PeriodStart:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
			PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
			Status:        model.CompensationPeriodStatusOpen,
		}}}
		uc := newUC(db, nil, nil, nil)
		_, err := uc.CreatePeriod(testCtx(), createReq("Overlap", "2026-08-01", "2026-08-15"))
		if errorName(t, err) != errDateRangeOverlap {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errDateRangeOverlap)
		}
	})

	t.Run("end before start", func(t *testing.T) {
		uc := newUC(nil, nil, nil, nil)
		_, err := uc.CreatePeriod(testCtx(), createReq("Bad", "2026-08-31", "2026-08-01"))
		if errorName(t, err) != errInvalidPeriodDates {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errInvalidPeriodDates)
		}
	})

	t.Run("empty label", func(t *testing.T) {
		uc := newUC(nil, nil, nil, nil)
		_, err := uc.CreatePeriod(testCtx(), createReq("  ", "2026-08-01", "2026-08-31"))
		if errorName(t, err) != errInvalidLabel {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errInvalidLabel)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := newUC(nil, nil, nil, nil)
		_, err := uc.CreatePeriod(context.Background(), createReq("Aug", "2026-08-01", "2026-08-31"))
		if err == nil {
			t.Fatal("expected unauthorized")
		}
	})
}

func TestListPeriods_DefaultLimit(t *testing.T) {
	db := &fakePeriodDB{}
	uc := newUC(db, nil, nil, nil)
	_, err := uc.ListPeriods(testCtx(), model.ListCompensationPeriodsRequest{})
	if err != nil {
		t.Fatalf("ListPeriods: %v", err)
	}
	if db.lastList.Limit != defaultListLimit {
		t.Fatalf("Limit = %d, want %d", db.lastList.Limit, defaultListLimit)
	}
	if db.lastList.InstitutionID != testInstitutionID {
		t.Fatalf("InstitutionID = %d", db.lastList.InstitutionID)
	}
}

func TestListPeriods_ExplicitLimitAndInvalidStatus(t *testing.T) {
	db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{
		{UUID: "a", InstitutionID: testInstitutionID, Status: model.CompensationPeriodStatusOpen, Label: "A"},
		{UUID: "b", InstitutionID: testInstitutionID, Status: model.CompensationPeriodStatusOpen, Label: "B"},
		{UUID: "c", InstitutionID: testInstitutionID, Status: model.CompensationPeriodStatusOpen, Label: "C"},
	}}
	uc := newUC(db, nil, nil, nil)

	got, err := uc.ListPeriods(testCtx(), model.ListCompensationPeriodsRequest{
		CommonRequestPayload: model.CommonRequestPayload{Limit: 2, Offset: 1},
	})
	if err != nil {
		t.Fatalf("ListPeriods: %v", err)
	}
	if db.lastList.Limit != 2 || db.lastList.Offset != 1 {
		t.Fatalf("params = %+v", db.lastList)
	}
	if got.Total != 3 || len(got.Periods) != 2 {
		t.Fatalf("got total=%d len=%d", got.Total, len(got.Periods))
	}

	_, err = uc.ListPeriods(testCtx(), model.ListCompensationPeriodsRequest{Status: "closed"})
	if errorName(t, err) != errInvalidPeriodStatus {
		t.Fatalf("error name = %s, want %s", errorName(t, err), errInvalidPeriodStatus)
	}
}

func TestGetPeriod_NotFound(t *testing.T) {
	uc := newUC(nil, nil, nil, nil)
	_, err := uc.GetPeriod(testCtx(), "missing")
	if errorName(t, err) != errPeriodNotFound {
		t.Fatalf("error name = %s, want %s", errorName(t, err), errPeriodNotFound)
	}
}

func TestDraftPeriod(t *testing.T) {
	open := &model.TrxCompensationPeriod{
		ID:            1,
		UUID:          "p1",
		InstitutionID: testInstitutionID,
		Label:         "Aug 2026",
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		Status:        model.CompensationPeriodStatusOpen,
	}

	t.Run("open to draft", func(t *testing.T) {
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		commissions := &fakeCommissions{totals: map[int64]compensationrepo.PeriodCommissionTotals{
			1: {TotalCommission: 1000, StaffCount: 2, VisitCount: 4},
		}}
		uc := newUC(db, commissions, nil, nil)
		got, err := uc.DraftPeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("DraftPeriod: %v", err)
		}
		if got.Status != model.CompensationPeriodStatusDraft || got.TotalCommission != 1000 || got.TotalPayout != 1000 || got.TotalWage != 0 {
			t.Fatalf("unexpected draft: %+v", got)
		}
		if db.updateCalls != 1 {
			t.Fatalf("updateCalls = %d", db.updateCalls)
		}
	})

	t.Run("re-draft", func(t *testing.T) {
		p := copyPeriod(open)
		p.Status = model.CompensationPeriodStatusDraft
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		uc := newUC(db, &fakeCommissions{}, nil, nil)
		got, err := uc.DraftPeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("DraftPeriod: %v", err)
		}
		if got.Status != model.CompensationPeriodStatusDraft {
			t.Fatalf("status = %s", got.Status)
		}
	})

	t.Run("finalized not allowed", func(t *testing.T) {
		p := copyPeriod(open)
		p.Status = model.CompensationPeriodStatusFinalized
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		uc := newUC(db, nil, nil, nil)
		_, err := uc.DraftPeriod(testCtx(), "p1")
		if errorName(t, err) != errIllegalTransition {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errIllegalTransition)
		}
		if db.updateCalls != 0 {
			t.Fatalf("updateCalls = %d", db.updateCalls)
		}
	})

	t.Run("overlap on draft", func(t *testing.T) {
		p := copyPeriod(open)
		other := &model.TrxCompensationPeriod{
			UUID:          "p2",
			InstitutionID: testInstitutionID,
			PeriodStart:   time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC),
			PeriodEnd:     time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC),
			Status:        model.CompensationPeriodStatusOpen,
		}
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{p, other}}
		uc := newUC(db, nil, nil, nil)
		_, err := uc.DraftPeriod(testCtx(), "p1")
		if errorName(t, err) != errDateRangeOverlap {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errDateRangeOverlap)
		}
	})
}

func TestReopenPeriod(t *testing.T) {
	locks := &fakeVisitLock{locked: map[int64][]int64{1: {10, 11}}}
	db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{{
		ID:            1,
		UUID:          "p1",
		InstitutionID: testInstitutionID,
		Status:        model.CompensationPeriodStatusFinalized,
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		FinalizedAt:   sql.NullTime{Time: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC), Valid: true},
		FinalizedBy:   sql.NullString{String: testStaffUUID, Valid: true},
	}}}
	uc := newUC(db, nil, locks, nil)
	got, err := uc.ReopenPeriod(testCtx(), "p1")
	if err != nil {
		t.Fatalf("ReopenPeriod: %v", err)
	}
	if got.Status != model.CompensationPeriodStatusDraft {
		t.Fatalf("status = %s", got.Status)
	}
	if locks.calls != 0 {
		t.Fatalf("LockVisits calls = %d, want 0", locks.calls)
	}
	if len(locks.locked[1]) != 2 {
		t.Fatalf("lock state cleared: %+v", locks.locked)
	}
	if !got.FinalizedAt.Valid {
		t.Fatal("finalized_at should remain after reopen")
	}

	db.periods[0].Status = model.CompensationPeriodStatusOpen
	_, err = uc.ReopenPeriod(testCtx(), "p1")
	if errorName(t, err) != errIllegalTransition {
		t.Fatalf("error name = %s, want %s", errorName(t, err), errIllegalTransition)
	}
}

func TestDeletePeriod(t *testing.T) {
	open := &model.TrxCompensationPeriod{
		UUID:          "p1",
		InstitutionID: testInstitutionID,
		Status:        model.CompensationPeriodStatusOpen,
	}

	t.Run("open ok", func(t *testing.T) {
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		uc := newUC(db, nil, nil, nil)
		got, err := uc.DeletePeriod(testCtx(), "p1")
		if err != nil || !got.Success {
			t.Fatalf("DeletePeriod: %+v %v", got, err)
		}
		if db.softDeleteCalls != 1 {
			t.Fatalf("softDeleteCalls = %d", db.softDeleteCalls)
		}
	})

	for _, status := range []model.CompensationPeriodStatus{
		model.CompensationPeriodStatusDraft,
		model.CompensationPeriodStatusFinalized,
	} {
		status := status
		t.Run("illegal "+string(status), func(t *testing.T) {
			p := copyPeriod(open)
			p.Status = status
			db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{p}}
			uc := newUC(db, nil, nil, nil)
			_, err := uc.DeletePeriod(testCtx(), "p1")
			if errorName(t, err) != errIllegalTransition {
				t.Fatalf("error name = %s", errorName(t, err))
			}
			if db.softDeleteCalls != 0 {
				t.Fatalf("softDeleteCalls = %d", db.softDeleteCalls)
			}
		})
	}
}

func copyPeriod(p *model.TrxCompensationPeriod) *model.TrxCompensationPeriod {
	cp := *p
	return &cp
}
