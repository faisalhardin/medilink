package compensation

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/volatiletech/null/v8"
)

type fakeSnapshotDB struct {
	rows      []model.TrxWagePeriodSnapshot
	nextID    int64
	listErr   error
	insertErr error
	getErr    error
	updateErr error
	deleteErr error
}

func (f *fakeSnapshotDB) List(_ context.Context, institutionID int64) ([]model.TrxWagePeriodSnapshot, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]model.TrxWagePeriodSnapshot, 0)
	for _, row := range f.rows {
		if row.DeleteTime != nil || row.InstitutionID != institutionID {
			continue
		}
		out = append(out, row)
	}
	return out, nil
}

func (f *fakeSnapshotDB) Insert(_ context.Context, row *model.TrxWagePeriodSnapshot) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.nextID++
	row.ID = f.nextID
	f.rows = append(f.rows, *row)
	return nil
}

func (f *fakeSnapshotDB) Get(_ context.Context, institutionID, id int64) (*model.TrxWagePeriodSnapshot, bool, error) {
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	for i := range f.rows {
		row := f.rows[i]
		if row.ID == id && row.InstitutionID == institutionID && row.DeleteTime == nil {
			return &row, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakeSnapshotDB) UpdateWageInputs(_ context.Context, row *model.TrxWagePeriodSnapshot) (bool, error) {
	if f.updateErr != nil {
		return false, f.updateErr
	}
	if row == nil {
		return false, nil
	}
	for i := range f.rows {
		if f.rows[i].ID != row.ID || f.rows[i].InstitutionID != row.InstitutionID || f.rows[i].DeleteTime != nil {
			continue
		}
		f.rows[i].MandatoryWorkingDays = row.MandatoryWorkingDays
		f.rows[i].StaffWorkingDays = row.StaffWorkingDays
		f.rows[i].FinalWage = row.FinalWage
		f.rows[i].TotalWage = row.TotalWage
		return true, nil
	}
	return false, nil
}

func (f *fakeSnapshotDB) SoftDelete(_ context.Context, institutionID, id int64) (bool, error) {
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	for i := range f.rows {
		if f.rows[i].ID != id || f.rows[i].InstitutionID != institutionID || f.rows[i].DeleteTime != nil {
			continue
		}
		now := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
		f.rows[i].DeleteTime = &now
		return true, nil
	}
	return false, nil
}

func newSnapshotUC(wages *fakeWageDB, snaps *fakeSnapshotDB, tx *fakeTx) *WageSnapshotUC {
	if wages == nil {
		wages = &fakeWageDB{}
	}
	if snaps == nil {
		snaps = &fakeSnapshotDB{}
	}
	if tx == nil {
		tx = &fakeTx{}
	}
	return NewWageSnapshotUC(&WageSnapshotUC{
		WageDB:      wages,
		SnapshotDB:  snaps,
		PeriodDB:    augustPeriodDB(),
		Transaction: tx,
		now: func() time.Time {
			return time.Date(2026, 8, 8, 2, 0, 0, 0, time.UTC)
		},
	})
}

func augustPeriodDB() *fakeCompensationPeriodDB {
	return &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{{
		ID:            7,
		UUID:          "period-aug",
		InstitutionID: testInstitutionID,
		PeriodStart:   parseDate("2026-08-01").Time(),
		PeriodEnd:     parseDate("2026-08-31").Time(),
	}}}
}

func generateReq(staffIDs []string) model.GenerateStaffWageSnapshotsRequest {
	return model.GenerateStaffWageSnapshotsRequest{
		CompensationPeriodUUID: "period-aug",
		StaffIDs:               staffIDs,
	}
}

func TestGenerate_AllActiveWages(t *testing.T) {
	wages := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", true, "2026-01-01", ""),
		wageOn(2, "staff-b", false, "2026-01-01", "2026-07-31"),
		wageOn(3, "staff-c", true, "2026-02-01", ""),
	}}
	wages.rows[2].WageAmount = 9000
	wages.rows[2].WageCadence = model.WageCadenceWeekly
	other := wageOn(4, "staff-d", true, "2026-01-01", "")
	other.InstitutionID = testInstitutionID + 1
	wages.rows = append(wages.rows, other)

	snaps := &fakeSnapshotDB{}
	tx := &fakeTx{}
	uc := newSnapshotUC(wages, snaps, tx)

	got, err := uc.Generate(testCtx(), generateReq(nil))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !tx.began || !tx.finished || tx.rolledBack {
		t.Fatalf("transaction began=%v finished=%v rolledBack=%v", tx.began, tx.finished, tx.rolledBack)
	}
	if len(got.Snapshots) != 2 || len(snaps.rows) != 2 {
		t.Fatalf("snapshots=%d rows=%d", len(got.Snapshots), len(snaps.rows))
	}
	if got.Snapshots[0].StaffID != "staff-a" || got.Snapshots[0].WageAmount != 1000 || got.Snapshots[0].WageCadence != model.WageCadenceMonthly {
		t.Fatalf("first: %+v", got.Snapshots[0])
	}
	if got.Snapshots[1].StaffID != "staff-c" || got.Snapshots[1].WageAmount != 9000 || got.Snapshots[1].WageCadence != model.WageCadenceWeekly {
		t.Fatalf("second: %+v", got.Snapshots[1])
	}
	if got.Snapshots[0].PeriodStart != "2026-08-01" || got.Snapshots[0].PeriodEnd != "2026-08-31" {
		t.Fatalf("dates: %+v", got.Snapshots[0])
	}
	if got.Snapshots[0].CompensationPeriodUUID != "period-aug" || snaps.rows[0].CompensationPeriodID != 7 {
		t.Fatalf("period: response=%+v row=%+v", got.Snapshots[0], snaps.rows[0])
	}
	if got.Snapshots[0].CreatedAt != "2026-08-08T02:00:00Z" {
		t.Fatalf("created_at = %s", got.Snapshots[0].CreatedAt)
	}
	if got.Snapshots[0].MandatoryWorkingDays.Valid || got.Snapshots[0].StaffWorkingDays.Valid || got.Snapshots[0].FinalWage.Valid || got.Snapshots[0].TotalWage.Valid {
		t.Fatalf("generated amounts should be empty: %+v", got.Snapshots[0])
	}
}

