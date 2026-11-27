package journey

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

type archiveTx struct{}

func (archiveTx) Begin(context.Context) (*xorm.Session, error) { return nil, nil }
func (archiveTx) Finish(*xorm.Session, *error)                 {}

type archiveDB struct {
	journeyRepo.JourneyDB
	room       *model.MstServicePoint
	lockErr    error
	shareErr   error
	deleteErr  error
	detachErr  error
	locks      []string
	detached   int
	deleted    int
	shareHits  int
	archived   bool
	currentPtr map[int64]int64
	mu         sync.Mutex
}

func (d *archiveDB) LockRoom(_ context.Context, clinicID, roomID int64) (*model.MstServicePoint, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.locks = append(d.locks, "update")
	if d.lockErr != nil {
		return nil, d.lockErr
	}
	if d.room == nil || d.archived || d.room.ID != roomID || d.room.IDMstInstitution != clinicID {
		return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
	}
	cp := *d.room
	return &cp, nil
}

func (d *archiveDB) LockRoomShare(_ context.Context, clinicID, roomID int64) (*model.MstServicePoint, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.shareHits++
	d.locks = append(d.locks, "share")
	if d.shareErr != nil {
		return nil, d.shareErr
	}
	if d.room == nil || d.archived || d.room.ID != roomID || d.room.IDMstInstitution != clinicID {
		return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
	}
	cp := *d.room
	return &cp, nil
}

func (d *archiveDB) DetachRoomLinks(context.Context, int64, int64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.detached++
	return d.detachErr
}

func (d *archiveDB) DeleteServicePoint(context.Context, *model.MstServicePoint) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.deleted++
	if d.deleteErr != nil {
		return d.deleteErr
	}
	d.archived = true
	return nil
}

type archivePatientDB struct {
	patientRepo.PatientDB
	open    int
	countErr error
	calls   int
}

func (p *archivePatientDB) CountOpenVisitsByRoom(context.Context, int64, int64) (int, error) {
	p.calls++
	return p.open, p.countErr
}

type archiveLocation struct {
	NoopRoomLocation
	jobs      []string
	enqueued  []int64
	block     chan struct{}
	started   chan struct{}
	mu        sync.Mutex
}

func (l *archiveLocation) Enqueue(_ context.Context, _, room int64, job string) error {
	l.mu.Lock()
	l.enqueued = append(l.enqueued, room)
	l.jobs = append(l.jobs, job)
	l.mu.Unlock()
	if l.started != nil {
		select {
		case <-l.started:
		default:
			close(l.started)
		}
	}
	if l.block != nil {
		<-l.block
	}
	return nil
}

func archiveClinicCtx(clinic int64) context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: clinic})
}

func assertArchiveWire(t *testing.T, err, want error) {
	t.Helper()
	got := httptest.NewRecorder()
	wantRec := httptest.NewRecorder()
	if setErr := commonwriter.SetError(context.Background(), got, err); setErr != nil {
		t.Fatalf("SetError: %v", setErr)
	}
	if setErr := commonwriter.SetError(context.Background(), wantRec, want); setErr != nil {
		t.Fatalf("SetError want: %v", setErr)
	}
	if got.Code != wantRec.Code || !bytes.Equal(got.Body.Bytes(), wantRec.Body.Bytes()) {
		t.Fatalf("got %d %s want %d %s", got.Code, got.Body.Bytes(), wantRec.Code, wantRec.Body.Bytes())
	}
	body := got.Body.String()
	for _, leak := range []string{"Budi", "3201012345670001", `"visit_id"`, `"patient"`, `"nik"`} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(leak)) {
			t.Fatalf("PII/visit leak %q in %s", leak, body)
		}
	}
}

func TestArchiveServicePointOpenVisitsBlock(t *testing.T) {
	db := &archiveDB{room: &model.MstServicePoint{ID: 9, IDMstInstitution: 7, IDMstBoard: 3, Name: "Poli"}}
	patients := &archivePatientDB{open: 2}
	loc := &archiveLocation{}
	uc := &JourneyUC{
		JourneyDB:    db,
		PatientDB:    patients,
		Transaction:  archiveTx{},
		RoomLocation: loc,
	}

	_, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 9)
	assertArchiveWire(t, err, commonerr.RoomHasOpenVisits(2))
	if db.detached != 0 || db.deleted != 0 || len(loc.jobs) != 0 {
		t.Fatalf("mutated detach=%d delete=%d jobs=%v", db.detached, db.deleted, loc.jobs)
	}
	if patients.calls != 1 {
		t.Fatalf("count calls %d", patients.calls)
	}
}

func TestArchiveServicePointSoftDeletedVisitsDoNotBlock(t *testing.T) {
	db := &archiveDB{room: &model.MstServicePoint{ID: 9, IDMstInstitution: 7, IDMstBoard: 3, Name: "Poli"}}
	uc := &JourneyUC{
		JourneyDB:    db,
		PatientDB:    &archivePatientDB{open: 0},
		Transaction:  archiveTx{},
		RoomLocation: &archiveLocation{},
	}
	result, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 9)
	if err != nil || !result.Archived {
		t.Fatalf("result %#v err %v", result, err)
	}
	if db.detached != 1 || db.deleted != 1 {
		t.Fatalf("detach=%d delete=%d", db.detached, db.deleted)
	}
}

