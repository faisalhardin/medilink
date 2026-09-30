package visit

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionrepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
	journeyrepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-xorm/xorm"
)

type lockPatientDB struct {
	patientrepo.PatientDB
	visit          model.TrxPatientVisit
	updatedVisit   bool
	deletedVisit   bool
	updatedDetail  bool
	insertedDetail bool
}

func (s *lockPatientDB) GetPatientVisitsByID(context.Context, int64) (model.TrxPatientVisit, error) {
	return s.visit, nil
}

func (s *lockPatientDB) GetPatientVisits(context.Context, model.GetPatientVisitParams) ([]model.GetPatientVisitResponse, error) {
	return []model.GetPatientVisitResponse{{
		TrxPatientVisit: model.TrxPatientVisit{ID: s.visit.ID, IDMstInstitution: s.visit.IDMstInstitution},
	}}, nil
}

func (s *lockPatientDB) GetDtlPatientVisit(context.Context, model.GetDtlPatientVisitParams) ([]model.DtlPatientVisitWithShortID, error) {
	return []model.DtlPatientVisitWithShortID{{
		DtlPatientVisit: model.DtlPatientVisit{IDTrxPatientVisit: s.visit.ID},
	}}, nil
}

func (s *lockPatientDB) UpdatePatientVisit(context.Context, model.UpdatePatientVisitRequest) (model.TrxPatientVisit, error) {
	s.updatedVisit = true
	return model.TrxPatientVisit{}, nil
}

func (s *lockPatientDB) DeletePatientVisit(context.Context, *model.TrxPatientVisit) error {
	s.deletedVisit = true
	return nil
}

func (s *lockPatientDB) UpdateDtlPatientVisit(context.Context, *model.DtlPatientVisit) error {
	s.updatedDetail = true
	return nil
}

func (s *lockPatientDB) InsertDtlPatientVisit(context.Context, *model.DtlPatientVisit) error {
	s.insertedDetail = true
	return nil
}

type stopInstitution struct {
	institutionrepo.InstitutionDB
}

func (stopInstitution) FindTrxInstitutionProductJoinStockByParams(context.Context, model.FindTrxInstitutionProductParams) ([]model.GetInstitutionProductResponse, error) {
	return nil, errors.New("stop")
}

type stopJourney struct {
	journeyrepo.JourneyDB
}

func (stopJourney) GetJourneyPoint(context.Context, model.MstJourneyPoint) (*model.MstJourneyPoint, error) {
	return nil, errors.New("stop")
}

type noopTx struct{}

func (noopTx) Begin(context.Context) (*xorm.Session, error) { return nil, nil }
func (noopTx) Finish(*xorm.Session, *error)                 {}

func adminCtx() context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		InstitutionID: testInstitutionID,
		UUID:          testCallerUUID,
		RolesIDSet:    map[string]bool{role.Administrator: true},
	})
}

func lockedVisitRow() model.TrxPatientVisit {
	return model.TrxPatientVisit{
		ID:               testVisitID,
		IDMstInstitution: testInstitutionID,
		CompensationLockedAt: sql.NullTime{
			Time:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Valid: true,
		},
	}
}

func openVisitRow() model.TrxPatientVisit {
	return model.TrxPatientVisit{ID: testVisitID, IDMstInstitution: testInstitutionID}
}

func assertLockError(t *testing.T, err error) {
	t.Helper()
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) {
		t.Fatalf("error type %T: %v", err, err)
	}
	if em.Code != 403 || len(em.ErrorList) != 1 || em.ErrorList[0].ErrorName != ErrVisitCompensationLocked {
		t.Fatalf("error %+v", em)
	}
}

func assertNotLockError(t *testing.T, err error) {
	t.Helper()
	var em *commonerr.ErrorMessage
	if errors.As(err, &em) && len(em.ErrorList) > 0 && em.ErrorList[0].ErrorName == ErrVisitCompensationLocked {
		t.Fatal("unlocked visit returned the lock error")
	}
}

func TestUpsertVisitProduct_LockedVisitRejectsAdministrator(t *testing.T) {
	db := &lockPatientDB{visit: lockedVisitRow()}
	uc := &VisitUC{PatientDB: db}
	err := uc.UpsertVisitProduct(adminCtx(), model.UpsertTrxVisitProductRequest{IDTrxPatientVisit: testVisitID})
	assertLockError(t, err)
}

func TestUpsertVisitProduct_UnlockedVisitIsNotLockError(t *testing.T) {
	db := &lockPatientDB{visit: openVisitRow()}
	uc := &VisitUC{PatientDB: db, InstitutionRepo: stopInstitution{}}
	err := uc.UpsertVisitProduct(adminCtx(), model.UpsertTrxVisitProductRequest{IDTrxPatientVisit: testVisitID})
	if err == nil {
		t.Fatal("expected the existing product lookup to fail")
	}
	assertNotLockError(t, err)
}

func TestInsertVisitProduct_LockedVisitDoesNotInsert(t *testing.T) {
	db := &lockPatientDB{visit: lockedVisitRow()}
	uc := &VisitUC{PatientDB: db}
	err := uc.InsertVisitProduct(adminCtx(), model.InsertTrxVisitProductRequest{IDDtlPatientVisit: 9})
	assertLockError(t, err)
}

