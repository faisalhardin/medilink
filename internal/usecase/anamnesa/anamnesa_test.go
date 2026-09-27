package anamnesa

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	visituc "github.com/faisalhardin/medilink/internal/usecase/visit"
	"github.com/go-xorm/xorm"
)

const lockTestInstitutionID int64 = 42

type lockPatientDB struct {
	patientrepo.PatientDB
	visit model.TrxPatientVisit
}

func (s lockPatientDB) GetPatientVisitsByID(context.Context, int64) (model.TrxPatientVisit, error) {
	return s.visit, nil
}

type stopTx struct {
	begun bool
}

var errStopWrite = errors.New("stop")

func (t *stopTx) Begin(context.Context) (*xorm.Session, error) {
	t.begun = true
	return nil, errStopWrite
}

func (t *stopTx) Finish(*xorm.Session, *error) {}

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

func openVisitRow() model.TrxPatientVisit {
	return model.TrxPatientVisit{ID: 100, IDMstInstitution: lockTestInstitutionID}
}

func assertLockError(t *testing.T, err error) {
	t.Helper()
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) {
		t.Fatalf("error type %T: %v", err, err)
	}
	if em.Code != 403 {
		t.Fatalf("status %d", em.Code)
	}
	if len(em.ErrorList) != 1 || em.ErrorList[0].ErrorName != visituc.ErrVisitCompensationLocked {
		t.Fatalf("error list %+v", em.ErrorList)
	}
}

func TestUpsert_LockedVisitRejectsAdministrator(t *testing.T) {
	tx := &stopTx{}
	uc := &AnamnesaUC{
		PatientDB:   lockPatientDB{visit: lockedVisitRow()},
		Transaction: tx,
	}
	_, err := uc.Upsert(lockCtx(), 100, model.UpsertAnamnesaRequest{})
	assertLockError(t, err)
	if tx.begun {
		t.Fatal("locked visit started a write transaction")
	}
}

func TestUpsert_UnlockedVisitIsNotLockError(t *testing.T) {
	tx := &stopTx{}
	uc := &AnamnesaUC{
		PatientDB:   lockPatientDB{visit: openVisitRow()},
		Transaction: tx,
	}
	_, err := uc.Upsert(lockCtx(), 100, model.UpsertAnamnesaRequest{})
	if err == nil || !strings.Contains(err.Error(), errStopWrite.Error()) {
		t.Fatalf("unlocked visit error %v", err)
	}
	var em *commonerr.ErrorMessage
	if errors.As(err, &em) && len(em.ErrorList) > 0 && em.ErrorList[0].ErrorName == visituc.ErrVisitCompensationLocked {
		t.Fatal("unlocked visit returned the lock error")
	}
	if !tx.begun {
		t.Fatal("unlocked visit did not reach the write")
	}
}

var _ xormlib.DBTransactionInterface = (*stopTx)(nil)
