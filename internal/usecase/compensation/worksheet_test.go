package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	"github.com/volatiletech/null/v8"
)

// patchFakeWorksheetDB stores worksheets for PatchWorksheet / GetWorksheet tests.
type patchFakeWorksheetDB struct {
	fakeWorksheetDB
	byUUID  map[string]*model.TrxWorksheet
	update  *model.TrxWorksheet
	created *model.TrxWorksheet
}

func (f *patchFakeWorksheetDB) GetByUUID(_ context.Context, institutionID int64, uuid string) (*model.TrxWorksheet, bool, error) {
	if f.byUUID == nil {
		return nil, false, nil
	}
	w, ok := f.byUUID[uuid]
	if !ok || w.InstitutionID != institutionID {
		return nil, false, nil
	}
	cp := *w
	return &cp, true, nil
}

func (f *patchFakeWorksheetDB) GetByID(_ context.Context, id int64) (*model.TrxWorksheet, bool, error) {
	for _, w := range f.byUUID {
		if w != nil && w.ID == id {
			cp := *w
			return &cp, true, nil
		}
	}
	return nil, false, nil
}

func (f *patchFakeWorksheetDB) SoftDelete(context.Context, int64, string) (bool, error) {
	return true, nil
}

func (f *patchFakeWorksheetDB) Create(_ context.Context, w *model.TrxWorksheet) error {
	if w == nil {
		return nil
	}
	cp := *w
	f.created = &cp
	return nil
}

func (f *patchFakeWorksheetDB) Update(_ context.Context, w *model.TrxWorksheet) error {
	cp := *w
	f.update = &cp
	if f.byUUID != nil {
		f.byUUID[w.UUID] = &cp
	}
	return nil
}

func openWorksheet(uuid string) *model.TrxWorksheet {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	return &model.TrxWorksheet{
		ID:             10,
		UUID:           uuid,
		InstitutionID:  testInstitutionID,
		StaffID:        testStaffUUID,
		Label:          "Aug WS",
		PeriodStart:    start,
		PeriodEnd:      end,
		Status:         model.WorksheetStatusOpen,
		GenerateStatus: model.WorksheetGenerateStatusIdle,
	}
}

func newWorksheetUC(ws *patchFakeWorksheetDB, periods *fakeCompensationPeriodDB) *WorksheetUC {
	if ws == nil {
		ws = &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{}}
	}
	if periods == nil {
		periods = &fakeCompensationPeriodDB{}
	}
	return &WorksheetUC{
		WorksheetDB:          ws,
		CompensationPeriodDB: periods,
	}
}

func augustPayday() *model.TrxCompensationPeriod {
	return &model.TrxCompensationPeriod{
		ID:            7,
		UUID:          "period-uuid-7",
		InstitutionID: testInstitutionID,
		Status:        model.CompensationPeriodStatusOpen,
		Label:         "Aug 2026",
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
	}
}

