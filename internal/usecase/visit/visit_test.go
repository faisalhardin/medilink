package visit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionrepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
	journeyrepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-xorm/xorm"
	"github.com/volatiletech/null/v8"
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

func (s *lockPatientDB) GetPatientVisits(_ context.Context, p model.GetPatientVisitParams) ([]model.GetPatientVisitResponse, error) {
	if s.visit.ID == 0 {
		return nil, nil
	}
	if p.IDPatientVisit != 0 && s.visit.ID != p.IDPatientVisit {
		return nil, nil
	}
	if p.IDMstInstitution != 0 && s.visit.IDMstInstitution != p.IDMstInstitution {
		return nil, nil
	}
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
	uc := &VisitUC{PatientDB: db, JourneyDB: stopJourney{}, Transaction: noopTx{}}
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

const (
	roomLinked   int64 = 11
	roomSecond   int64 = 12
	roomMissing  int64 = 900
	roomForeign  int64 = 901
	roomOrphan   int64 = 902
	roomArchived int64 = 903
)

type stepEvents struct {
	seq []string
}

func (e *stepEvents) add(name string) {
	e.seq = append(e.seq, name)
}

func (e *stepEvents) has(name string) bool {
	for _, step := range e.seq {
		if step == name {
			return true
		}
	}
	return false
}

func (e *stepEvents) before(earlier, later string) bool {
	ei, li := -1, -1
	for i, step := range e.seq {
		if step == earlier && ei < 0 {
			ei = i
		}
		if step == later && li < 0 {
			li = i
		}
	}
	return ei >= 0 && li >= 0 && ei < li
}

type stepTx struct{ ev *stepEvents }

func (t stepTx) Begin(context.Context) (*xorm.Session, error) {
	t.ev.add("begin")
	return nil, nil
}

func (t stepTx) Finish(*xorm.Session, *error) {}

type stepPatient struct {
	lockPatientDB
	ev       *stepEvents
	recorded *model.TrxPatientVisit
	detail   *model.DtlPatientVisit
	update   model.UpdatePatientVisitRequest
}

func (p *stepPatient) GetPatients(context.Context, model.GetPatientParams) ([]model.MstPatientInstitution, error) {
	return []model.MstPatientInstitution{{ID: 5}}, nil
}

func (p *stepPatient) RecordPatientVisit(_ context.Context, visit *model.TrxPatientVisit) error {
	p.ev.add("record")
	copied := *visit
	p.recorded = &copied
	visit.ID = testVisitID
	return nil
}

func (p *stepPatient) InsertDtlPatientVisit(_ context.Context, detail *model.DtlPatientVisit) error {
	p.ev.add("detail")
	p.insertedDetail = true
	copied := *detail
	p.detail = &copied
	return nil
}

func (p *stepPatient) UpdatePatientVisit(_ context.Context, req model.UpdatePatientVisitRequest) (model.TrxPatientVisit, error) {
	p.ev.add("update")
	p.updatedVisit = true
	p.update = req
	return model.TrxPatientVisit{}, nil
}

type stepJourney struct {
	journeyrepo.JourneyDB
	ev     *stepEvents
	rooms  []model.PickerRoom
	hidden map[int64]struct{}
	locks  []int64
	lists  int
	fail   error
}

func (j *stepJourney) ListColumnRooms(context.Context, int64, int64) ([]model.PickerRoom, error) {
	j.lists++
	j.ev.add("list")
	return j.rooms, nil
}

func (j *stepJourney) LockRoomShare(_ context.Context, clinicID, roomID int64) (*model.MstServicePoint, error) {
	j.locks = append(j.locks, roomID)
	j.ev.add("lock")
	if j.fail != nil {
		return nil, j.fail
	}
	if _, hidden := j.hidden[roomID]; hidden {
		return nil, fmt.Errorf("hidden room: %w", constant.ErrorNoAffectedRow)
	}
	return &model.MstServicePoint{ID: roomID, IDMstInstitution: clinicID}, nil
}

func (j *stepJourney) GetJourneyPoint(_ context.Context, point model.MstJourneyPoint) (*model.MstJourneyPoint, error) {
	return j.column(point.ShortID), nil
}

func (j *stepJourney) GetJourneyPointByShortID(_ context.Context, shortID string) (*model.MstJourneyPoint, error) {
	return j.column(shortID), nil
}

func (j *stepJourney) GetJourneyBoardByJourneyPoint(context.Context, model.MstJourneyPoint) (model.MstJourneyBoard, error) {
	return model.MstJourneyBoard{ID: 3}, nil
}

func (j *stepJourney) column(shortID string) *model.MstJourneyPoint {
	if shortID == "" {
		shortID = "jp"
	}
	return &model.MstJourneyPoint{ID: 7, Name: "Start", ShortID: shortID}
}

func newStepUC(rooms []int64) (*VisitUC, *stepPatient, *stepJourney, *stepEvents) {
	ev := &stepEvents{}
	picked := make([]model.PickerRoom, len(rooms))
	for i, id := range rooms {
		picked[i] = model.PickerRoom{ID: id, Name: "R"}
	}
	journey := &stepJourney{
		ev:    ev,
		rooms: picked,
		hidden: map[int64]struct{}{
			roomMissing:  {},
			roomForeign:  {},
			roomOrphan:   {},
			roomArchived: {},
		},
	}
	patient := &stepPatient{lockPatientDB: lockPatientDB{visit: openVisitRow()}, ev: ev}
	uc := &VisitUC{PatientDB: patient, JourneyDB: journey, Transaction: stepTx{ev: ev}}
	return uc, patient, journey, ev
}

func assertRoomErr(t *testing.T, err error, code int, name string) {
	t.Helper()
	var em *commonerr.ErrorMessage
	if !errors.As(err, &em) || em == nil || em.Code != code || len(em.ErrorList) != 1 || em.ErrorList[0].ErrorName != name {
		t.Fatalf("error %v", err)
	}
}

func assertNoStepWrite(t *testing.T, patient *stepPatient) {
	t.Helper()
	if patient.recorded != nil || patient.detail != nil || patient.updatedVisit {
		t.Fatalf("write happened record=%v detail=%v updated=%v", patient.recorded != nil, patient.detail != nil, patient.updatedVisit)
	}
}

func assertLocks(t *testing.T, journey *stepJourney, want ...int64) {
	t.Helper()
	if len(journey.locks) != len(want) {
		t.Fatalf("locks %v want %v", journey.locks, want)
	}
	for i := range want {
		if journey.locks[i] != want[i] {
			t.Fatalf("locks %v want %v", journey.locks, want)
		}
	}
}

func assertDetailRoom(t *testing.T, detail *model.DtlPatientVisit, room int64, notes json.RawMessage) {
	t.Helper()
	if detail == nil {
		t.Fatal("missing step")
	}
	if detail.IDMstServicePoint != room || detail.IDMstJourneyPoint != 7 || detail.JourneyPointName != "Start" || detail.IDTrxPatientVisit != testVisitID {
		t.Fatalf("step %+v", detail)
	}
	if string(detail.Notes) != string(notes) {
		t.Fatalf("notes %s want %s", detail.Notes, notes)
	}
}

type stepRoomCase struct {
	name     string
	rooms    []int64
	room     null.Int64
	notes    json.RawMessage
	wantCode int
	wantName string
	wantRoom int64
	lock     int64
}

func stepRoomCases(notes json.RawMessage) []stepRoomCase {
	return []stepRoomCase{
		{name: "nil notes and zero rooms", wantRoom: 0},
		{name: "nil notes autofills the only room", rooms: []int64{roomLinked}, wantRoom: roomLinked, lock: roomLinked},
		{name: "explicit zero autofills the only room", rooms: []int64{roomLinked}, room: null.Int64From(0), wantRoom: roomLinked, lock: roomLinked},
		{name: "notes stay on the first step", rooms: []int64{roomLinked}, notes: notes, wantRoom: roomLinked, lock: roomLinked},
		{name: "given linked room", rooms: []int64{roomLinked, roomSecond}, room: null.Int64From(roomLinked), wantRoom: roomLinked, lock: roomLinked},
		{name: "two rooms require a pick", rooms: []int64{roomLinked, roomSecond}, wantCode: 400, wantName: "room_required"},
		{name: "zero with two rooms requires a pick", rooms: []int64{roomLinked, roomSecond}, room: null.Int64From(0), wantCode: 400, wantName: "room_required"},
		{name: "in clinic room not linked", rooms: []int64{roomSecond}, room: null.Int64From(roomLinked), wantCode: 400, wantName: "room_not_in_column", lock: roomLinked},
		{name: "missing", room: null.Int64From(roomMissing), wantCode: 404, wantName: "404", lock: roomMissing},
		{name: "other clinic", room: null.Int64From(roomForeign), wantCode: 404, wantName: "404", lock: roomForeign},
		{name: "orphan", room: null.Int64From(roomOrphan), wantCode: 404, wantName: "404", lock: roomOrphan},
		{name: "archived", room: null.Int64From(roomArchived), wantCode: 404, wantName: "404", lock: roomArchived},
	}
}

func TestInsertNewVisit_StoresStepRoom(t *testing.T) {
	notes := json.RawMessage(`{"a":1}`)
	for _, tc := range stepRoomCases(notes) {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			uc, patient, journey, ev := newStepUC(tc.rooms)
			err := uc.InsertNewVisit(adminCtx(), model.InsertNewVisitRequest{
				PatientUUID:         "patient",
				JourneyPointShortID: "jp",
				Notes:               tc.notes,
				ServicePointID:      tc.room,
			})
			if tc.wantName != "" {
				assertRoomErr(t, err, tc.wantCode, tc.wantName)
				assertNoStepWrite(t, patient)
				if tc.lock == 0 {
					assertLocks(t, journey)
				} else {
					assertLocks(t, journey, tc.lock)
				}
				if !ev.has("begin") {
					t.Fatalf("rejected create never opened a transaction: %v", ev.seq)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			assertDetailRoom(t, patient.detail, tc.wantRoom, tc.notes)
			if patient.recorded == nil || patient.recorded.IDMstServicePoint != tc.wantRoom || patient.recorded.IDMstJourneyPoint != 7 {
				t.Fatalf("visit %+v", patient.recorded)
			}
			if tc.lock == 0 {
				assertLocks(t, journey)
			} else {
				assertLocks(t, journey, tc.lock)
			}
			if !ev.before("begin", "record") || !ev.before("record", "detail") {
				t.Fatalf("order %v", ev.seq)
			}
			if tc.lock != 0 && !ev.before("lock", "record") {
				t.Fatalf("lock order %v", ev.seq)
			}
		})
	}
}

func TestUpdatePatientVisit_MoveStoresStepRoom(t *testing.T) {
	for _, tc := range stepRoomCases(nil) {
		tc := tc
		if tc.name == "notes stay on the first step" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			uc, patient, journey, ev := newStepUC(tc.rooms)
			err := uc.UpdatePatientVisit(adminCtx(), model.UpdatePatientVisitRequest{
				ID:                     testVisitID,
				ShortIDMstJourneyPoint: null.StringFrom("jp"),
				IDMstServicePoint:      tc.room,
			})
			if patient.updatedDetail {
				t.Fatal("older step changed")
			}
			if tc.wantName != "" {
				assertRoomErr(t, err, tc.wantCode, tc.wantName)
				assertNoStepWrite(t, patient)
				if tc.lock == 0 {
					assertLocks(t, journey)
				} else {
					assertLocks(t, journey, tc.lock)
				}
				if !ev.has("begin") {
					t.Fatalf("rejected move never opened a transaction: %v", ev.seq)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			assertDetailRoom(t, patient.detail, tc.wantRoom, nil)
			if !patient.update.IDMstServicePoint.Valid || patient.update.IDMstServicePoint.Int64 != tc.wantRoom {
				t.Fatalf("current room %+v", patient.update.IDMstServicePoint)
			}
			if !patient.update.IDMstJourneyPoint.Valid || patient.update.IDMstJourneyPoint.Int64 != 7 {
				t.Fatalf("column %+v", patient.update.IDMstJourneyPoint)
			}
			if tc.lock == 0 {
				assertLocks(t, journey)
			} else {
				assertLocks(t, journey, tc.lock)
			}
			if !ev.before("begin", "detail") || !ev.before("begin", "update") {
				t.Fatalf("order %v", ev.seq)
			}
			if tc.lock != 0 && !ev.before("lock", "detail") {
				t.Fatalf("lock order %v", ev.seq)
			}
		})
	}
}

func TestUpdatePatientVisit_StatusDoesNotWriteStep(t *testing.T) {
	uc, patient, journey, ev := newStepUC([]int64{roomLinked, roomSecond})
	err := uc.UpdatePatientVisit(adminCtx(), model.UpdatePatientVisitRequest{
		ID:                testVisitID,
		Status:            null.StringFrom("open"),
		IDMstServicePoint: null.Int64From(roomForeign),
	})
	if err != nil {
		t.Fatal(err)
	}
	if patient.detail != nil || journey.lists != 0 || len(journey.locks) != 0 || ev.has("begin") {
		t.Fatalf("status update entered the move path lists=%d locks=%v seq=%v", journey.lists, journey.locks, ev.seq)
	}
	if !patient.updatedVisit || patient.update.IDMstServicePoint.Valid {
		t.Fatalf("update %+v", patient.update)
	}
}

func TestUpdatePatientVisit_MoveRejectsOtherInstitution(t *testing.T) {
	uc, patient, journey, ev := newStepUC([]int64{roomLinked})
	patient.visit = lockedVisitRow()
	patient.visit.IDMstInstitution = testInstitutionID + 1
	err := uc.UpdatePatientVisit(adminCtx(), model.UpdatePatientVisitRequest{
		ID:                     testVisitID,
		ShortIDMstJourneyPoint: null.StringFrom("jp"),
	})
	if err == nil {
		t.Fatal("expected other-clinic move to fail")
	}
	if patient.detail != nil || patient.updatedVisit || ev.has("begin") || journey.lists != 0 {
		t.Fatalf("other-clinic move wrote history detail=%v updated=%v seq=%v lists=%d", patient.detail != nil, patient.updatedVisit, ev.seq, journey.lists)
	}
}

func TestInsertVisitTouchpoint_RevalidatesRoom(t *testing.T) {
	cases := []stepRoomCase{
		{name: "missing", room: null.Int64From(roomMissing), lock: roomMissing},
		{name: "other clinic", room: null.Int64From(roomForeign), lock: roomForeign},
		{name: "orphan", room: null.Int64From(roomOrphan), lock: roomOrphan},
		{name: "archived", room: null.Int64From(roomArchived), lock: roomArchived},
		{name: "not in column", rooms: []int64{roomSecond}, room: null.Int64From(roomLinked), lock: roomLinked},
		{name: "linked room", rooms: []int64{roomLinked}, room: null.Int64From(roomLinked), wantRoom: roomLinked, lock: roomLinked},
		{name: "omit with many rooms", rooms: []int64{roomLinked, roomSecond}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			uc, patient, journey, ev := newStepUC(tc.rooms)
			_, err := uc.InsertVisitTouchpoint(adminCtx(), model.DtlPatientVisitRequest{
				IDTrxPatientVisit: testVisitID,
				IDMstJourneyPoint: "jp",
				IDMstServicePoint: tc.room,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertDetailRoom(t, patient.detail, tc.wantRoom, nil)
			if patient.updatedVisit || patient.recorded != nil {
				t.Fatal("note save changed the visit pointer")
			}
			if tc.lock == 0 {
				assertLocks(t, journey)
			} else {
				assertLocks(t, journey, tc.lock)
			}
			if !ev.before("begin", "detail") {
				t.Fatalf("order %v", ev.seq)
			}
			if tc.lock != 0 && !ev.before("lock", "detail") {
				t.Fatalf("lock order %v", ev.seq)
			}
		})
	}
}

func TestStepRoom_DatabaseErrorSkipsWrites(t *testing.T) {
	dbDown := errors.New("db down")
	uc, patient, journey, _ := newStepUC([]int64{roomLinked})
	journey.fail = dbDown
	err := uc.InsertNewVisit(adminCtx(), model.InsertNewVisitRequest{
		PatientUUID:         "patient",
		JourneyPointShortID: "jp",
		ServicePointID:      null.Int64From(roomLinked),
	})
	if !errors.Is(err, dbDown) {
		t.Fatalf("create error %v", err)
	}
	assertNoStepWrite(t, patient)

	uc, patient, journey, _ = newStepUC([]int64{roomLinked})
	journey.fail = dbDown
	_, err = uc.InsertVisitTouchpoint(adminCtx(), model.DtlPatientVisitRequest{
		IDTrxPatientVisit: testVisitID,
		IDMstJourneyPoint: "jp",
		IDMstServicePoint: null.Int64From(roomLinked),
	})
	if !errors.Is(err, dbDown) {
		t.Fatalf("note error %v", err)
	}
	assertNoStepWrite(t, patient)
}
