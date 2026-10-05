package compensation

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/pkg/errors"
)

type fakeWageDB struct {
	rows       []model.MstStaffWage
	nextID     int64
	listErr    error
	liveErr    error
	closeErr   error
	insertErr  error
	deleteErr  error
	closed     []closedWage
	inserted   []*model.MstStaffWage
	deletedIDs []int64
}

type closedWage struct {
	ID          int64
	EffectiveTo time.Time
	UpdatedBy   string
}

func (f *fakeWageDB) ListActive(_ context.Context, institutionID int64, staffID string) ([]model.MstStaffWage, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]model.MstStaffWage, 0)
	for _, row := range f.rows {
		if row.DeleteTime != nil || !row.IsActive || row.InstitutionID != institutionID {
			continue
		}
		if staffID != "" && row.StaffID != staffID {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func (f *fakeWageDB) ListLiveByStaff(_ context.Context, institutionID int64, staffID string) ([]model.MstStaffWage, error) {
	if f.liveErr != nil {
		return nil, f.liveErr
	}
	out := make([]model.MstStaffWage, 0)
	for _, row := range f.rows {
		if row.DeleteTime != nil || row.InstitutionID != institutionID || row.StaffID != staffID {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func (f *fakeWageDB) Close(_ context.Context, id, institutionID int64, effectiveTo time.Time, updatedBy string) error {
	if f.closeErr != nil {
		return f.closeErr
	}
	for i := range f.rows {
		if f.rows[i].ID != id || f.rows[i].InstitutionID != institutionID || f.rows[i].DeleteTime != nil {
			continue
		}
		f.rows[i].IsActive = false
		f.rows[i].EffectiveTo = sql.NullTime{Time: effectiveTo, Valid: true}
		f.rows[i].UpdatedBy = sql.NullString{String: updatedBy, Valid: updatedBy != ""}
		f.closed = append(f.closed, closedWage{ID: id, EffectiveTo: effectiveTo, UpdatedBy: updatedBy})
		return nil
	}
	return errors.New("wage to close was not found")
}

func (f *fakeWageDB) Insert(_ context.Context, w *model.MstStaffWage) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.nextID++
	w.ID = f.nextID
	f.rows = append(f.rows, *w)
	copied := *w
	f.inserted = append(f.inserted, &copied)
	return nil
}

func (f *fakeWageDB) SoftDelete(_ context.Context, institutionID, id int64) (bool, error) {
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	for i := range f.rows {
		if f.rows[i].ID != id || f.rows[i].InstitutionID != institutionID || f.rows[i].DeleteTime != nil {
			continue
		}
		now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
		f.rows[i].DeleteTime = &now
		f.deletedIDs = append(f.deletedIDs, id)
		return true, nil
	}
	return false, nil
}

func newWageUC(db *fakeWageDB, tx *fakeTx) *WageUC {
	if db == nil {
		db = &fakeWageDB{}
	}
	if tx == nil {
		tx = &fakeTx{}
	}
	return NewWageUC(&WageUC{WageDB: db, Transaction: tx})
}

func wageOn(id int64, staff string, active bool, from, to string) model.MstStaffWage {
	row := model.MstStaffWage{
		ID:            id,
		StaffID:       staff,
		InstitutionID: testInstitutionID,
		WageAmount:    1000,
		WageCadence:   model.WageCadenceMonthly,
		IsActive:      active,
		EffectiveFrom: parseDate(from).Time(),
	}
	if to != "" {
		row.EffectiveTo = sql.NullTime{Time: parseDate(to).Time(), Valid: true}
	}
	return row
}

func upsertReq(staff, cadence, from, to string, amount int64) model.UpsertStaffWageRequest {
	req := model.UpsertStaffWageRequest{
		StaffID:       staff,
		WageAmount:    amount,
		WageCadence:   model.WageCadence(cadence),
		EffectiveFrom: parseDate(from),
	}
	if to != "" {
		end := parseDate(to)
		req.EffectiveTo = &end
	}
	return req
}

func TestUpsertWage_InsertsWhenNoneExist(t *testing.T) {
	db := &fakeWageDB{}
	tx := &fakeTx{}
	uc := newWageUC(db, tx)

	got, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "daily", "2026-03-01", "", 250000))
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if !tx.began || !tx.finished || tx.rolledBack {
		t.Fatalf("transaction began=%v finished=%v rolledBack=%v", tx.began, tx.finished, tx.rolledBack)
	}
	if got.Wage.ID == 0 || !got.Wage.IsActive || got.Wage.WageCadence != model.WageCadenceDaily {
		t.Fatalf("response: %+v", got.Wage)
	}
	if got.Wage.EffectiveFrom != "2026-03-01" || got.Wage.EffectiveTo.Valid {
		t.Fatalf("dates: %+v", got.Wage)
	}
	if len(db.closed) != 0 || len(db.inserted) != 1 {
		t.Fatalf("closed=%d inserted=%d", len(db.closed), len(db.inserted))
	}
}

func TestUpsertWage_ClosesPreviousActive(t *testing.T) {
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(4, "staff-a", true, "2026-01-01", ""),
	}, nextID: 4}
	uc := newWageUC(db, nil)

	got, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "weekly", "2026-03-01", "2026-03-31", 80000))
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if len(db.closed) != 1 || db.closed[0].ID != 4 {
		t.Fatalf("closed: %+v", db.closed)
	}
	if !db.closed[0].EffectiveTo.Equal(time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("closed effective_to = %s", db.closed[0].EffectiveTo)
	}
	if db.closed[0].UpdatedBy != testStaffUUID {
		t.Fatalf("updated_by = %s", db.closed[0].UpdatedBy)
	}
	if db.rows[0].IsActive {
		t.Fatal("previous row still active")
	}
	if got.Wage.WageCadence != model.WageCadenceWeekly || got.Wage.EffectiveTo.String != "2026-03-31" || !got.Wage.IsActive {
		t.Fatalf("new wage: %+v", got.Wage)
	}
}

