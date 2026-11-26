package journey

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

const newRoomID int64 = 42

type writeDB struct {
	journeyRepo.JourneyDB
	board     model.MstJourneyBoard
	boardErr  error
	boardGets int
	row       *model.MstServicePoint
	getErr    error
	gets      int
	insertErr error
	inserts   int
	inserted  *model.MstServicePoint
	updateErr error
	updates   int
	updated   *model.MstServicePoint
}

func (w *writeDB) GetJourneyBoardByParams(context.Context, model.MstJourneyBoard) (model.MstJourneyBoard, error) {
	w.boardGets++
	return w.board, w.boardErr
}

func (w *writeDB) InserNewServicePoint(_ context.Context, sp *model.MstServicePoint) error {
	w.inserts++
	cp := *sp
	w.inserted = &cp
	if w.insertErr != nil {
		return w.insertErr
	}
	sp.ID = newRoomID
	return nil
}

func (w *writeDB) UpdateServicePoint(_ context.Context, sp *model.MstServicePoint) error {
	w.updates++
	cp := *sp
	w.updated = &cp
	return w.updateErr
}

func (w *writeDB) GetServicePoint(_ context.Context, sp model.MstServicePoint) (*model.MstServicePoint, error) {
	w.gets++
	if w.getErr != nil {
		return nil, w.getErr
	}
	if w.row == nil {
		return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
	}
	return w.row, nil
}

func (w *writeDB) ListJourneyPointIDs(context.Context, []int64) (map[int64][]int64, error) {
	return map[int64][]int64{}, nil
}

type lockedLocation struct {
	NoopRoomLocation
	locked   bool
	enqueued []int64
	jobs     []string
}

func (l *lockedLocation) Clearing(context.Context, int64, int64) (bool, error) {
	return l.locked, nil
}

func (l *lockedLocation) Enqueue(_ context.Context, _, room int64, job string) error {
	l.enqueued = append(l.enqueued, room)
	l.jobs = append(l.jobs, job)
	return nil
}

func clinicPtr(id int64) *int64 { return &id }

func clinicContext(id *int64) context.Context {
	if id == nil {
		return context.Background()
	}
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: *id})
}

func assertWire(t *testing.T, err, want error) {
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
}

func assertNot409(t *testing.T, err error) {
	t.Helper()
	rec := httptest.NewRecorder()
	if setErr := commonwriter.SetError(context.Background(), rec, err); setErr != nil {
		t.Fatal(setErr)
	}
	if rec.Code == 409 {
		t.Fatalf("status 409 body %s", rec.Body.Bytes())
	}
}

