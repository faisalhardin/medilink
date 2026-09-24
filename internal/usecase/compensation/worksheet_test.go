package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/volatiletech/null/v8"
)

// patchFakeWorksheetDB stores worksheets for PatchWorksheet / GetWorksheet tests.
type patchFakeWorksheetDB struct {
	fakeWorksheetDB
	byUUID map[string]*model.TrxWorksheet
	update *model.TrxWorksheet
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

func TestPatchWorksheet_AttachByPeriodUUID(t *testing.T) {
	wsUUID := "ws-1"
	period := &model.TrxCompensationPeriod{
		ID:            7,
		UUID:          "period-uuid-7",
		InstitutionID: testInstitutionID,
		Status:        model.CompensationPeriodStatusOpen,
		Label:         "Aug 2026",
	}
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{wsUUID: openWorksheet(wsUUID)}}
	periodDB := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{period}}
	uc := newWorksheetUC(wsDB, periodDB)

	periodUUID := null.StringFrom("period-uuid-7")
	got, err := uc.PatchWorksheet(testCtx(), model.PatchWorksheetRequest{
		UUID:                   wsUUID,
		CompensationPeriodUUID: &periodUUID,
	})
	if err != nil {
		t.Fatalf("PatchWorksheet: %v", err)
	}
	if !got.CompensationPeriodUUID.Valid || got.CompensationPeriodUUID.String != "period-uuid-7" {
		t.Fatalf("response compensation_period_uuid = %+v, want period-uuid-7", got.CompensationPeriodUUID)
	}
	if wsDB.update == nil || !wsDB.update.CompensationPeriodID.Valid || wsDB.update.CompensationPeriodID.Int64 != 7 {
		t.Fatalf("DB FK not set to period.ID=7: %+v", wsDB.update)
	}

	// Public JSON must expose UUID string, never numeric period id.
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, has := m["compensation_period_id"]; has {
		t.Fatalf("response must not include compensation_period_id: %s", raw)
	}
	if m["compensation_period_uuid"] != "period-uuid-7" {
		t.Fatalf("json compensation_period_uuid = %v, want period-uuid-7", m["compensation_period_uuid"])
	}
}

func TestPatchWorksheet_DetachViaNull(t *testing.T) {
	wsUUID := "ws-2"
	w := openWorksheet(wsUUID)
	w.CompensationPeriodID = sql.NullInt64{Int64: 7, Valid: true}
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{wsUUID: w}}
	periodDB := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{{
		ID: 7, UUID: "period-uuid-7", InstitutionID: testInstitutionID,
		Status: model.CompensationPeriodStatusOpen,
	}}}
	uc := newWorksheetUC(wsDB, periodDB)

	detach := null.String{} // Valid=false → JSON null
	got, err := uc.PatchWorksheet(testCtx(), model.PatchWorksheetRequest{
		UUID:                   wsUUID,
		CompensationPeriodUUID: &detach,
	})
	if err != nil {
		t.Fatalf("PatchWorksheet: %v", err)
	}
	if got.CompensationPeriodUUID.Valid {
		t.Fatalf("expected detached null uuid, got %+v", got.CompensationPeriodUUID)
	}
	if wsDB.update == nil || wsDB.update.CompensationPeriodID.Valid {
		t.Fatalf("expected FK cleared, got %+v", wsDB.update)
	}
}

func TestPatchWorksheet_UnknownPeriodUUID(t *testing.T) {
	wsUUID := "ws-3"
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{wsUUID: openWorksheet(wsUUID)}}
	uc := newWorksheetUC(wsDB, &fakeCompensationPeriodDB{})

	periodUUID := null.StringFrom("missing-period")
	_, err := uc.PatchWorksheet(testCtx(), model.PatchWorksheetRequest{
		UUID:                   wsUUID,
		CompensationPeriodUUID: &periodUUID,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if name := errorName(t, err); name != errPaydayPeriodNotFound {
		t.Fatalf("error name = %q, want %q", name, errPaydayPeriodNotFound)
	}
}

func TestPatchWorksheet_FinalizedPeriodRejected(t *testing.T) {
	wsUUID := "ws-4"
	period := &model.TrxCompensationPeriod{
		ID:            9,
		UUID:          "period-final",
		InstitutionID: testInstitutionID,
		Status:        model.CompensationPeriodStatusFinalized,
	}
	wsDB := &patchFakeWorksheetDB{byUUID: map[string]*model.TrxWorksheet{wsUUID: openWorksheet(wsUUID)}}
	uc := newWorksheetUC(wsDB, &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{period}})

	periodUUID := null.StringFrom("period-final")
	_, err := uc.PatchWorksheet(testCtx(), model.PatchWorksheetRequest{
		UUID:                   wsUUID,
		CompensationPeriodUUID: &periodUUID,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if name := errorName(t, err); name != errPaydayPeriodFinalized {
		t.Fatalf("error name = %q, want %q", name, errPaydayPeriodFinalized)
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
