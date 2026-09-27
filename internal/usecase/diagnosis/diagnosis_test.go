package diagnosis

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	diagnosisrepo "github.com/faisalhardin/medilink/internal/entity/repo/diagnosis"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	visituc "github.com/faisalhardin/medilink/internal/usecase/visit"
)

const lockTestInstitutionID int64 = 42

type lockPatientDB struct {
	patientrepo.PatientDB
	visit model.TrxPatientVisit
}

func (s lockPatientDB) GetPatientVisitsByID(context.Context, int64) (model.TrxPatientVisit, error) {
	return s.visit, nil
}

type lockDiagnosisDB struct {
	diagnosisrepo.DiagnosisDB
	softDeleted bool
}

func (s *lockDiagnosisDB) SoftDeleteByID(context.Context, int64, int64, int64) (bool, error) {
	s.softDeleted = true
	return false, nil
}

func lockCtx() context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		InstitutionID: lockTestInstitutionID,
		RolesIDSet:    map[string]bool{role.Administrator: true},
	})
}

func lockedVisitRow() model.TrxPatientVisit {
	return model.TrxPatientVisit{
		ID:               100,
		IDMstInstitution: lockTestInstitutionID,
		CompensationLockedAt: sql.NullTime{
			Time:  time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
			Valid: true,
		},
	}
}

func assertLockError(t *testing.T, err error) {
	t.Helper()
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) {
		t.Fatalf("error type %T: %v", err, err)
	}
	if em.Code != 403 || len(em.ErrorList) != 1 || em.ErrorList[0].ErrorName != visituc.ErrVisitCompensationLocked {
		t.Fatalf("error %+v", em)
	}
}

func TestSave_LockedVisitRejectsAdministrator(t *testing.T) {
	uc := &DiagnosisUC{PatientDB: lockPatientDB{visit: lockedVisitRow()}}
	_, err := uc.Save(lockCtx(), 100, model.SaveDiagnosesRequest{})
	assertLockError(t, err)
}

func TestDelete_LockedVisitDoesNotSoftDelete(t *testing.T) {
	db := &lockDiagnosisDB{}
	uc := &DiagnosisUC{
		PatientDB:   lockPatientDB{visit: lockedVisitRow()},
		DiagnosisDB: db,
	}
	err := uc.Delete(lockCtx(), 100, 7)
	assertLockError(t, err)
	if db.softDeleted {
		t.Fatal("locked visit soft-deleted a diagnosis")
	}
}

func TestDelete_UnlockedVisitIsNotLockError(t *testing.T) {
	db := &lockDiagnosisDB{}
	uc := &DiagnosisUC{
		PatientDB: lockPatientDB{visit: model.TrxPatientVisit{ID: 100, IDMstInstitution: lockTestInstitutionID}},
		DiagnosisDB: db,
	}
	err := uc.Delete(lockCtx(), 100, 7)
	if err == nil {
		t.Fatal("expected not-found from the existing delete path")
	}
	var em *commonerr.ErrorMessage
	if errors.As(err, &em) && len(em.ErrorList) > 0 && em.ErrorList[0].ErrorName == visituc.ErrVisitCompensationLocked {
		t.Fatal("unlocked visit returned the lock error")
	}
	if !db.softDeleted {
		t.Fatal("unlocked visit did not reach soft-delete")
	}
}