func sameIDs(got, want []int64) bool {
	if len(got) == 0 && len(want) == 0 {
		return true
	}
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func TestInsertNewServicePointWrite(t *testing.T) {
	taken := errors.Wrap(&pgconn.PgError{Code: "23505", ConstraintName: "ux_service_point_institution_name"}, "insert")
	takenQualified := errors.Wrap(&pgconn.PgError{Code: "23505", ConstraintName: "public.ux_service_point_institution_name"}, "insert")
	otherUnique := errors.Wrap(&pgconn.PgError{Code: "23505", ConstraintName: "ux_other"}, "insert")
	owned := model.MstJourneyBoard{ID: 3, IDMstInstitution: 7}
	clinic := clinicPtr(7)

	cases := []struct {
		name      string
		clinic    *int64
		req       model.CreateRoomRequest
		board     model.MstJourneyBoard
		boardErr  error
		insertErr error
		loc       *lockedLocation
		wantErr   error
		reject409 bool
		boardGets int
		inserted  bool
		roomName  string
		inst      int64
		queue     []int64
	}{
		{name: "blank", clinic: clinic, req: model.CreateRoomRequest{Name: "", BoardID: 3}, wantErr: commonerr.RoomNameRequired()},
		{name: "spaces", clinic: clinic, req: model.CreateRoomRequest{Name: "   ", BoardID: 3}, wantErr: commonerr.RoomNameRequired()},
		{name: "trim and stamp", clinic: clinic, req: model.CreateRoomRequest{Name: " Poli A ", BoardID: 3}, board: owned, boardGets: 1, inserted: true, roomName: "Poli A", inst: 7},
		{name: "foreign board", clinic: clinic, req: model.CreateRoomRequest{Name: "room", BoardID: 3}, board: model.MstJourneyBoard{ID: 3, IDMstInstitution: 9}, boardGets: 1, wantErr: commonerr.Set404()},
		{name: "missing board", clinic: clinic, req: model.CreateRoomRequest{Name: "room", BoardID: 3}, boardErr: errors.Wrap(constant.ErrorNoAffectedRow, "lookup"), boardGets: 1, wantErr: commonerr.Set404()},
		{name: "clinic 0", clinic: clinicPtr(0), req: model.CreateRoomRequest{Name: "room", BoardID: 3}, wantErr: commonerr.Set404()},
		{name: "no session", req: model.CreateRoomRequest{Name: "room", BoardID: 3}, wantErr: commonerr.SetNewUnauthorizedAPICall()},
		{name: "23505 name taken", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3, IsClinical: true}, board: owned, insertErr: taken, loc: &lockedLocation{}, boardGets: 1, inserted: true, inst: 7, wantErr: commonerr.RoomNameTaken()},
		{name: "23505 qualified constraint", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3, IsClinical: true}, board: owned, insertErr: takenQualified, loc: &lockedLocation{}, boardGets: 1, inserted: true, inst: 7, wantErr: commonerr.RoomNameTaken()},
		{name: "23505 other constraint", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3, IsClinical: true}, board: owned, insertErr: otherUnique, loc: &lockedLocation{}, boardGets: 1, inserted: true, inst: 7, reject409: true},
		{name: "clinical noop", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3, IsClinical: true}, board: owned, boardGets: 1, inserted: true, roomName: "Poli", inst: 7},
		{name: "clinical locked", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3, IsClinical: true}, board: owned, loc: &lockedLocation{locked: true}, boardGets: 1, wantErr: commonerr.RoomClearing()},
		{name: "clinical unlocked", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3, IsClinical: true}, board: owned, loc: &lockedLocation{}, boardGets: 1, inserted: true, roomName: "Poli", inst: 7, queue: []int64{newRoomID}},
		{name: "clinical off ignores lock", clinic: clinic, req: model.CreateRoomRequest{Name: "Poli", BoardID: 3}, board: owned, loc: &lockedLocation{locked: true}, boardGets: 1, inserted: true, roomName: "Poli", inst: 7},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			db := &writeDB{board: tt.board, boardErr: tt.boardErr, insertErr: tt.insertErr}
			uc := &JourneyUC{JourneyDB: db}
			if tt.loc != nil {
				uc.RoomLocation = tt.loc
			}
			room, err := uc.InsertNewServicePoint(clinicContext(tt.clinic), tt.req)
			if tt.reject409 {
				if err == nil {
					t.Fatal("expected error")
				}
				assertNot409(t, err)
			} else if tt.wantErr != nil {
				assertWire(t, err, tt.wantErr)
				if room != nil {
					t.Fatalf("room %#v", room)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			inserts := 0
			if tt.inserted {
				inserts = 1
			}
			if db.boardGets != tt.boardGets || db.inserts != inserts {
				t.Fatalf("boardGets=%d inserts=%d", db.boardGets, db.inserts)
			}
			if tt.inserted {
				if db.inserted == nil || db.inserted.IDMstInstitution != tt.inst || db.inserted.IDMstBoard != tt.req.BoardID {
					t.Fatalf("stored %#v", db.inserted)
				}
				if tt.roomName != "" && db.inserted.Name != tt.roomName {
					t.Fatalf("stored name %q", db.inserted.Name)
				}
			}
			if tt.wantErr == nil && !tt.reject409 {
				if room == nil || room.Name != tt.roomName || room.ID != newRoomID || room.BoardID != tt.req.BoardID || room.IsClinical {
					t.Fatalf("room %#v", room)
				}
			}
			if tt.loc != nil && !sameIDs(tt.loc.enqueued, tt.queue) {
				t.Fatalf("enqueued %v want %v", tt.loc.enqueued, tt.queue)
			}
			if len(tt.queue) > 0 && (len(tt.loc.jobs) != 1 || tt.loc.jobs[0] != "location.upsert") {
				t.Fatalf("jobs %v", tt.loc.jobs)
			}
		})
	}
}

