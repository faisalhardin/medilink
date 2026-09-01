package compensation

import (
	"errors"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
)

func draftPeriod(id int64, uuid string) *model.TrxCompensationPeriod {
	return &model.TrxCompensationPeriod{
		ID:            id,
		UUID:          uuid,
		InstitutionID: testInstitutionID,
		Label:         "Aug 2026",
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		Status:        model.CompensationPeriodStatusDraft,
	}
}

func TestFinalizePeriod(t *testing.T) {
	t.Run("from open not allowed", func(t *testing.T) {
		p := draftPeriod(1, "p1")
		p.Status = model.CompensationPeriodStatusOpen
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		locks := &fakeVisitLock{}
		tx := &fakeTx{}
		uc := newUC(db, nil, locks, tx)
		_, err := uc.FinalizePeriod(testCtx(), "p1")
		if errorName(t, err) != errIllegalTransition {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errIllegalTransition)
		}
		if locks.calls != 0 || db.updateCalls != 0 || tx.began {
			t.Fatalf("unexpected side effects: locks=%d updates=%d began=%v", locks.calls, db.updateCalls, tx.began)
		}
	})

	t.Run("from draft locks commission visit ids", func(t *testing.T) {
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{draftPeriod(1, "p1")}}
		commissions := &fakeCommissions{
			totals: map[int64]compensationrepo.PeriodCommissionTotals{
				1: {TotalCommission: 2500, StaffCount: 2, VisitCount: 3},
			},
			visits: map[int64][]int64{1: {101, 202, 303}},
		}
		locks := &fakeVisitLock{}
		tx := &fakeTx{}
		uc := newUC(db, commissions, locks, tx)
		got, err := uc.FinalizePeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("FinalizePeriod: %v", err)
		}
		if got.Period.Status != model.CompensationPeriodStatusFinalized {
			t.Fatalf("status = %s", got.Period.Status)
		}
		if got.LockedVisitCount != 3 {
			t.Fatalf("locked_visit_count = %d", got.LockedVisitCount)
		}
		if got.Period.TotalCommission != 2500 || got.Period.TotalWage != 0 {
			t.Fatalf("unexpected totals: %+v", got.Period)
		}
		if locks.calls != 1 {
			t.Fatalf("LockVisits calls = %d", locks.calls)
		}
		if len(locks.locked[1]) != 3 || locks.locked[1][0] != 101 {
			t.Fatalf("locked ids = %+v", locks.locked[1])
		}
		if !tx.began || !tx.finished || tx.rolledBack {
			t.Fatalf("tx began=%v finished=%v rolledBack=%v", tx.began, tx.finished, tx.rolledBack)
		}
	})

	t.Run("aggregator error rolls back and does not lock", func(t *testing.T) {
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{draftPeriod(1, "p1")}}
		commissions := &fakeCommissions{sumErr: errors.New("sum failed")}
		locks := &fakeVisitLock{}
		tx := &fakeTx{}
		uc := newUC(db, commissions, locks, tx)
		_, err := uc.FinalizePeriod(testCtx(), "p1")
		if err == nil {
			t.Fatal("expected error")
		}
		if locks.calls != 0 {
			t.Fatalf("LockVisits calls = %d", locks.calls)
		}
		if db.updateCalls != 0 {
			t.Fatalf("updateCalls = %d", db.updateCalls)
		}
		if !tx.began || !tx.finished || !tx.rolledBack {
			t.Fatalf("tx began=%v finished=%v rolledBack=%v", tx.began, tx.finished, tx.rolledBack)
		}
	})

	t.Run("already finalized is no-op", func(t *testing.T) {
		p := draftPeriod(1, "p1")
		p.Status = model.CompensationPeriodStatusFinalized
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		commissions := &fakeCommissions{visits: map[int64][]int64{1: {9, 8}}}
		locks := &fakeVisitLock{}
		tx := &fakeTx{}
		uc := newUC(db, commissions, locks, tx)
		got, err := uc.FinalizePeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("FinalizePeriod: %v", err)
		}
		if got.LockedVisitCount != 2 {
			t.Fatalf("locked_visit_count = %d", got.LockedVisitCount)
		}
		if db.updateCalls != 0 || locks.calls != 0 || tx.began {
			t.Fatalf("no-op wrote state: updates=%d locks=%d began=%v", db.updateCalls, locks.calls, tx.began)
		}
	})

	t.Run("empty commission list", func(t *testing.T) {
		db := &fakePeriodDB{periods: []*model.TrxCompensationPeriod{draftPeriod(1, "p1")}}
		locks := &fakeVisitLock{}
		tx := &fakeTx{}
		uc := newUC(db, &fakeCommissions{}, locks, tx)
		got, err := uc.FinalizePeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("FinalizePeriod: %v", err)
		}
		if got.LockedVisitCount != 0 {
			t.Fatalf("locked_visit_count = %d", got.LockedVisitCount)
		}
		if locks.calls != 1 {
			t.Fatalf("LockVisits should still be called, calls=%d", locks.calls)
		}
	})
}
