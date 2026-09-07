package visit

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"
	"time"

	roleconst "github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
)

const (
	testInstitutionID int64 = 42
	testVisitID       int64 = 100
	testStaffA              = "staff-a"
	testStaffB              = "staff-b"
	testStaffC              = "staff-c"
	testStaffUUID           = "11111111-1111-1111-1111-111111111111"
	testCallerUUID          = "22222222-2222-2222-2222-222222222222"
)

func testCtx() context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		InstitutionID: testInstitutionID,
		UUID:          testCallerUUID,
	})
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

func errorCode(t *testing.T, err error) int {
	t.Helper()
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) {
		t.Fatalf("error type %T: %v", err, err)
	}
	return em.Code
}

type fakeVisitGetter struct {
	visit model.TrxPatientVisit
	err   error
}

func (f *fakeVisitGetter) GetPatientVisitsByID(_ context.Context, visitID int64) (model.TrxPatientVisit, error) {
	if f.err != nil {
		return model.TrxPatientVisit{}, f.err
	}
	if f.visit.ID != 0 && f.visit.ID != visitID {
		return model.TrxPatientVisit{}, nil
	}
	return f.visit, nil
}

type fakeContributorDB struct {
	rows       []compensationrepo.DetectedAttribution
	err        error
	upsertErr  error
	lastUpsert *model.MapVisitContributor
}

func (f *fakeContributorDB) DetectForVisit(_ context.Context, _, _ int64) ([]compensationrepo.DetectedAttribution, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]compensationrepo.DetectedAttribution, len(f.rows))
	copy(out, f.rows)
	return out, nil
}

func (f *fakeContributorDB) UpsertManualContributor(_ context.Context, row model.MapVisitContributor) error {
	cp := row
	f.lastUpsert = &cp
	if f.upsertErr != nil {
		return f.upsertErr
	}
	return nil
}

type fakeStaffGetter struct {
	staff model.StaffWithRolesResponse
	err   error
}

func (f *fakeStaffGetter) GetStaffByUUID(_ context.Context, _ int64, _ string, _ bool) (model.StaffWithRolesResponse, error) {
	if f.err != nil {
		return model.StaffWithRolesResponse{}, f.err
	}
	return f.staff, nil
}

func newUC(visit model.TrxPatientVisit, rows []compensationrepo.DetectedAttribution) *VisitContributorUC {
	return NewVisitContributorUC(&VisitContributorUC{
		PatientDB:     &fakeVisitGetter{visit: visit},
		ContributorDB: &fakeContributorDB{rows: rows},
	})
}

func liveVisit() model.TrxPatientVisit {
	return model.TrxPatientVisit{
		ID:               testVisitID,
		IDMstInstitution: testInstitutionID,
	}
}

func byStaff(t *testing.T, rows []model.VisitContributorResponse) map[string]model.VisitContributorResponse {
	t.Helper()
	out := make(map[string]model.VisitContributorResponse, len(rows))
	for _, row := range rows {
		if _, exists := out[row.StaffID]; exists {
			t.Fatalf("duplicate staff_id %s", row.StaffID)
		}
		out[row.StaffID] = row
	}
	return out
}