func TestGenerate_StaffSubset(t *testing.T) {
	wages := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", true, "2026-01-01", ""),
		wageOn(2, "staff-c", true, "2026-01-01", ""),
	}}
	snaps := &fakeSnapshotDB{}
	uc := newSnapshotUC(wages, snaps, nil)

	got, err := uc.Generate(testCtx(), generateReq([]string{"staff-c", "staff-c"}))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(got.Snapshots) != 1 || got.Snapshots[0].StaffID != "staff-c" || len(snaps.rows) != 1 {
		t.Fatalf("got %+v rows %d", got.Snapshots, len(snaps.rows))
	}
}

func TestGenerate_MissingContractInsertsNothing(t *testing.T) {
	wages := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", true, "2026-01-01", ""),
	}}
	snaps := &fakeSnapshotDB{}
	tx := &fakeTx{}
	uc := newSnapshotUC(wages, snaps, tx)

	_, err := uc.Generate(testCtx(), generateReq([]string{"staff-a", "staff-missing"}))
	if errorName(t, err) != errWageContractNotFound {
		t.Fatalf("error = %v", err)
	}
	if len(snaps.rows) != 0 {
		t.Fatalf("inserted %d rows", len(snaps.rows))
	}
	if tx.began {
		t.Fatal("transaction began")
	}
}

func TestGenerate_UsesWageCoveringPeriod(t *testing.T) {
	closed := wageOn(1, "staff-a", false, "2026-01-01", "2026-08-31")
	closed.WageAmount = 3000000
	later := wageOn(2, "staff-a", true, "2026-09-01", "")
	later.WageAmount = 3100000
	later.WageCadence = model.WageCadenceDaily
	older := wageOn(3, "staff-b", true, "2026-01-01", "2026-08-31")
	newer := wageOn(4, "staff-b", false, "2026-06-01", "")
	newer.WageAmount = 2000
	newer.WageCadence = model.WageCadenceWeekly
	outside := wageOn(5, "staff-c", true, "2026-09-01", "")

	wages := &fakeWageDB{rows: []model.MstStaffWage{closed, later, older, newer, outside}}
	snaps := &fakeSnapshotDB{}
	uc := newSnapshotUC(wages, snaps, nil)

	got, err := uc.Generate(testCtx(), generateReq(nil))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(got.Snapshots) != 2 {
		t.Fatalf("snapshots=%d", len(got.Snapshots))
	}
	if got.Snapshots[0].StaffID != "staff-a" || got.Snapshots[0].WageAmount != 3000000 || got.Snapshots[0].WageCadence != model.WageCadenceMonthly {
		t.Fatalf("closed contract: %+v", got.Snapshots[0])
	}
	if got.Snapshots[1].StaffID != "staff-b" || got.Snapshots[1].WageAmount != 2000 || got.Snapshots[1].WageCadence != model.WageCadenceWeekly {
		t.Fatalf("later contract: %+v", got.Snapshots[1])
	}
}

