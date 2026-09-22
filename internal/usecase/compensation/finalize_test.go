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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		tx := &fakeTx{}
		uc := newUC(db, nil, tx)
		_, err := uc.FinalizePeriod(testCtx(), "p1")
		if errorName(t, err) != errIllegalTransition {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errIllegalTransition)
		}
		if db.updateCalls != 0 || tx.began {
			t.Fatalf("unexpected side effects: updates=%d began=%v", db.updateCalls, tx.began)
		}
	})

	t.Run("from draft finalizes using worksheet rollup", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{draftPeriod(1, "p1")}}
		worksheets := &fakeWorksheetDB{periodTotals: map[int64]compensationrepo.WorksheetPeriodTotals{
			1: {TotalCommission: 2500, StaffCount: 2, VisitCount: 3},
		}}
		tx := &fakeTx{}
		uc := newUC(db, worksheets, tx)
		got, err := uc.FinalizePeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("FinalizePeriod: %v", err)
		}
		if got.Period.Status != model.CompensationPeriodStatusFinalized {
			t.Fatalf("status = %s", got.Period.Status)
		}
		// LockedVisitCount is the visit count from worksheet rollup (not from visit locking).
		if got.LockedVisitCount != 3 {
			t.Fatalf("locked_visit_count = %d, want 3 (from worksheet rollup)", got.LockedVisitCount)
		}
		if got.Period.TotalCommission != 2500 || got.Period.TotalWage != 0 {
			t.Fatalf("unexpected totals: %+v", got.Period)
		}
		if db.updateCalls != 1 {
			t.Fatalf("updateCalls = %d", db.updateCalls)
		}
		if !tx.began || !tx.finished || tx.rolledBack {
			t.Fatalf("tx began=%v finished=%v rolledBack=%v", tx.began, tx.finished, tx.rolledBack)
		}
	})

	t.Run("worksheet sum error rolls back", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{draftPeriod(1, "p1")}}
		worksheets := &fakeWorksheetDB{sumErr: errors.New("sum failed")}
		tx := &fakeTx{}
		uc := newUC(db, worksheets, tx)
		_, err := uc.FinalizePeriod(testCtx(), "p1")
		if err == nil {
			t.Fatal("expected error")
		}
		if db.updateCalls != 0 {
			t.Fatalf("updateCalls = %d", db.updateCalls)
		}
		if !tx.began || !tx.finished || !tx.rolledBack {
			t.Fatalf("tx began=%v finished=%v rolledBack=%v", tx.began, tx.finished, tx.rolledBack)
		}
	})

	t.Run("already finalized is no-op returns worksheet totals", func(t *testing.T) {
		p := draftPeriod(1, "p1")
		p.Status = model.CompensationPeriodStatusFinalized
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		worksheets := &fakeWorksheetDB{periodTotals: map[int64]compensationrepo.WorksheetPeriodTotals{
			1: {TotalCommission: 0, StaffCount: 0, VisitCount: 2},
		}}
		tx := &fakeTx{}
		uc := newUC(db, worksheets, tx)
		got, err := uc.FinalizePeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("FinalizePeriod: %v", err)
		}
		if got.LockedVisitCount != 2 {
			t.Fatalf("locked_visit_count = %d, want 2 (from worksheets)", got.LockedVisitCount)
		}
		if db.updateCalls != 0 || tx.began {
			t.Fatalf("no-op wrote state: updates=%d began=%v", db.updateCalls, tx.began)
		}
	})

	t.Run("empty worksheet list", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{draftPeriod(1, "p1")}}
		tx := &fakeTx{}
		uc := newUC(db, &fakeWorksheetDB{}, tx)
		got, err := uc.FinalizePeriod(testCtx(), "p1")
		if err != nil {
			t.Fatalf("FinalizePeriod: %v", err)
		}
		if got.LockedVisitCount != 0 {
			t.Fatalf("locked_visit_count = %d, want 0 for empty worksheets", got.LockedVisitCount)
		}
		if db.updateCalls != 1 {
			t.Fatalf("updateCalls = %d, want 1", db.updateCalls)
		}
	})
}