func TestListVisitContributors_MergeClinicalAndMap(t *testing.T) {
	uc := newUC(liveVisit(), []compensationrepo.DetectedAttribution{
		{
			Type:          model.ContributionSourceTypeDiagnosis,
			StaffID:       testStaffA,
			Name:          "Ana",
			ClinicalRowID: 8,
			DiagnosisID:   8,
			Label:         sql.NullString{String: "K00.1", Valid: true},
		},
		{
			Type:          model.ContributionSourceTypeManual,
			StaffID:       testStaffA,
			Name:          "Ana",
			ClinicalRowID: 1,
		},
	})

	resp, err := uc.ListVisitContributors(testCtx(), testVisitID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := byStaff(t, resp.Contributors)
	row, ok := got[testStaffA]
	if !ok {
		t.Fatal("missing staff-a")
	}
	if row.Source.Type != model.ContributionSourceTypeDiagnosis {
		t.Fatalf("source type %s, want diagnosis", row.Source.Type)
	}
	if !row.AddedManually {
		t.Fatal("expected added_manually true when map row exists")
	}
	if !row.Source.DiagnosisID.Valid || row.Source.DiagnosisID.Int64 != 8 {
		t.Fatalf("diagnosis_id %+v", row.Source.DiagnosisID)
	}
	if !row.Source.Label.Valid || row.Source.Label.String != "K00.1" {
		t.Fatalf("label %+v", row.Source.Label)
	}
	if !row.Source.LabelSource.Valid || row.Source.LabelSource.String != labelSourceICD10Display {
		t.Fatalf("label_source %+v", row.Source.LabelSource)
	}
}

func TestListVisitContributors_Priority(t *testing.T) {
	uc := newUC(liveVisit(), []compensationrepo.DetectedAttribution{
		{
			Type:          model.ContributionSourceTypeDiagnosis,
			StaffID:       testStaffA,
			Name:          "Ana",
			ClinicalRowID: 2,
			DiagnosisID:   2,
			Label:         sql.NullString{String: "K00.1", Valid: true},
		},
		{
			Type:          model.ContributionSourceTypeProcedure,
			StaffID:       testStaffA,
			Name:          "Ana",
			ClinicalRowID: 9,
			ProcedureID:   9,
			ProductID:     sql.NullInt64{Int64: 55, Valid: true},
			Label:         sql.NullString{String: "Scaling", Valid: true},
		},
		{
			Type:          model.ContributionSourceTypeDiagnosis,
			StaffID:       testStaffB,
			Name:          "Budi",
			ClinicalRowID: 3,
			DiagnosisID:   3,
			Label:         sql.NullString{String: "K02", Valid: true},
		},
		{
			Type:          model.ContributionSourceTypeManual,
			StaffID:       testStaffB,
			Name:          "Budi",
			ClinicalRowID: 4,
		},
		{
			Type:          model.ContributionSourceTypeManual,
			StaffID:       testStaffC,
			Name:          "Citra",
			ClinicalRowID: 5,
		},
	})

	resp, err := uc.ListVisitContributors(testCtx(), testVisitID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := byStaff(t, resp.Contributors)

	a := got[testStaffA]
	if a.Source.Type != model.ContributionSourceTypeProcedure {
		t.Fatalf("staff-a type %s, want procedure", a.Source.Type)
	}
	if a.AddedManually {
		t.Fatal("staff-a should not be added_manually")
	}
	if !a.Source.ProcedureID.Valid || a.Source.ProcedureID.Int64 != 9 {
		t.Fatalf("procedure_id %+v", a.Source.ProcedureID)
	}
	if !a.Source.ProductID.Valid || a.Source.ProductID.Int64 != 55 {
		t.Fatalf("product_id %+v", a.Source.ProductID)
	}
	if !a.Source.LabelSource.Valid || a.Source.LabelSource.String != labelSourceProductName {
		t.Fatalf("label_source %+v", a.Source.LabelSource)
	}

	b := got[testStaffB]
	if b.Source.Type != model.ContributionSourceTypeDiagnosis {
		t.Fatalf("staff-b type %s, want diagnosis (beats map)", b.Source.Type)
	}
	if !b.AddedManually {
		t.Fatal("staff-b map row should set added_manually")
	}

	c := got[testStaffC]
	if c.Source.Type != model.ContributionSourceTypeManual {
		t.Fatalf("staff-c type %s, want manual", c.Source.Type)
	}
	if !c.AddedManually {
		t.Fatal("map-only staff must be added_manually")
	}

	if len(resp.Contributors) != 3 {
		t.Fatalf("got %d contributors, want 3", len(resp.Contributors))
	}
	if resp.Contributors[0].Name != "Ana" || resp.Contributors[1].Name != "Budi" || resp.Contributors[2].Name != "Citra" {
		t.Fatalf("sort order %+v", []string{resp.Contributors[0].Name, resp.Contributors[1].Name, resp.Contributors[2].Name})
	}
}

func TestListVisitContributors_LowestClinicalRowID(t *testing.T) {
	uc := newUC(liveVisit(), []compensationrepo.DetectedAttribution{
		{
			Type:          model.ContributionSourceTypeProcedure,
			StaffID:       testStaffA,
			Name:          "Ana",
			ClinicalRowID: 20,
			ProcedureID:   20,
			Label:         sql.NullString{String: "Later", Valid: true},
		},
		{
			Type:          model.ContributionSourceTypeProcedure,
			StaffID:       testStaffA,
			Name:          "Ana",
			ClinicalRowID: 4,
			ProcedureID:   4,
			Label:         sql.NullString{String: "Earlier", Valid: true},
		},
	})

	resp, err := uc.ListVisitContributors(testCtx(), testVisitID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	row := byStaff(t, resp.Contributors)[testStaffA]
	if !row.Source.ProcedureID.Valid || row.Source.ProcedureID.Int64 != 4 {
		t.Fatalf("expected lowest procedure id 4, got %+v", row.Source.ProcedureID)
	}
	if row.Source.Label.String != "Earlier" {
		t.Fatalf("label %s, want Earlier", row.Source.Label.String)
	}
}

func TestListVisitContributors_NotFoundMissingVisit(t *testing.T) {
	uc := newUC(model.TrxPatientVisit{}, nil)
	_, err := uc.ListVisitContributors(testCtx(), testVisitID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errVisitNotFound {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusNotFound {
		t.Fatalf("status %d, want 404", errorCode(t, err))
	}
}

func TestListVisitContributors_NotFoundWrongInstitution(t *testing.T) {
	uc := newUC(model.TrxPatientVisit{
		ID:               testVisitID,
		IDMstInstitution: testInstitutionID + 1,
	}, nil)
	_, err := uc.ListVisitContributors(testCtx(), testVisitID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errVisitNotFound {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusNotFound {
		t.Fatalf("status %d, want 404", errorCode(t, err))
	}
}

func liveStaff() model.StaffWithRolesResponse {
	return model.StaffWithRolesResponse{
		UUID: testStaffUUID,
		Name: "Dewi",
	}
}

func newAddUC(visit model.TrxPatientVisit, staff model.StaffWithRolesResponse, upsertErr error) (*VisitContributorUC, *fakeContributorDB) {
	db := &fakeContributorDB{upsertErr: upsertErr}
	return NewVisitContributorUC(&VisitContributorUC{
		PatientDB:     &fakeVisitGetter{visit: visit},
		StaffDB:       &fakeStaffGetter{staff: staff},
		ContributorDB: db,
	}), db
}

func assertManualContributor(t *testing.T, resp model.AddVisitContributorResponse) {
	t.Helper()
	c := resp.Contributor
	if c.StaffID != testStaffUUID {
		t.Fatalf("staff_id %s", c.StaffID)
	}
	if c.Name != "Dewi" {
		t.Fatalf("name %s", c.Name)
	}
	if c.Source.Type != model.ContributionSourceTypeManual {
		t.Fatalf("source type %s", c.Source.Type)
	}
	if !c.AddedManually {
		t.Fatal("expected added_manually true")
	}
}

func assertUpsertPayload(t *testing.T, got *model.MapVisitContributor) {
	t.Helper()
	if got == nil {
		t.Fatal("upsert was not called")
	}
	if got.VisitID != testVisitID || got.StaffID != testStaffUUID || got.InstitutionID != testInstitutionID || got.AddedBy != testCallerUUID {
		t.Fatalf("upsert payload %+v", got)
	}
}

func TestAddVisitContributor_Insert(t *testing.T) {
	uc, db := newAddUC(liveVisit(), liveStaff(), nil)
	resp, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertManualContributor(t, resp)
	assertUpsertPayload(t, db.lastUpsert)
}

func TestAddVisitContributor_RestoreSoftDeleted(t *testing.T) {
	uc, db := newAddUC(liveVisit(), liveStaff(), nil)
	resp, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err != nil {
		t.Fatalf("restore path should succeed: %v", err)
	}
	assertManualContributor(t, resp)
	assertUpsertPayload(t, db.lastUpsert)
}

func TestAddVisitContributor_DuplicateLive(t *testing.T) {
	uc, _ := newAddUC(liveVisit(), liveStaff(), compensationrepo.ErrContributorAlreadyAdded)
	_, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errContributorAlreadyAdded {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusConflict {
		t.Fatalf("status %d, want 409", errorCode(t, err))
	}
}

func TestAddVisitContributor_LockedVisit(t *testing.T) {
	visit := liveVisit()
	visit.CompensationLockedAt = sql.NullTime{Time: time.Now(), Valid: true}
	uc, db := newAddUC(visit, liveStaff(), nil)
	_, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errVisitCompensationLocked {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusForbidden {
		t.Fatalf("status %d, want 403", errorCode(t, err))
	}
	if db.lastUpsert != nil {
		t.Fatal("locked visit must not write a map row")
	}
}

func TestAddVisitContributor_LockedVisitAdministrator(t *testing.T) {
	visit := liveVisit()
	visit.CompensationLockedAt = sql.NullTime{Time: time.Now(), Valid: true}
	uc, db := newAddUC(visit, liveStaff(), nil)
	ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		InstitutionID: testInstitutionID,
		UUID:          testCallerUUID,
		RolesIDSet:    map[string]bool{roleconst.Administrator: true},
	})
	_, err := uc.AddVisitContributor(ctx, testVisitID, testStaffUUID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errVisitCompensationLocked {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusForbidden {
		t.Fatalf("status %d, want 403", errorCode(t, err))
	}
	if db.lastUpsert != nil {
		t.Fatal("administrator must not bypass the lock")
	}
}

func TestAddVisitContributor_VisitNotFound(t *testing.T) {
	uc, _ := newAddUC(model.TrxPatientVisit{}, liveStaff(), nil)
	_, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errVisitNotFound {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusNotFound {
		t.Fatalf("status %d, want 404", errorCode(t, err))
	}
}

func TestAddVisitContributor_WrongInstitution(t *testing.T) {
	uc, _ := newAddUC(model.TrxPatientVisit{
		ID:               testVisitID,
		IDMstInstitution: testInstitutionID + 1,
	}, liveStaff(), nil)
	_, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != errVisitNotFound {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusNotFound {
		t.Fatalf("status %d, want 404", errorCode(t, err))
	}
}

func TestAddVisitContributor_StaffNotFound(t *testing.T) {
	uc := NewVisitContributorUC(&VisitContributorUC{
		PatientDB: &fakeVisitGetter{visit: liveVisit()},
		StaffDB: &fakeStaffGetter{
			err: commonerr.SetNewError(http.StatusNotFound, "staff_not_found", "staff was not found in this institution"),
		},
		ContributorDB: &fakeContributorDB{},
	})
	_, err := uc.AddVisitContributor(testCtx(), testVisitID, testStaffUUID)
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != "staff_not_found" {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusNotFound {
		t.Fatalf("status %d, want 404", errorCode(t, err))
	}
}

func TestAddVisitContributor_InvalidStaffID(t *testing.T) {
	uc, db := newAddUC(liveVisit(), liveStaff(), nil)
	_, err := uc.AddVisitContributor(testCtx(), testVisitID, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error")
	}
	if errorName(t, err) != "invalid" {
		t.Fatalf("error name %s", errorName(t, err))
	}
	if errorCode(t, err) != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", errorCode(t, err))
	}
	if db.lastUpsert != nil {
		t.Fatal("invalid staff_id must not write a map row")
	}
}