func TestCreateWorksheet_CopiesPayday(t *testing.T) {
	wsDB := &patchFakeWorksheetDB{}
	periodDB := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{augustPayday()}}
	uc := newWorksheetUC(wsDB, periodDB)

	got, err := uc.CreateWorksheet(testCtx(), model.CreateWorksheetRequest{
		StaffID:                testStaffUUID,
		Label:                  "Aug WS",
		CompensationPeriodUUID: "period-uuid-7",
		PeriodStart:            model.Time(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		PeriodEnd:              model.Time(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("CreateWorksheet: %v", err)
	}
	if got.PeriodStart != "2026-08-01" || got.PeriodEnd != "2026-08-31" || got.Status != model.WorksheetStatusOpen {
		t.Fatalf("got %+v", got)
	}
	if !got.CompensationPeriodUUID.Valid || got.CompensationPeriodUUID.String != "period-uuid-7" {
		t.Fatalf("period uuid = %+v", got.CompensationPeriodUUID)
	}
	if wsDB.created == nil || !wsDB.created.CompensationPeriodID.Valid || wsDB.created.CompensationPeriodID.Int64 != 7 {
		t.Fatalf("stored %+v", wsDB.created)
	}
	raw, _ := json.Marshal(got)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if _, has := m["compensation_period_id"]; has {
		t.Fatalf("response must not include compensation_period_id: %s", raw)
	}
}

func TestCreateWorksheet_RequiresPayday(t *testing.T) {
	uc := newWorksheetUC(nil, nil)
	_, err := uc.CreateWorksheet(testCtx(), model.CreateWorksheetRequest{StaffID: testStaffUUID, Label: "Aug"})
	if errorName(t, err) != errWorksheetPeriodRequired {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateWorksheet_FinalizedPaydayRejected(t *testing.T) {
	period := augustPayday()
	period.Status = model.CompensationPeriodStatusFinalized
	uc := newWorksheetUC(nil, &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{period}})
	_, err := uc.CreateWorksheet(testCtx(), model.CreateWorksheetRequest{
		StaffID: testStaffUUID, Label: "Aug", CompensationPeriodUUID: period.UUID,
	})
	if errorName(t, err) != errPaydayPeriodFinalized {
		t.Fatalf("error = %v", err)
	}
}

func TestCreateWorksheet_SecondWorksheetRejected(t *testing.T) {
	wsDB := &patchFakeWorksheetDB{}
	wsDB.staffOnPeriod = true
	uc := newWorksheetUC(wsDB, &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{augustPayday()}})
	_, err := uc.CreateWorksheet(testCtx(), model.CreateWorksheetRequest{
		StaffID: testStaffUUID, Label: "Aug", CompensationPeriodUUID: "period-uuid-7",
	})
	if errorName(t, err) != errWorksheetAlreadyOnPayday {
		t.Fatalf("error = %v", err)
	}
}

func TestGetWorksheet_ResolvesPeriodUUID(t *testing.T) {
	wsUUID := "ws-5"
	w := openWorksheet(wsUUID)
	w.CompensationPeriodID = sql.NullInt64{Int64: 7, Valid: true}
	// Simulated LEFT JOIN from GetByUUID.
	w.CompensationPeriodUUID = sql.NullString{String: "period-uuid-7", Valid: true}
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{wsUUID: w}}
	uc := newWorksheetUC(wsDB, &fakeCompensationPeriodDB{})

	got, err := uc.GetWorksheet(testCtx(), wsUUID)
	if err != nil {
		t.Fatalf("GetWorksheet: %v", err)
	}
	if !got.CompensationPeriodUUID.Valid || got.CompensationPeriodUUID.String != "period-uuid-7" {
		t.Fatalf("compensation_period_uuid = %+v, want period-uuid-7", got.CompensationPeriodUUID)
	}
}

func TestListWorksheets_PassesPeriodUUID(t *testing.T) {
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{}}
	uc := newWorksheetUC(wsDB, nil)

	_, err := uc.ListWorksheets(testCtx(), model.ListWorksheetsRequest{
		CompensationPeriodUUID: "  period-uuid-7  ",
		Limit:                  10,
	})
	if err != nil {
		t.Fatalf("ListWorksheets: %v", err)
	}
	if wsDB.lastList.CompensationPeriodUUID != "period-uuid-7" {
		t.Fatalf("CompensationPeriodUUID = %q, want period-uuid-7", wsDB.lastList.CompensationPeriodUUID)
	}
	if wsDB.lastList.InstitutionID != testInstitutionID {
		t.Fatalf("InstitutionID = %d, want %d", wsDB.lastList.InstitutionID, testInstitutionID)
	}
}

func TestListWorksheets_PassesDateOverlap(t *testing.T) {
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{}}
	uc := newWorksheetUC(wsDB, nil)
	start := time.Date(2026, 9, 1, 15, 0, 0, 0, time.UTC)
	end := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	_, err := uc.ListWorksheets(testCtx(), model.ListWorksheetsRequest{
		PeriodStart: model.Time(start),
		PeriodEnd:   model.Time(end),
	})
	if err != nil {
		t.Fatalf("ListWorksheets: %v", err)
	}
	if !wsDB.lastList.PeriodStart.Time().Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("period_start = %v", wsDB.lastList.PeriodStart.Time())
	}
	if !wsDB.lastList.PeriodEnd.Time().Equal(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("period_end = %v", wsDB.lastList.PeriodEnd.Time())
	}
}

func TestListWorksheets_RejectsInvertedDates(t *testing.T) {
	uc := newWorksheetUC(nil, nil)
	_, err := uc.ListWorksheets(testCtx(), model.ListWorksheetsRequest{
		PeriodStart: model.Time(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)),
		PeriodEnd:   model.Time(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)),
	})
	if errorName(t, err) != errInvalidPeriodDates {
		t.Fatalf("error = %v", err)
	}
}

func TestListWorksheets_OmitsPeriodFilterWhenEmpty(t *testing.T) {
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{}}
	uc := newWorksheetUC(wsDB, nil)

	_, err := uc.ListWorksheets(testCtx(), model.ListWorksheetsRequest{})
	if err != nil {
		t.Fatalf("ListWorksheets: %v", err)
	}
	if wsDB.lastList.CompensationPeriodUUID != "" {
		t.Fatalf("CompensationPeriodUUID = %q, want empty", wsDB.lastList.CompensationPeriodUUID)
	}
}

func TestTrxWorksheet_ToResponse_PeriodUUID(t *testing.T) {
	w := openWorksheet("ws-json")
	w.CompensationPeriodUUID = sql.NullString{String: "p-uuid", Valid: true}
	got := w.ToResponse()
	if !got.CompensationPeriodUUID.Valid || got.CompensationPeriodUUID.String != "p-uuid" {
		t.Fatalf("got %+v", got.CompensationPeriodUUID)
	}
	raw, _ := json.Marshal(got)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	if _, ok := m["compensation_period_id"]; ok {
		t.Fatalf("must not emit compensation_period_id: %s", raw)
	}
}