func TestUpdateVisitProduct_LockedAndUnlocked(t *testing.T) {
	lockedDB := &lockPatientDB{visit: lockedVisitRow()}
	uc := &VisitUC{PatientDB: lockedDB}
	err := uc.UpdateVisitProduct(adminCtx(), model.InsertTrxVisitProductRequest{IDTrxPatientVisit: testVisitID})
	assertLockError(t, err)
	if lockedDB.updatedDetail {
		t.Fatal("locked visit updated a visit detail")
	}

	openDB := &lockPatientDB{visit: openVisitRow()}
	uc.PatientDB = openDB
	err = uc.UpdateVisitProduct(adminCtx(), model.InsertTrxVisitProductRequest{IDTrxPatientVisit: testVisitID})
	assertNotLockError(t, err)
	if err != nil {
		t.Fatalf("unlocked update %v", err)
	}
	if !openDB.updatedDetail {
		t.Fatal("unlocked visit did not update the detail")
	}
}

func TestUpsertVisitTouchpoint_LockedVisitRejectsAdministrator(t *testing.T) {
	db := &lockPatientDB{visit: lockedVisitRow()}
	uc := &VisitUC{PatientDB: db, Transaction: noopTx{}}
	_, err := uc.UpsertVisitTouchpoint(adminCtx(), model.DtlPatientVisitRequest{IDTrxPatientVisit: testVisitID})
	assertLockError(t, err)
	if db.insertedDetail {
		t.Fatal("locked visit inserted a touchpoint")
	}

	_, err = uc.UpsertVisitTouchpoint(adminCtx(), model.DtlPatientVisitRequest{ID: 8, IDTrxPatientVisit: testVisitID})
	assertLockError(t, err)
	if db.updatedDetail {
		t.Fatal("locked visit updated a touchpoint")
	}
}

func TestUpsertVisitTouchpoint_UnlockedVisitIsNotLockError(t *testing.T) {
	db := &lockPatientDB{visit: openVisitRow()}
	uc := &VisitUC{PatientDB: db, JourneyDB: stopJourney{}}
	_, err := uc.UpsertVisitTouchpoint(adminCtx(), model.DtlPatientVisitRequest{IDTrxPatientVisit: testVisitID})
	if err == nil {
		t.Fatal("expected the existing journey lookup to fail")
	}
	assertNotLockError(t, err)
	if db.insertedDetail {
		t.Fatal("journey failure still inserted a touchpoint")
	}
}

func TestUpdatePatientVisit_LockedAndUnlocked(t *testing.T) {
	lockedDB := &lockPatientDB{visit: lockedVisitRow()}
	uc := &VisitUC{PatientDB: lockedDB}
	err := uc.UpdatePatientVisit(adminCtx(), model.UpdatePatientVisitRequest{ID: testVisitID})
	assertLockError(t, err)
	if lockedDB.updatedVisit {
		t.Fatal("locked visit updated the visit row")
	}

	openDB := &lockPatientDB{visit: openVisitRow()}
	uc.PatientDB = openDB
	err = uc.UpdatePatientVisit(adminCtx(), model.UpdatePatientVisitRequest{ID: testVisitID})
	assertNotLockError(t, err)
	if err != nil {
		t.Fatalf("unlocked update %v", err)
	}
	if !openDB.updatedVisit {
		t.Fatal("unlocked visit did not update")
	}
}

func TestUpdatePatientVisit_OtherInstitutionStaysOnExistingPath(t *testing.T) {
	other := lockedVisitRow()
	other.IDMstInstitution = testInstitutionID + 1
	db := &lockPatientDB{visit: other}
	uc := &VisitUC{PatientDB: db}
	err := uc.UpdatePatientVisit(adminCtx(), model.UpdatePatientVisitRequest{ID: testVisitID})
	assertNotLockError(t, err)
	if !db.updatedVisit {
		t.Fatal("a visit outside the caller institution left the existing update path")
	}
}

func TestArchivePatientVisit_LockedAndUnlocked(t *testing.T) {
	lockedDB := &lockPatientDB{visit: lockedVisitRow()}
	uc := &VisitUC{PatientDB: lockedDB}
	err := uc.ArchivePatientVisit(adminCtx(), model.ArchivePatientVisitRequest{ID: testVisitID})
	assertLockError(t, err)
	if lockedDB.deletedVisit {
		t.Fatal("locked visit archived the visit")
	}

	openDB := &lockPatientDB{visit: openVisitRow()}
	uc.PatientDB = openDB
	err = uc.ArchivePatientVisit(adminCtx(), model.ArchivePatientVisitRequest{ID: testVisitID})
	assertNotLockError(t, err)
	if err != nil {
		t.Fatalf("unlocked archive %v", err)
	}
	if !openDB.deletedVisit {
		t.Fatal("unlocked visit did not archive")
	}
}

func TestRejectIfLocked_Unset(t *testing.T) {
	if err := RejectIfLocked(sql.NullTime{}); err != nil {
		t.Fatalf("unlocked visit returned %v", err)
	}
}

func TestRejectIfLocked_Set(t *testing.T) {
	err := RejectIfLocked(sql.NullTime{Time: time.Now().UTC(), Valid: true})
	if err == nil {
		t.Fatal("expected lock error")
	}
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) {
		t.Fatalf("error type %T", err)
	}
	if em.Code != 403 {
		t.Fatalf("status %d", em.Code)
	}
	if len(em.ErrorList) != 1 || em.ErrorList[0].ErrorName != ErrVisitCompensationLocked {
		t.Fatalf("error list %+v", em.ErrorList)
	}
	if em.ErrorList[0].ErrorDescription != MsgVisitCompensationLocked {
		t.Fatalf("description %q", em.ErrorList[0].ErrorDescription)
	}
}