func TestUpsertWage_RejectsOverlapWithOtherLiveWage(t *testing.T) {
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", false, "2026-06-01", "2026-06-30"),
		wageOn(2, "staff-a", true, "2026-01-01", ""),
	}, nextID: 2}
	tx := &fakeTx{}
	uc := newWageUC(db, tx)

	_, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "monthly", "2026-06-15", "", 100))
	if errorName(t, err) != errWageOverlap {
		t.Fatalf("error = %v", err)
	}
	if !tx.rolledBack || len(db.inserted) != 0 || len(db.closed) != 0 {
		t.Fatalf("rolledBack=%v inserted=%d closed=%d", tx.rolledBack, len(db.inserted), len(db.closed))
	}
	if !db.rows[1].IsActive {
		t.Fatal("active row was closed despite rejection")
	}
}

func TestUpsertWage_RejectsCloseThatOverlapsHistory(t *testing.T) {
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", false, "2026-02-01", "2026-02-28"),
		wageOn(2, "staff-a", true, "2026-01-01", ""),
	}, nextID: 2}
	uc := newWageUC(db, nil)

	// Closing the January row through 28 Feb overlaps the February history row.
	_, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "monthly", "2026-03-01", "", 100))
	if errorName(t, err) != errWageOverlap {
		t.Fatalf("error = %v", err)
	}
	if len(db.inserted) != 0 || !db.rows[1].IsActive {
		t.Fatalf("inserted=%d active=%v", len(db.inserted), db.rows[1].IsActive)
	}
}

func TestUpsertWage_RejectsStartOnOrBeforeActive(t *testing.T) {
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(2, "staff-a", true, "2026-03-01", ""),
	}}
	uc := newWageUC(db, nil)

	_, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "monthly", "2026-03-01", "", 100))
	if errorName(t, err) != errWageOverlap {
		t.Fatalf("error = %v", err)
	}
}

func TestUpsertWage_RejectsMultipleActive(t *testing.T) {
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", true, "2026-01-01", "2026-01-31"),
		wageOn(2, "staff-a", true, "2026-03-01", ""),
	}}
	uc := newWageUC(db, nil)

	_, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "daily", "2026-04-01", "", 10))
	if errorName(t, err) != errWageMultipleActive {
		t.Fatalf("error = %v", err)
	}
}

