package compensation

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
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

type fakeCompensationPeriodDB struct {
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

func (f *fakeCompensationPeriodDB) Create(_ context.Context, period *model.TrxCompensationPeriod) error {
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

func (f *fakeCompensationPeriodDB) GetByUUID(_ context.Context, institutionID int64, uuid string) (*model.TrxCompensationPeriod, bool, error) {
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

func (f *fakeCompensationPeriodDB) GetByID(_ context.Context, institutionID, id int64) (*model.TrxCompensationPeriod, bool, error) {
	if f.getErr != nil {
		return nil, false, f.getErr
	}
	for _, p := range f.periods {
		if p.ID == id && p.InstitutionID == institutionID {
			cp := *p
			return &cp, true, nil
		}
	}
	return nil, false, nil
}

func (f *fakeCompensationPeriodDB) List(_ context.Context, params model.ListCompensationPeriodParams) ([]model.TrxCompensationPeriod, int, error) {
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

func (f *fakeCompensationPeriodDB) UpdateStatusAndTotals(_ context.Context, period *model.TrxCompensationPeriod) error {
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

func (f *fakeCompensationPeriodDB) SoftDelete(_ context.Context, institutionID int64, uuid string) (bool, error) {
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

// fakeWorksheetDB implements compensationrepo.WorksheetDB for tests.
type fakeWorksheetDB struct {
	// periodTotals maps compensationPeriodID → WorksheetPeriodTotals
	periodTotals map[int64]compensationrepo.WorksheetPeriodTotals
	// staffTotals maps compensationPeriodID → per-staff totals
	staffTotals map[int64][]compensationrepo.StaffCommissionTotals
	lastList    model.ListWorksheetsRequest
	sumErr      error
	staffErr    error
}

func (f *fakeWorksheetDB) Create(_ context.Context, _ *model.TrxWorksheet) error { return nil }
func (f *fakeWorksheetDB) GetByUUID(_ context.Context, _ int64, _ string) (*model.TrxWorksheet, bool, error) {
	return nil, false, nil
}
func (f *fakeWorksheetDB) GetByID(_ context.Context, _ int64) (*model.TrxWorksheet, bool, error) {
	return nil, false, nil
}
func (f *fakeWorksheetDB) List(_ context.Context, params model.ListWorksheetsRequest) ([]model.TrxWorksheet, error) {
	f.lastList = params
	return nil, nil
}
func (f *fakeWorksheetDB) Update(_ context.Context, _ *model.TrxWorksheet) error { return nil }
func (f *fakeWorksheetDB) SoftDelete(_ context.Context, _ int64, _ string) (bool, error) {
	return false, nil
}
func (f *fakeWorksheetDB) MarkGeneratePending(_ context.Context, _ int64) error { return nil }
func (f *fakeWorksheetDB) MarkGenerateFinished(_ context.Context, _ int64, _ model.WorksheetGenerateStatus, _ string) error {
	return nil
}
func (f *fakeWorksheetDB) UpdateTotals(_ context.Context, _ int64, _ int64, _ int64) error {
	return nil
}
func (f *fakeWorksheetDB) ExistsOverlapping(_ context.Context, _ int64, _ string, _, _ time.Time, _ int64) (bool, error) {
	return false, nil
}

func (f *fakeWorksheetDB) SumByCompensationPeriod(_ context.Context, compensationPeriodID int64) (compensationrepo.WorksheetPeriodTotals, error) {
	if f.sumErr != nil {
		return compensationrepo.WorksheetPeriodTotals{}, f.sumErr
	}
	if f.periodTotals == nil {
		return compensationrepo.WorksheetPeriodTotals{}, nil
	}
	return f.periodTotals[compensationPeriodID], nil
}

func (f *fakeWorksheetDB) SumByStaffForCompensationPeriod(_ context.Context, compensationPeriodID int64) ([]compensationrepo.StaffCommissionTotals, error) {
	if f.staffErr != nil {
		return nil, f.staffErr
	}
	if f.staffTotals == nil {
		return nil, nil
	}
	return append([]compensationrepo.StaffCommissionTotals(nil), f.staffTotals[compensationPeriodID]...), nil
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

func newUC(db *fakeCompensationPeriodDB, worksheets *fakeWorksheetDB, tx *fakeTx) *CompensationPeriodUC {
	if db == nil {
		db = &fakeCompensationPeriodDB{}
	}
	if worksheets == nil {
		worksheets = &fakeWorksheetDB{}
	}
	if tx == nil {
		tx = &fakeTx{}
	}
	return &CompensationPeriodUC{
		CompensationPeriodDB: db,
		WorksheetDB:          worksheets,
		Transaction:          tx,
		now: func() time.Time {
			return time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
		},
	}
}

func TestCreatePeriod(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{}
		uc := newUC(db, nil, nil)
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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{{
			UUID:          "existing",
			InstitutionID: testInstitutionID,
			PeriodStart:   time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
			PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
			Status:        model.CompensationPeriodStatusOpen,
		}}}
		uc := newUC(db, nil, nil)
		_, err := uc.CreatePeriod(testCtx(), createReq("Overlap", "2026-08-01", "2026-08-15"))
		if errorName(t, err) != errDateRangeOverlap {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errDateRangeOverlap)
		}
	})

	t.Run("end before start", func(t *testing.T) {
		uc := newUC(nil, nil, nil)
		_, err := uc.CreatePeriod(testCtx(), createReq("Bad", "2026-08-31", "2026-08-01"))
		if errorName(t, err) != errInvalidPeriodDates {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errInvalidPeriodDates)
		}
	})

	t.Run("empty label", func(t *testing.T) {
		uc := newUC(nil, nil, nil)
		_, err := uc.CreatePeriod(testCtx(), createReq("  ", "2026-08-01", "2026-08-31"))
		if errorName(t, err) != errInvalidLabel {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errInvalidLabel)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		uc := newUC(nil, nil, nil)
		_, err := uc.CreatePeriod(context.Background(), createReq("Aug", "2026-08-01", "2026-08-31"))
		if err == nil {
			t.Fatal("expected unauthorized")
		}
	})
}

func TestListPeriods_DefaultLimit(t *testing.T) {
	db := &fakeCompensationPeriodDB{}
	uc := newUC(db, nil, nil)
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
	db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{
		{UUID: "a", InstitutionID: testInstitutionID, Status: model.CompensationPeriodStatusOpen, Label: "A"},
		{UUID: "b", InstitutionID: testInstitutionID, Status: model.CompensationPeriodStatusOpen, Label: "B"},
		{UUID: "c", InstitutionID: testInstitutionID, Status: model.CompensationPeriodStatusOpen, Label: "C"},
	}}
	uc := newUC(db, nil, nil)

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
	uc := newUC(nil, nil, nil)
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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		worksheets := &fakeWorksheetDB{periodTotals: map[int64]compensationrepo.WorksheetPeriodTotals{
			1: {TotalCommission: 1000, StaffCount: 2, VisitCount: 4},
		}}
		uc := newUC(db, worksheets, nil)
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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		uc := newUC(db, &fakeWorksheetDB{}, nil)
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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{p}}
		uc := newUC(db, nil, nil)
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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{p, other}}
		uc := newUC(db, nil, nil)
		_, err := uc.DraftPeriod(testCtx(), "p1")
		if errorName(t, err) != errDateRangeOverlap {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errDateRangeOverlap)
		}
	})
}