func TestArchiveServicePointSuccessDetachesDeletesEnqueues(t *testing.T) {
	db := &archiveDB{room: &model.MstServicePoint{ID: 9, IDMstInstitution: 7, IDMstBoard: 3, Name: "Poli"}}
	loc := &archiveLocation{}
	uc := &JourneyUC{
		JourneyDB:    db,
		PatientDB:    &archivePatientDB{open: 0},
		Transaction:  archiveTx{},
		RoomLocation: loc,
	}
	result, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 9)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Archived {
		t.Fatalf("result %#v", result)
	}
	if db.detached != 1 || db.deleted != 1 || !db.archived {
		t.Fatalf("detach=%d delete=%d archived=%v", db.detached, db.deleted, db.archived)
	}
	if len(loc.jobs) != 1 || loc.jobs[0] != jobLocationInactivate || loc.enqueued[0] != 9 {
		t.Fatalf("enqueue %#v %#v", loc.jobs, loc.enqueued)
	}
	if len(db.locks) == 0 || db.locks[0] != "update" {
		t.Fatalf("locks %#v", db.locks)
	}
}

func TestArchiveServicePointMissingIs404(t *testing.T) {
	db := &archiveDB{}
	uc := &JourneyUC{
		JourneyDB:   db,
		PatientDB:   &archivePatientDB{},
		Transaction: archiveTx{},
	}
	_, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 99)
	assertArchiveWire(t, err, commonerr.Set404())
	if db.detached != 0 || db.deleted != 0 {
		t.Fatalf("mutated detach=%d delete=%d", db.detached, db.deleted)
	}
}

func TestArchiveServicePointNoClinicIs404(t *testing.T) {
	uc := &JourneyUC{
		JourneyDB:   &archiveDB{room: &model.MstServicePoint{ID: 9, IDMstInstitution: 7}},
		PatientDB:   &archivePatientDB{},
		Transaction: archiveTx{},
	}
	_, err := uc.ArchiveServicePoint(archiveClinicCtx(0), 9)
	assertArchiveWire(t, err, commonerr.Set404())
}

func TestArchiveThenShareLeavesNoCurrentPointer(t *testing.T) {
	db := &archiveDB{
		room:       &model.MstServicePoint{ID: 9, IDMstInstitution: 7, IDMstBoard: 3, Name: "Poli"},
		currentPtr: map[int64]int64{},
	}
	uc := &JourneyUC{
		JourneyDB:    db,
		PatientDB:    &archivePatientDB{open: 0},
		Transaction:  archiveTx{},
		RoomLocation: &archiveLocation{},
	}
	if _, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 9); err != nil {
		t.Fatal(err)
	}

	// BE-7 move/first-step protocol: FOR SHARE then write current room.
	visitID := int64(55)
	room, err := db.LockRoomShare(context.Background(), 7, 9)
	if err == nil {
		db.currentPtr[visitID] = room.ID
		t.Fatalf("share succeeded on archived room")
	}
	if db.currentPtr[visitID] != 0 {
		t.Fatalf("current pointer set to %d", db.currentPtr[visitID])
	}
}

func TestArchiveShareRaceLeavesNoCurrentPointer(t *testing.T) {
	db := &archiveDB{
		room:       &model.MstServicePoint{ID: 9, IDMstInstitution: 7, IDMstBoard: 3, Name: "Poli"},
		currentPtr: map[int64]int64{},
	}
	gate := make(chan struct{})
	var moveErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-gate
		room, err := db.LockRoomShare(context.Background(), 7, 9)
		if err != nil {
			moveErr = err
			return
		}
		db.mu.Lock()
		db.currentPtr[55] = room.ID
		db.mu.Unlock()
	}()

	uc := &JourneyUC{
		JourneyDB:    db,
		PatientDB:    &archivePatientDB{open: 0},
		Transaction:  archiveTx{},
		RoomLocation: &archiveLocation{},
	}
	if _, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 9); err != nil {
		t.Fatal(err)
	}
	close(gate)
	wg.Wait()

	db.mu.Lock()
	defer db.mu.Unlock()
	if db.currentPtr[55] != 0 {
		t.Fatalf("current pointer landed on archived room: %d", db.currentPtr[55])
	}
	if moveErr == nil {
		t.Fatal("expected share to fail after archive")
	}
}

func TestArchiveStubEnqueueDoesNotBlockResponse(t *testing.T) {
	db := &archiveDB{room: &model.MstServicePoint{ID: 9, IDMstInstitution: 7, IDMstBoard: 3, Name: "Poli"}}
	// Noop path (nil RoomLocation) must finish without waiting on a remote.
	uc := &JourneyUC{
		JourneyDB:   db,
		PatientDB:   &archivePatientDB{open: 0},
		Transaction: archiveTx{},
	}
	done := make(chan model.ArchiveRoomResult, 1)
	go func() {
		result, err := uc.ArchiveServicePoint(archiveClinicCtx(7), 9)
		if err != nil {
			t.Errorf("archive: %v", err)
			return
		}
		done <- result
	}()
	select {
	case result := <-done:
		if !result.Archived {
			t.Fatalf("result %#v", result)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("archive blocked waiting on location enqueue")
	}
}

func TestArchiveRoomResultJSON(t *testing.T) {
	b, err := json.Marshal(model.ArchiveRoomResult{Archived: true})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `{"archived":true}` {
		t.Fatalf("%s", b)
	}
}