func TestGenerate_ContractOutsidePeriod(t *testing.T) {
	wages := &fakeWageDB{rows: []model.MstStaffWage{
		wageOn(1, "staff-a", true, "2026-09-01", ""),
	}}
	snaps := &fakeSnapshotDB{}
	tx := &fakeTx{}
	uc := newSnapshotUC(wages, snaps, tx)

	_, err := uc.Generate(testCtx(), generateReq([]string{"staff-a"}))
	if errorName(t, err) != errWageContractNotFound {
		t.Fatalf("error = %v", err)
	}
	if len(snaps.rows) != 0 || tx.began {
		t.Fatalf("rows=%d began=%v", len(snaps.rows), tx.began)
	}
}

func TestGenerate_NoActiveWages(t *testing.T) {
	snaps := &fakeSnapshotDB{}
	tx := &fakeTx{}
	uc := newSnapshotUC(&fakeWageDB{}, snaps, tx)

	got, err := uc.Generate(testCtx(), generateReq(nil))
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(got.Snapshots) != 0 || len(snaps.rows) != 0 || tx.began {
		t.Fatalf("snapshots=%d rows=%d began=%v", len(got.Snapshots), len(snaps.rows), tx.began)
	}
}

func TestGenerate_PeriodRequired(t *testing.T) {
	uc := newSnapshotUC(nil, nil, nil)
	_, err := uc.Generate(testCtx(), model.GenerateStaffWageSnapshotsRequest{})
	if errorName(t, err) != errSnapshotPeriodRequired {
		t.Fatalf("error = %v", err)
	}
}

func TestGenerate_PeriodNotFound(t *testing.T) {
	tx := &fakeTx{}
	uc := newSnapshotUC(nil, nil, tx)
	_, err := uc.Generate(testCtx(), model.GenerateStaffWageSnapshotsRequest{CompensationPeriodUUID: "missing"})
	if errorName(t, err) != errPaydayPeriodNotFound {
		t.Fatalf("error = %v", err)
	}
	if tx.began {
		t.Fatal("transaction began")
	}
}

func TestList_InstitutionScope(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{
		{ID: 1, StaffID: "staff-a", InstitutionID: testInstitutionID, PeriodStart: parseDate("2026-08-01").Time(), PeriodEnd: parseDate("2026-08-31").Time(), WageAmount: 10, WageCadence: model.WageCadenceDaily, CreateTime: time.Date(2026, 8, 8, 2, 0, 0, 0, time.UTC)},
		{ID: 2, StaffID: "staff-b", InstitutionID: testInstitutionID + 1, WageCadence: model.WageCadenceMonthly},
	}}
	uc := newSnapshotUC(nil, snaps, nil)
	got, err := uc.List(testCtx())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got.Snapshots) != 1 || got.Snapshots[0].ID != 1 || got.Snapshots[0].StaffID != "staff-a" {
		t.Fatalf("got %+v", got.Snapshots)
	}
}

func TestDelete_NotFound(t *testing.T) {
	uc := newSnapshotUC(nil, &fakeSnapshotDB{}, nil)
	_, err := uc.Delete(testCtx(), 9)
	if errorName(t, err) != errSnapshotNotFound {
		t.Fatalf("error = %v", err)
	}
}

func TestDelete_OtherInstitution(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{
		{ID: 3, InstitutionID: testInstitutionID + 1, StaffID: "staff-a"},
	}}
	uc := newSnapshotUC(nil, snaps, nil)
	_, err := uc.Delete(testCtx(), 3)
	if errorName(t, err) != errSnapshotNotFound {
		t.Fatalf("error = %v", err)
	}
	if snaps.rows[0].DeleteTime != nil {
		t.Fatal("deleted another institution")
	}
}