func TestReopenPeriod(t *testing.T) {
	db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{{
		ID:            1,
		UUID:          "p1",
		InstitutionID: testInstitutionID,
		Status:        model.CompensationPeriodStatusFinalized,
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		FinalizedAt:   sql.NullTime{Time: time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC), Valid: true},
		FinalizedBy:   sql.NullString{String: testStaffUUID, Valid: true},
	}}}
	uc := newUC(db, nil, nil)
	got, err := uc.ReopenPeriod(testCtx(), "p1")
	if err != nil {
		t.Fatalf("ReopenPeriod: %v", err)
	}
	if got.Status != model.CompensationPeriodStatusDraft {
		t.Fatalf("status = %s", got.Status)
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
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		uc := newUC(db, nil, nil)
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
			db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{p}}
			uc := newUC(db, nil, nil)
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

type fakeContributorDB struct {
	detections         []compensationrepo.PeriodStaffDetection
	attributions       []compensationrepo.DetectedAttribution
	err                error
	detectForVisitsErr error
	lastInst           int64
	lastStart          time.Time
	lastEnd            time.Time
	lastStaffID        string
}

func (f *fakeContributorDB) DetectForVisit(context.Context, int64, int64) ([]compensationrepo.DetectedAttribution, error) {
	return nil, nil
}

func (f *fakeContributorDB) DetectForVisits(_ context.Context, _ int64, visitIDs []int64) ([]compensationrepo.DetectedAttribution, error) {
	if f.detectForVisitsErr != nil {
		return nil, f.detectForVisitsErr
	}
	if len(visitIDs) == 0 || f.attributions == nil {
		return nil, nil
	}
	want := make(map[int64]struct{}, len(visitIDs))
	for _, id := range visitIDs {
		want[id] = struct{}{}
	}
	out := make([]compensationrepo.DetectedAttribution, 0)
	for _, attr := range f.attributions {
		if _, ok := want[attr.VisitID]; ok {
			out = append(out, attr)
		}
	}
	return out, nil
}