func TestUpdateWorksheetTotals_AddsPaydayCommissionDelta(t *testing.T) {
	w := openWorksheet("ws-delta")
	w.TotalCommission = 100
	w.CompensationPeriodID = sql.NullInt64{Int64: 7, Valid: true}
	periods := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{augustPayday()}}
	periods.periods[0].TotalCommission = sql.NullInt64{Int64: 100, Valid: true}
	uc := &VisitCommissionUC{
		WorksheetDB:  &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{w.UUID: w}},
		CommissionDB: &stubCommissionDB{sum: 250},
		PeriodDB:     periods,
	}
	got, err := uc.updateWorksheetTotals(context.Background(), w.ID)
	if err != nil {
		t.Fatalf("updateWorksheetTotals: %v", err)
	}
	if got != 250 || periods.commissionDeltaCalls != 1 || periods.periods[0].TotalCommission.Int64 != 250 || periods.periods[0].TotalPayout.Int64 != 250 {
		t.Fatalf("total=%d period=%+v calls=%d", got, periods.periods[0], periods.commissionDeltaCalls)
	}
}

func TestDeleteWorksheet_SubtractsCommissionAndWage(t *testing.T) {
	w := openWorksheet("ws-del")
	w.TotalCommission = 40
	w.CompensationPeriodID = sql.NullInt64{Int64: 7, Valid: true}
	periods := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{augustPayday()}}
	periods.periods[0].TotalCommission = sql.NullInt64{Int64: 40, Valid: true}
	periods.periods[0].TotalWage = sql.NullInt64{Int64: 500, Valid: true}
	periods.periods[0].TotalPayout = sql.NullInt64{Int64: 540, Valid: true}
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID: 3, StaffID: testStaffUUID, InstitutionID: testInstitutionID, CompensationPeriodID: 7,
		TotalWage: sql.NullInt64{Int64: 500, Valid: true},
	}}}
	uc := newWorksheetUC(&patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{w.UUID: w}}, periods)
	uc.CommissionDB = &stubCommissionDB{}
	uc.SnapshotDB = snaps
	uc.Transaction = &fakeTx{}
	if _, err := uc.DeleteWorksheet(testCtx(), w.UUID); err != nil {
		t.Fatalf("DeleteWorksheet: %v", err)
	}
	if periods.periods[0].TotalCommission.Int64 != 0 || periods.periods[0].TotalWage.Int64 != 0 || periods.periods[0].TotalPayout.Int64 != 0 {
		t.Fatalf("period totals = wage %d commission %d payout %d", periods.periods[0].TotalWage.Int64, periods.periods[0].TotalCommission.Int64, periods.periods[0].TotalPayout.Int64)
	}
	if snaps.rows[0].DeleteTime == nil {
		t.Fatal("snapshot was not deleted")
	}
}

func TestUpdateWage_FinalizedPaydayRejected(t *testing.T) {
	snaps := &fakeSnapshotDB{rows: []model.TrxWagePeriodSnapshot{{
		ID: 4, StaffID: "staff-a", InstitutionID: testInstitutionID, CompensationPeriodID: 7,
		WageAmount: 1000, WageCadence: model.WageCadenceMonthly,
	}}}
	uc := newSnapshotUC(nil, snaps, nil)
	uc.PeriodDB.(*fakeCompensationPeriodDB).periods[0].Status = model.CompensationPeriodStatusFinalized
	_, err := uc.Update(testCtx(), 4, model.UpdateStaffWageSnapshotRequest{FinalWage: null.Int64From(10)})
	if errorName(t, err) != errPaydayPeriodFinalized {
		t.Fatalf("error = %v", err)
	}
}

type stubCommissionDB struct {
	sum int64
}

func (s *stubCommissionDB) InsertGeneratedIfMissing(context.Context, []model.TrxVisitCommission) (int, error) {
	return 0, nil
}
func (s *stubCommissionDB) ListByWorksheet(context.Context, compensationrepo.ListVisitCommissionParams) ([]compensationrepo.VisitCommissionListRow, int, error) {
	return nil, 0, nil
}
func (s *stubCommissionDB) ListByStaffDateRange(context.Context, compensationrepo.ListStaffDateCommissionParams) ([]compensationrepo.VisitCommissionListRow, int, error) {
	return nil, 0, nil
}
func (s *stubCommissionDB) SumValidByWorksheet(context.Context, int64) (int64, error) {
	return s.sum, nil
}
func (s *stubCommissionDB) CountVisitsByWorksheet(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s *stubCommissionDB) DistinctVisitIDsByWorksheet(context.Context, int64) ([]int64, error) {
	return nil, nil
}
func (s *stubCommissionDB) GetLiveByID(context.Context, int64, int64) (*model.TrxVisitCommission, bool, error) {
	return nil, false, nil
}
func (s *stubCommissionDB) UpdateAssignmentAmounts(context.Context, *model.TrxVisitCommission) error {
	return nil
}
func (s *stubCommissionDB) SoftDelete(context.Context, int64) (bool, error) { return false, nil }
func (s *stubCommissionDB) SoftDeleteByWorksheet(context.Context, int64) (int64, error) {
	return 0, nil
}
func (s *stubCommissionDB) SumRevenueByVisitIDs(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}
func (s *stubCommissionDB) SoftWarningAggregates(context.Context, int64) ([]compensationrepo.VisitCommissionWarning, error) {
	return nil, nil
}