func TestUpdateServicePointWrite(t *testing.T) {
	taken := errors.Wrap(&pgconn.PgError{Code: "23505", ConstraintName: "ux_service_point_institution_name"}, "update")
	otherUnique := errors.Wrap(&pgconn.PgError{Code: "23505", ConstraintName: "ux_other"}, "update")
	hidden := errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
	clinic := clinicPtr(7)
	row := &model.MstServicePoint{ID: 12, Name: "Poli", IDMstBoard: 3, IDMstInstitution: 7}
	blank := "   "
	empty := ""
	spaced := " X "

	cases := []struct {
		name      string
		clinic    *int64
		id        int64
		req       model.RenameRoomRequest
		row       *model.MstServicePoint
		getErr    error
		updErr    error
		loc       *lockedLocation
		wantErr   error
		reject409 bool
		gets      int
		updates   int
		updated   *model.MstServicePoint
		roomName  string
		queue     []int64
	}{
		{name: "blank", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &blank}, wantErr: commonerr.RoomNameRequired()},
		{name: "empty", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &empty}, wantErr: commonerr.RoomNameRequired()},
		{name: "omitted name keeps stored", clinic: clinic, id: 12, req: model.RenameRoomRequest{}, row: row, gets: 1, updates: 1, updated: &model.MstServicePoint{ID: 12, Name: "Poli"}, roomName: "Poli"},
		{name: "hidden room", clinic: clinic, id: 99, req: model.RenameRoomRequest{Name: &spaced}, getErr: hidden, gets: 1, wantErr: commonerr.Set404()},
		{name: "trim", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &spaced}, row: row, gets: 1, updates: 1, updated: &model.MstServicePoint{ID: 12, Name: "X"}, roomName: "X"},
		{name: "23505 name taken", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &spaced, IsClinical: true}, row: row, updErr: taken, loc: &lockedLocation{}, gets: 1, updates: 1, wantErr: commonerr.RoomNameTaken()},
		{name: "23505 other constraint", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &spaced, IsClinical: true}, row: row, updErr: otherUnique, loc: &lockedLocation{}, gets: 1, updates: 1, reject409: true},
		{name: "update misses row", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &spaced}, row: row, updErr: hidden, gets: 1, updates: 1, wantErr: commonerr.Set404()},
		{name: "clinical locked", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &spaced, IsClinical: true}, row: row, loc: &lockedLocation{locked: true}, gets: 1, wantErr: commonerr.RoomClearing()},
		{name: "clinical unlocked", clinic: clinic, id: 12, req: model.RenameRoomRequest{Name: &spaced, IsClinical: true}, row: row, loc: &lockedLocation{}, gets: 1, updates: 1, updated: &model.MstServicePoint{ID: 12, Name: "X"}, roomName: "X", queue: []int64{12}},
		{name: "no session", id: 12, req: model.RenameRoomRequest{Name: &spaced}, wantErr: commonerr.SetNewUnauthorizedAPICall()},
		{name: "clinic 0", clinic: clinicPtr(0), id: 12, req: model.RenameRoomRequest{Name: &spaced}, wantErr: commonerr.Set404()},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var row *model.MstServicePoint
			if tt.row != nil {
				cp := *tt.row
				row = &cp
			}
			db := &writeDB{row: row, getErr: tt.getErr, updateErr: tt.updErr}
			uc := &JourneyUC{JourneyDB: db}
			if tt.loc != nil {
				uc.RoomLocation = tt.loc
			}
			room, err := uc.UpdateServicePoint(clinicContext(tt.clinic), tt.id, tt.req)
			if tt.reject409 {
				if err == nil {
					t.Fatal("expected error")
				}
				assertNot409(t, err)
			} else if tt.wantErr != nil {
				assertWire(t, err, tt.wantErr)
				if room != nil {
					t.Fatalf("room %#v", room)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if db.gets != tt.gets || db.updates != tt.updates {
				t.Fatalf("gets=%d updates=%d", db.gets, db.updates)
			}
			if tt.updated != nil && (db.updated == nil || *db.updated != *tt.updated) {
				t.Fatalf("updated %#v want %#v", db.updated, tt.updated)
			}
			if tt.wantErr == nil && !tt.reject409 {
				if room == nil || room.Name != tt.roomName || room.ID != tt.id || room.IsClinical {
					t.Fatalf("room %#v", room)
				}
			}
			if tt.loc != nil && !sameIDs(tt.loc.enqueued, tt.queue) {
				t.Fatalf("enqueued %v want %v", tt.loc.enqueued, tt.queue)
			}
			if len(tt.queue) > 0 && (len(tt.loc.jobs) != 1 || tt.loc.jobs[0] != "location.upsert") {
				t.Fatalf("jobs %v", tt.loc.jobs)
			}
		})
	}
}