func (f *fakeContributorDB) DetectStaffForPeriod(_ context.Context, institutionID int64, periodStart, periodEndExclusive time.Time) ([]compensationrepo.PeriodStaffDetection, error) {
	f.lastInst = institutionID
	f.lastStart = periodStart
	f.lastEnd = periodEndExclusive
	if f.err != nil {
		return nil, f.err
	}
	out := make([]compensationrepo.PeriodStaffDetection, len(f.detections))
	copy(out, f.detections)
	return out, nil
}

func (f *fakeContributorDB) DetectForPeriodStaff(_ context.Context, institutionID int64, staffID string, periodStart, periodEndExclusive time.Time) ([]compensationrepo.DetectedAttribution, error) {
	f.lastInst = institutionID
	f.lastStaffID = staffID
	f.lastStart = periodStart
	f.lastEnd = periodEndExclusive
	if f.err != nil {
		return nil, f.err
	}
	out := make([]compensationrepo.DetectedAttribution, 0, len(f.attributions))
	for _, attr := range f.attributions {
		if attr.StaffID != "" && attr.StaffID != staffID {
			continue
		}
		out = append(out, attr)
	}
	return out, nil
}

func (f *fakeContributorDB) UpsertManualContributor(context.Context, model.MapVisitContributor) error {
	return nil
}

func (f *fakeContributorDB) DeleteManualContributor(context.Context, int64, int64, string) (bool, error) {
	return false, nil
}

type fakeStaffDB struct {
	staff model.StaffWithRolesResponse
	err   error
}

func (f *fakeStaffDB) GetStaffByUUID(context.Context, int64, string, bool) (model.StaffWithRolesResponse, error) {
	if f.err != nil {
		return model.StaffWithRolesResponse{}, f.err
	}
	return f.staff, nil
}

func (f *fakeStaffDB) ListStaffByInstitution(context.Context, int64, bool) ([]model.StaffWithRolesResponse, error) {
	return nil, nil
}
func (f *fakeStaffDB) GetStaffIDByUUID(context.Context, int64, string) (int64, error) {
	return 0, nil
}
func (f *fakeStaffDB) InsertStaff(context.Context, *model.MstStaff) error { return nil }
func (f *fakeStaffDB) DeactivateStaff(context.Context, int64, string) error {
	return nil
}
func (f *fakeStaffDB) ActivateStaff(context.Context, int64, string) error { return nil }
func (f *fakeStaffDB) AssignRole(context.Context, int64, int64) error      { return nil }
func (f *fakeStaffDB) UnassignRole(context.Context, int64, int64) error    { return nil }
func (f *fakeStaffDB) HasRoleAssignment(context.Context, int64, int64) (bool, error) {
	return false, nil
}
func (f *fakeStaffDB) GetRolePKByBusinessID(context.Context, int64) (model.MstRole, bool, error) {
	return model.MstRole{}, false, nil
}
func (f *fakeStaffDB) GetRolePKsByBusinessIDs(context.Context, []int64) ([]model.MstRole, error) {
	return nil, nil
}
func (f *fakeStaffDB) EmailExistsActiveGlobally(context.Context, string) (bool, error) {
	return false, nil
}
func (f *fakeStaffDB) EmailExistsActiveInOtherInstitution(context.Context, int64, string) (bool, error) {
	return false, nil
}
func (f *fakeStaffDB) CountActiveStaffWithRole(context.Context, int64, string) (int64, error) {
	return 0, nil
}

