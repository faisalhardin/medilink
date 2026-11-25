package journey

import (
	"bytes"
	"context"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

type boardLookup struct {
	journeyRepo.JourneyDB
	gets    int
	inserts int
	board   model.MstJourneyBoard
	getErr  error
}

func (b *boardLookup) GetJourneyBoardByParams(context.Context, model.MstJourneyBoard) (model.MstJourneyBoard, error) {
	b.gets++
	return b.board, b.getErr
}

func (b *boardLookup) InserNewServicePoint(context.Context, *model.MstServicePoint) error {
	b.inserts++
	return nil
}

func TestInsertNewServicePointHidesMissingClinicAndForeignBoard(t *testing.T) {
	for _, clinic := range []int64{0, -1} {
		repo := &boardLookup{}
		uc := &JourneyUC{JourneyDB: repo}
		ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: clinic})
		err := uc.InsertNewServicePoint(ctx, &model.MstServicePoint{IDMstBoard: 3, Name: "room"})
		assertSame404(t, err)
		if repo.gets != 0 || repo.inserts != 0 {
			t.Fatalf("clinic %d gets=%d inserts=%d", clinic, repo.gets, repo.inserts)
		}
	}

	hidden := []struct {
		name   string
		board  model.MstJourneyBoard
		getErr error
	}{
		{name: "other clinic row", board: model.MstJourneyBoard{ID: 3, IDMstInstitution: 9}},
		{name: "not found", getErr: errors.Wrap(constant.ErrorNoAffectedRow, "lookup")},
	}
	for _, tt := range hidden {
		t.Run(tt.name, func(t *testing.T) {
			repo := &boardLookup{board: tt.board, getErr: tt.getErr}
			uc := &JourneyUC{JourneyDB: repo}
			ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: 7})
			err := uc.InsertNewServicePoint(ctx, &model.MstServicePoint{IDMstBoard: 3, Name: "room"})
			assertSame404(t, err)
			if repo.gets != 1 || repo.inserts != 0 {
				t.Fatalf("gets=%d inserts=%d", repo.gets, repo.inserts)
			}
		})
	}

	repo := &boardLookup{board: model.MstJourneyBoard{ID: 3, IDMstInstitution: 7}}
	uc := &JourneyUC{JourneyDB: repo}
	ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: 7})
	if err := uc.InsertNewServicePoint(ctx, &model.MstServicePoint{IDMstBoard: 3, Name: "room"}); err != nil {
		t.Fatal(err)
	}
	if repo.inserts != 1 {
		t.Fatalf("inserts=%d", repo.inserts)
	}
}

func assertSame404(t *testing.T, err error) {
	t.Helper()
	got := httptest.NewRecorder()
	want := httptest.NewRecorder()
	if setErr := commonwriter.SetError(context.Background(), got, err); setErr != nil {
		t.Fatalf("SetError: %v", setErr)
	}
	if setErr := commonwriter.SetError(context.Background(), want, commonerr.Set404()); setErr != nil {
		t.Fatalf("SetError want: %v", setErr)
	}
	if got.Code != want.Code || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
		t.Fatalf("got %d %s want %d %s", got.Code, got.Body.Bytes(), want.Code, want.Body.Bytes())
	}
}

type roomListDB struct {
	journeyRepo.JourneyDB
	listParams model.GetServicePointParams
	listRows   []model.MstServicePoint
	getRow     *model.MstServicePoint
	getErr     error
	jp         map[int64][]int64
}

func (r *roomListDB) ListServicePoints(_ context.Context, params model.GetServicePointParams) ([]model.MstServicePoint, error) {
	r.listParams = params
	return r.listRows, nil
}

func (r *roomListDB) GetServicePoint(_ context.Context, sp model.MstServicePoint) (*model.MstServicePoint, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.getRow, nil
}

func (r *roomListDB) ListJourneyPointIDs(_ context.Context, roomIDs []int64) (map[int64][]int64, error) {
	if r.jp != nil {
		return r.jp, nil
	}
	return map[int64][]int64{}, nil
}

