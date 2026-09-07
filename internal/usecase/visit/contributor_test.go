package visit

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"testing"

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
)

func testCtx() context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		InstitutionID: testInstitutionID,
		UUID:          "caller-staff",
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
	rows []compensationrepo.DetectedAttribution
	err  error
}

func (f *fakeContributorDB) DetectForVisit(_ context.Context, _, _ int64) ([]compensationrepo.DetectedAttribution, error) {
	if f.err != nil {
		return nil, f.err
	}
	out := make([]compensationrepo.DetectedAttribution, len(f.rows))
	copy(out, f.rows)
	return out, nil
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