func TestListPeriodStaff(t *testing.T) {
	open := &model.TrxCompensationPeriod{
		ID:            7,
		UUID:          "p-staff",
		InstitutionID: testInstitutionID,
		Label:         "Aug 2026",
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		Status:        model.CompensationPeriodStatusOpen,
	}

	t.Run("not found", func(t *testing.T) {
		uc := newUC(nil, nil, nil)
		uc.ContributorDB = &fakeContributorDB{}
		_, err := uc.ListPeriodStaff(testCtx(), "missing", model.ListCompensationPeriodStaffRequest{})
		if errorName(t, err) != errPeriodNotFound {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errPeriodNotFound)
		}
	})

	t.Run("other institution", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{{
			ID:            7,
			UUID:          "p-staff",
			InstitutionID: testInstitutionID + 1,
			PeriodStart:   open.PeriodStart,
			PeriodEnd:     open.PeriodEnd,
		}}}
		uc := newUC(db, nil, nil)
		uc.ContributorDB = &fakeContributorDB{}
		_, err := uc.ListPeriodStaff(testCtx(), "p-staff", model.ListCompensationPeriodStaffRequest{})
		if errorName(t, err) != errPeriodNotFound {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errPeriodNotFound)
		}
	})

	t.Run("assignment status wage zero pagination", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		contributors := &fakeContributorDB{detections: []compensationrepo.PeriodStaffDetection{
			{StaffID: "s-a", Name: "Ada", Roles: []string{"Doctor"}, VisitCount: 3},
			{StaffID: "s-b", Name: "Budi", Roles: []string{"Nurse"}, VisitCount: 2},
			{StaffID: "s-c", Name: "Citra", Roles: nil, VisitCount: 1},
		}}
		worksheets := &fakeWorksheetDB{staffTotals: map[int64][]compensationrepo.StaffCommissionTotals{
			7: {
				{StaffID: "s-a", TotalCommission: 1500, VisitCount: 3},
				{StaffID: "s-b", TotalCommission: 400, VisitCount: 1},
			},
		}}
		uc := newUC(db, worksheets, nil)
		uc.ContributorDB = contributors

		got, err := uc.ListPeriodStaff(testCtx(), "p-staff", model.ListCompensationPeriodStaffRequest{
			CommonRequestPayload: model.CommonRequestPayload{Limit: 2, Offset: 0},
		})
		if err != nil {
			t.Fatalf("ListPeriodStaff: %v", err)
		}
		if got.Total != 3 {
			t.Fatalf("total = %d, want 3", got.Total)
		}
		if len(got.Staff) != 2 {
			t.Fatalf("len(staff) = %d, want 2", len(got.Staff))
		}
		if got.Staff[0].StaffID != "s-a" || got.Staff[0].AssignmentStatus != model.CompensationAssignmentStatusComplete || got.Staff[0].Wage != 0 || got.Staff[0].PayTotal != 1500 {
			t.Fatalf("staff[0] = %+v", got.Staff[0])
		}
		if got.Staff[1].StaffID != "s-b" || got.Staff[1].AssignmentStatus != model.CompensationAssignmentStatusPartial || got.Staff[1].CommissionSubtotal != 400 {
			t.Fatalf("staff[1] = %+v", got.Staff[1])
		}

		page2, err := uc.ListPeriodStaff(testCtx(), "p-staff", model.ListCompensationPeriodStaffRequest{
			CommonRequestPayload: model.CommonRequestPayload{Limit: 2, Offset: 2},
		})
		if err != nil {
			t.Fatalf("ListPeriodStaff page2: %v", err)
		}
		if page2.Total != 3 || len(page2.Staff) != 1 {
			t.Fatalf("page2 total=%d len=%d", page2.Total, len(page2.Staff))
		}
		if page2.Staff[0].StaffID != "s-c" || page2.Staff[0].AssignmentStatus != model.CompensationAssignmentStatusUnassigned || page2.Staff[0].PayTotal != 0 {
			t.Fatalf("staff page2 = %+v", page2.Staff[0])
		}
		if page2.Staff[0].Roles == nil || len(page2.Staff[0].Roles) != 0 {
			t.Fatalf("nil roles should serialize as empty slice: %#v", page2.Staff[0].Roles)
		}
		wantEnd := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		if contributors.lastInst != testInstitutionID || !contributors.lastStart.Equal(open.PeriodStart) || !contributors.lastEnd.Equal(wantEnd) {
			t.Fatalf("detect window inst=%d start=%s end=%s", contributors.lastInst, contributors.lastStart, contributors.lastEnd)
		}
	})

	t.Run("default limit", func(t *testing.T) {
		detections := make([]compensationrepo.PeriodStaffDetection, 0, 51)
		for i := 0; i < 51; i++ {
			detections = append(detections, compensationrepo.PeriodStaffDetection{
				StaffID:    "s-" + string(rune('a'+i%26)) + string(rune('0'+i/26)),
				Name:       "N",
				VisitCount: 1,
			})
		}
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		uc := newUC(db, &fakeWorksheetDB{}, nil)
		uc.ContributorDB = &fakeContributorDB{detections: detections}
		got, err := uc.ListPeriodStaff(testCtx(), "p-staff", model.ListCompensationPeriodStaffRequest{})
		if err != nil {
			t.Fatalf("ListPeriodStaff: %v", err)
		}
		if got.Total != 51 || len(got.Staff) != defaultListLimit {
			t.Fatalf("total=%d len=%d, want total 51 len %d", got.Total, len(got.Staff), defaultListLimit)
		}
	})

	t.Run("contributors only", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		worksheets := &fakeWorksheetDB{staffTotals: map[int64][]compensationrepo.StaffCommissionTotals{
			7: {{StaffID: "wage-only", TotalCommission: 0, VisitCount: 0}},
		}}
		uc := newUC(db, worksheets, nil)
		uc.ContributorDB = &fakeContributorDB{detections: []compensationrepo.PeriodStaffDetection{
			{StaffID: "s-a", Name: "Ada", VisitCount: 1},
		}}
		got, err := uc.ListPeriodStaff(testCtx(), "p-staff", model.ListCompensationPeriodStaffRequest{})
		if err != nil {
			t.Fatalf("ListPeriodStaff: %v", err)
		}
		if got.Total != 1 || got.Staff[0].StaffID != "s-a" {
			t.Fatalf("got %+v, want contributors only", got)
		}
	})
}