func TestUpsertWage_RejectsInvalidCadenceAndDates(t *testing.T) {
	uc := newWageUC(nil, nil)
	ctx := testCtx()

	_, err := uc.UpsertWage(ctx, upsertReq("staff-a", "hourly", "2026-03-01", "", 10))
	if errorName(t, err) != errInvalidWageCadence {
		t.Fatalf("cadence error = %v", err)
	}

	_, err = uc.UpsertWage(ctx, upsertReq("", "monthly", "2026-03-01", "", 10))
	if errorName(t, err) != errWageRangeInvalid {
		t.Fatalf("staff error = %v", err)
	}

	req := upsertReq("staff-a", "monthly", "2026-03-10", "2026-03-01", 10)
	_, err = uc.UpsertWage(ctx, req)
	if errorName(t, err) != errWageRangeInvalid {
		t.Fatalf("order error = %v", err)
	}
}

func TestUpsertWage_AcceptsEachCadence(t *testing.T) {
	for _, cadence := range []string{"monthly", "weekly", "daily"} {
		t.Run(cadence, func(t *testing.T) {
			uc := newWageUC(nil, nil)
			got, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", cadence, "2026-04-01", "", 1))
			if err != nil {
				t.Fatalf("upsert: %v", err)
			}
			if string(got.Wage.WageCadence) != cadence {
				t.Fatalf("cadence = %s", got.Wage.WageCadence)
			}
		})
	}
}

func TestUpsertWage_IgnoresSoftDeletedOverlap(t *testing.T) {
	deletedAt := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	row := wageOn(1, "staff-a", true, "2026-03-01", "")
	row.DeleteTime = &deletedAt
	db := &fakeWageDB{rows: []model.MstStaffWage{row}, nextID: 1}
	uc := newWageUC(db, nil)

	got, err := uc.UpsertWage(testCtx(), upsertReq("staff-a", "monthly", "2026-03-01", "", 50))
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if !got.Wage.IsActive || len(db.closed) != 0 {
		t.Fatalf("closed a deleted row: %+v closed=%d", got.Wage, len(db.closed))
	}
}

func TestListWages_ActiveOnlyAndStaffFilter(t *testing.T) {
	deletedAt := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	gone := wageOn(3, "staff-a", true, "2026-01-01", "")
	gone.DeleteTime = &deletedAt
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", true, "2026-01-01", ""),
		wageOn(2, "staff-b", true, "2026-02-01", ""),
		wageOn(4, "staff-a", false, "2025-01-01", "2025-12-31"),
		gone,
	}}
	uc := newWageUC(db, nil)

	all, err := uc.ListWages(testCtx(), model.ListStaffWagesRequest{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all.Wages) != 2 {
		t.Fatalf("active count = %d", len(all.Wages))
	}

	filtered, err := uc.ListWages(testCtx(), model.ListStaffWagesRequest{StaffID: "staff-b"})
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	if len(filtered.Wages) != 1 || filtered.Wages[0].StaffID != "staff-b" {
		t.Fatalf("filtered: %+v", filtered.Wages)
	}
}

func TestDeleteWage_SoftDeleteAndNotFound(t *testing.T) {
	db := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(7, "staff-a", true, "2026-01-01", ""),
	}}
	uc := newWageUC(db, nil)
	ctx := testCtx()

	got, err := uc.DeleteWage(ctx, 7)
	if err != nil || !got.Success {
		t.Fatalf("delete: %+v %v", got, err)
	}
	listed, err := uc.ListWages(ctx, model.ListStaffWagesRequest{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed.Wages) != 0 {
		t.Fatalf("deleted wage still listed: %+v", listed.Wages)
	}

	_, err = uc.UpsertWage(ctx, upsertReq("staff-a", "daily", "2026-01-01", "", 10))
	if err != nil {
		t.Fatalf("reinsert after delete: %v", err)
	}

	_, err = uc.DeleteWage(ctx, 7)
	if errorName(t, err) != errWageNotFound {
		t.Fatalf("second delete = %v", err)
	}

	other := wageOn(8, "staff-a", true, "2026-05-01", "")
	other.InstitutionID = testInstitutionID + 1
	db.rows = append(db.rows, other)
	_, err = uc.DeleteWage(ctx, 8)
	if errorName(t, err) != errWageNotFound {
		t.Fatalf("other institution = %v", err)
	}
}