func snapshotDays(n int64) null.Int64 {
	return null.Int64From(n)
}

func TestUpdate_ProrateWinsOverFinal(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID:                     4,
		StaffID:                "staff-a",
		InstitutionID:          testInstitutionID,
		CompensationPeriodID:   7,
		CompensationPeriodUUID: sql.NullString{String: "period-aug", Valid: true},
		WageAmount:             1000,
		WageCadence:            model.WageCadenceMonthly,
	}}}
	uc := newSnapshotUC(nil, snaps, nil)

	got, err := uc.Update(testCtx(), 4, model.UpdateStaffWageSnapshotRequest{
		MandatoryWorkingDays: snapshotDays(22),
		StaffWorkingDays:     snapshotDays(20),
		FinalWage:            snapshotDays(500),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !got.TotalWage.Valid || got.TotalWage.Int64 != 909 {
		t.Fatalf("total = %+v, want 909", got.TotalWage)
	}
	if !got.FinalWage.Valid || got.FinalWage.Int64 != 500 {
		t.Fatalf("final = %+v", got.FinalWage)
	}
	if snaps.rows[0].TotalWage.Int64 != 909 || snaps.rows[0].FinalWage.Int64 != 500 {
		t.Fatalf("stored %+v", snaps.rows[0])
	}
}

func TestUpdate_FinalWhenDaysMissing(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID:            4,
		InstitutionID: testInstitutionID,
		WageAmount:    1000,
		WageCadence:   model.WageCadenceMonthly,
	}}}
	uc := newSnapshotUC(nil, snaps, nil)

	got, err := uc.Update(testCtx(), 4, model.UpdateStaffWageSnapshotRequest{
		FinalWage: snapshotDays(750),
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.MandatoryWorkingDays.Valid || got.StaffWorkingDays.Valid {
		t.Fatalf("days should stay empty: %+v", got)
	}
	if !got.TotalWage.Valid || got.TotalWage.Int64 != 750 {
		t.Fatalf("total = %+v, want 750", got.TotalWage)
	}
}

func TestUpdate_ClearsToEmpty(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID:                   4,
		InstitutionID:        testInstitutionID,
		WageAmount:           1000,
		MandatoryWorkingDays: sql.NullInt64{Int64: 22, Valid: true},
		StaffWorkingDays:     sql.NullInt64{Int64: 20, Valid: true},
		TotalWage:            sql.NullInt64{Int64: 909, Valid: true},
	}}}
	uc := newSnapshotUC(nil, snaps, nil)

	got, err := uc.Update(testCtx(), 4, model.UpdateStaffWageSnapshotRequest{})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.TotalWage.Valid || snaps.rows[0].TotalWage.Valid || snaps.rows[0].MandatoryWorkingDays.Valid {
		t.Fatalf("expected cleared totals, got %+v stored %+v", got, snaps.rows[0])
	}
}

func TestUpdate_InvalidDays(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID:            4,
		InstitutionID: testInstitutionID,
		WageAmount:    1000,
	}}}
	uc := newSnapshotUC(nil, snaps, nil)

	_, err := uc.Update(testCtx(), 4, model.UpdateStaffWageSnapshotRequest{
		MandatoryWorkingDays: snapshotDays(0),
		StaffWorkingDays:     snapshotDays(20),
	})
	if errorName(t, err) != model.SnapshotDaysInvalidCode {
		t.Fatalf("error = %v", err)
	}
	if snaps.rows[0].TotalWage.Valid {
		t.Fatal("stored a total after a rejected update")
	}
}

func TestUpdate_NegativeFinal(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID:            4,
		InstitutionID: testInstitutionID,
		WageAmount:    1000,
	}}}
	uc := newSnapshotUC(nil, snaps, nil)

	_, err := uc.Update(testCtx(), 4, model.UpdateStaffWageSnapshotRequest{
		FinalWage: snapshotDays(-1),
	})
	if errorName(t, err) != model.SnapshotFinalInvalidCode {
		t.Fatalf("error = %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	uc := newSnapshotUC(nil, &fakeSnapshotDB{}, nil)
	_, err := uc.Update(testCtx(), 9, model.UpdateStaffWageSnapshotRequest{FinalWage: snapshotDays(1)})
	if errorName(t, err) != errSnapshotNotFound {
		t.Fatalf("error = %v", err)
	}
}