func TestGetPeriodStaff(t *testing.T) {
	open := &model.TrxCompensationPeriod{
		ID:            7,
		UUID:          "p-staff",
		InstitutionID: testInstitutionID,
		Label:         "Aug 2026",
		PeriodStart:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:     time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		Status:        model.CompensationPeriodStatusOpen,
	}
	staffID := "s-a"
	staff := model.StaffWithRolesResponse{
		UUID: staffID,
		Name: "Ada",
		Roles: []model.StaffRoleResponse{
			{RoleID: 1, Name: "Doctor"},
		},
	}

	t.Run("unauthorized", func(t *testing.T) {
		uc := newUC(nil, nil, nil)
		_, err := uc.GetPeriodStaff(context.Background(), model.GetCompensationPeriodStaffRequest{
			PeriodUUID: "p-staff",
			StaffID:    staffID,
		})
		if err == nil {
			t.Fatal("expected unauthorized")
		}
	})

	t.Run("period not found", func(t *testing.T) {
		uc := newUC(nil, nil, nil)
		uc.StaffDB = &fakeStaffDB{staff: staff}
		_, err := uc.GetPeriodStaff(testCtx(), model.GetCompensationPeriodStaffRequest{
			PeriodUUID: "missing",
			StaffID:    staffID,
		})
		if errorName(t, err) != errPeriodNotFound {
			t.Fatalf("error name = %s, want %s", errorName(t, err), errPeriodNotFound)
		}
	})

	t.Run("staff not found", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		uc := newUC(db, nil, nil)
		uc.StaffDB = &fakeStaffDB{err: commonerr.SetNewError(http.StatusNotFound, "staff_not_found", "staff was not found in this institution")}
		_, err := uc.GetPeriodStaff(testCtx(), model.GetCompensationPeriodStaffRequest{
			PeriodUUID: "p-staff",
			StaffID:    staffID,
		})
		if errorName(t, err) != "staff_not_found" {
			t.Fatalf("error name = %s, want staff_not_found", errorName(t, err))
		}
	})

	t.Run("header staff_info wage stubs empty roles", func(t *testing.T) {
		db := &fakeCompensationPeriodDB{periods: []*model.TrxCompensationPeriod{copyPeriod(open)}}
		uc := newUC(db, nil, nil)
		uc.StaffDB = &fakeStaffDB{staff: staff}

		got, err := uc.GetPeriodStaff(testCtx(), model.GetCompensationPeriodStaffRequest{
			PeriodUUID: "p-staff",
			StaffID:    staffID,
		})
		if err != nil {
			t.Fatalf("GetPeriodStaff: %v", err)
		}
		if got.ComputedWage != 0 || got.WageOverride.Valid {
			t.Fatalf("wage stub failed: computed=%d override=%+v", got.ComputedWage, got.WageOverride)
		}
		if got.StaffInfo.StaffID != staffID || got.StaffInfo.Name != "Ada" || len(got.StaffInfo.Roles) != 1 || got.StaffInfo.Roles[0] != "Doctor" {
			t.Fatalf("staff_info = %+v", got.StaffInfo)
		}

		uc.StaffDB = &fakeStaffDB{staff: model.StaffWithRolesResponse{UUID: staffID, Name: "Ada", Roles: nil}}
		emptyRoles, err := uc.GetPeriodStaff(testCtx(), model.GetCompensationPeriodStaffRequest{
			PeriodUUID: "p-staff",
			StaffID:    staffID,
		})
		if err != nil {
			t.Fatalf("GetPeriodStaff empty roles: %v", err)
		}
		if emptyRoles.StaffInfo.Roles == nil || len(emptyRoles.StaffInfo.Roles) != 0 {
			t.Fatalf("roles should be empty slice: %#v", emptyRoles.StaffInfo.Roles)
		}
	})
}