func TestListServicePointsBuildsRoomObject(t *testing.T) {
	repo := &roomListDB{
		listRows: []model.MstServicePoint{{
			ID: 12, Name: "Poli Gigi 1", IDMstBoard: 3, IDMstInstitution: 7,
		}},
		jp: map[int64][]int64{12: {7, 9}},
	}
	uc := &JourneyUC{JourneyDB: repo}
	rooms, err := uc.ListServicePoints(context.Background(), model.GetServicePointParams{IDMstInstitution: 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) != 1 {
		t.Fatalf("len=%d", len(rooms))
	}
	got := rooms[0]
	if got.ID != 12 || got.Name != "Poli Gigi 1" || got.BoardID != 3 || got.IsClinical || got.Location != nil {
		t.Fatalf("room=%#v", got)
	}
	if len(got.JourneyPointIDs) != 2 || got.JourneyPointIDs[0] != 7 || got.JourneyPointIDs[1] != 9 {
		t.Fatalf("journey_point_ids=%v", got.JourneyPointIDs)
	}
}

func TestListServicePointsEmptyClinicNeverLists(t *testing.T) {
	repo := &roomListDB{}
	uc := &JourneyUC{JourneyDB: repo}
	rooms, err := uc.ListServicePoints(context.Background(), model.GetServicePointParams{IDMstInstitution: 0})
	if err != nil {
		t.Fatal(err)
	}
	if rooms == nil || len(rooms) != 0 {
		t.Fatalf("got %#v", rooms)
	}
}

func TestGetServicePointRoomShapeAnd404(t *testing.T) {
	repo := &roomListDB{
		getRow: &model.MstServicePoint{ID: 12, Name: "Poli", IDMstBoard: 3, IDMstInstitution: 7},
	}
	uc := &JourneyUC{JourneyDB: repo}
	room, err := uc.GetServicePoint(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
	if room.ID != 12 || room.IsClinical || room.Location != nil || room.JourneyPointIDs == nil || len(room.JourneyPointIDs) != 0 {
		t.Fatalf("%#v", room)
	}

	repo.getErr = errors.Wrap(constant.ErrorNoAffectedRow, "orphan")
	_, err = uc.GetServicePoint(context.Background(), 99)
	assertSame404(t, err)
}

type roomDB struct {
	journeyRepo.JourneyDB
	rows []model.MstServicePoint
	jp   map[int64][]int64
}

func (r roomDB) ListServicePoints(context.Context, model.GetServicePointParams) ([]model.MstServicePoint, error) {
	return r.rows, nil
}

func (r roomDB) GetServicePoint(_ context.Context, sp model.MstServicePoint) (*model.MstServicePoint, error) {
	for _, row := range r.rows {
		if row.ID == sp.ID {
			return &row, nil
		}
	}
	return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
}

func (r roomDB) ListJourneyPointIDs(context.Context, []int64) (map[int64][]int64, error) {
	return r.jp, nil
}

type fixedLocation struct {
	NoopRoomLocation
	locs map[int64]*model.RoomLocation
}

func (f fixedLocation) Read(context.Context, int64, []int64) (map[int64]*model.RoomLocation, error) {
	return f.locs, nil
}

func TestListServicePointsReturnsRooms(t *testing.T) {
	repo := roomDB{
		rows: []model.MstServicePoint{
			{ID: 12, Name: "Poli Gigi", IDMstBoard: 3, IDMstInstitution: 7},
			{ID: 13, Name: "Kasir", IDMstBoard: 3, IDMstInstitution: 7},
		},
		jp: map[int64][]int64{12: {7, 9}},
	}
	uc := &JourneyUC{JourneyDB: repo}
	got, err := uc.ListServicePoints(context.Background(), model.GetServicePointParams{IDMstInstitution: 7})
	if err != nil {
		t.Fatal(err)
	}
	want := []model.Room{
		{ID: 12, Name: "Poli Gigi", BoardID: 3, JourneyPointIDs: []int64{7, 9}},
		{ID: 13, Name: "Kasir", BoardID: 3, JourneyPointIDs: []int64{}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v\nwant %#v", got, want)
	}
}

func TestListServicePointsCopiesOnlyKnownLocationStatus(t *testing.T) {
	repo := roomDB{rows: []model.MstServicePoint{
		{ID: 1, IDMstInstitution: 7}, {ID: 2, IDMstInstitution: 7}, {ID: 3, IDMstInstitution: 7},
	}}
	synced := &model.RoomLocation{SyncStatus: model.SyncSynced, LastJobType: "create"}
	uc := &JourneyUC{JourneyDB: repo, RoomLocation: fixedLocation{locs: map[int64]*model.RoomLocation{
		1: synced,
		2: {SyncStatus: "queued"},
	}}}
	got, err := uc.ListServicePoints(context.Background(), model.GetServicePointParams{IDMstInstitution: 7})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Location != synced || got[1].Location != nil || got[2].Location != nil {
		t.Fatalf("locations %v %v %v", got[0].Location, got[1].Location, got[2].Location)
	}
}

func TestGetServicePointReturnsRoomOr404(t *testing.T) {
	uc := &JourneyUC{JourneyDB: roomDB{rows: []model.MstServicePoint{{ID: 12, Name: "Poli", IDMstBoard: 3, IDMstInstitution: 7}}}}
	got, err := uc.GetServicePoint(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
	want := &model.Room{ID: 12, Name: "Poli", BoardID: 3, JourneyPointIDs: []int64{}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}

	_, err = uc.GetServicePoint(context.Background(), 99)
	assertSame404(t, err)
}
