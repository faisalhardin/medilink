package journey

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	liblog "github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

type linkTx struct{}

func (linkTx) Begin(context.Context) (*xorm.Session, error) { return nil, nil }
func (linkTx) Finish(*xorm.Session, *error)                 {}

type linkDB struct {
	journeyRepo.JourneyDB
	rooms       map[int64]model.LinkNode
	columns     map[int64]model.LinkNode
	saved       []model.LinkSet
	locked      []int64
	columnRooms map[int64][]model.PickerRoom
	listRooms   int
	row         *model.MstServicePoint
	jp          map[int64][]int64
	getErr      error
}

func (d *linkDB) ListLinkNodes(_ context.Context, query model.LinkQuery) ([]model.LinkNode, error) {
	src := d.columns
	if query.Kind == model.LinkRoom {
		src = d.rooms
	}
	out := []model.LinkNode{}
	for _, id := range query.IDs {
		if node, ok := src[id]; ok {
			out = append(out, node)
		}
	}
	return out, nil
}

func (d *linkDB) LockJourneyBoard(_ context.Context, boardID int64) error {
	d.locked = append(d.locked, boardID)
	return nil
}

func (d *linkDB) ReplaceLinks(_ context.Context, set model.LinkSet) error {
	d.saved = append(d.saved, set)
	return nil
}

func (d *linkDB) ListColumnRooms(_ context.Context, _, columnID int64) ([]model.PickerRoom, error) {
	d.listRooms++
	rooms := d.columnRooms[columnID]
	if rooms == nil {
		return []model.PickerRoom{}, nil
	}
	return rooms, nil
}

func (d *linkDB) GetServicePoint(_ context.Context, sp model.MstServicePoint) (*model.MstServicePoint, error) {
	if d.getErr != nil {
		return nil, d.getErr
	}
	if d.row == nil || d.row.ID != sp.ID {
		return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
	}
	return d.row, nil
}

func (d *linkDB) ListJourneyPointIDs(context.Context, []int64) (map[int64][]int64, error) {
	if d.jp == nil {
		return map[int64][]int64{}, nil
	}
	return d.jp, nil
}

func linkCtx() context.Context {
	return auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{
		UserID:        4,
		InstitutionID: 7,
	})
}

func newLinkUC(db *linkDB) *JourneyUC {
	return &JourneyUC{JourneyDB: db, Transaction: linkTx{}}
}

func sameBoard() *linkDB {
	return &linkDB{
		rooms: map[int64]model.LinkNode{
			12: {ID: 12, BoardID: 3},
			13: {ID: 13, BoardID: 3},
		},
		columns: map[int64]model.LinkNode{
			7: {ID: 7, BoardID: 3},
			9: {ID: 9, BoardID: 3},
			4: {ID: 4, BoardID: 8},
		},
	}
}

type linkLogs struct {
	lines []string
}

func (l *linkLogs) hook(_ string, msg string, fields logger.KV) {
	name, _ := fields["error_name"].(string)
	l.lines = append(l.lines, msg+" "+name)
}

func (l *linkLogs) reset() { l.lines = nil }

func TestReplaceLinksDedupesEmptyAndRepeat(t *testing.T) {
	logs := &linkLogs{}
	restore := liblog.SetTestLogHook(logs.hook)
	defer restore()

	db := sameBoard()
	uc := newLinkUC(db)
	ctx := linkCtx()

	cases := []struct {
		name    string
		side    model.LinkKind
		subject int64
		targets []int64
		want    []int64
	}{
		{name: "room dedupe", side: model.LinkRoom, subject: 12, targets: []int64{9, 7, 9}, want: []int64{7, 9}},
		{name: "room clear", side: model.LinkRoom, subject: 12, targets: []int64{}, want: []int64{}},
		{name: "column dedupe", side: model.LinkColumn, subject: 9, targets: []int64{13, 12, 13}, want: []int64{12, 13}},
		{name: "column clear", side: model.LinkColumn, subject: 9, targets: []int64{}, want: []int64{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logs.reset()
			before := len(db.saved)
			var first []int64
			for n := 0; n < 2; n++ {
				got, err := uc.ReplaceLinks(ctx, tc.side, tc.subject, tc.targets)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got, tc.want) {
					t.Fatalf("put %d got %v want %v", n+1, got, tc.want)
				}
				if first == nil {
					first = got
				}
				if !reflect.DeepEqual(got, first) {
					t.Fatalf("put %d got %v first %v", n+1, got, first)
				}
			}
			if len(db.saved) != before+2 {
				t.Fatalf("saves %d", len(db.saved)-before)
			}
			for _, set := range db.saved[before:] {
				if set.Side != tc.side || set.Subject != tc.subject || set.Board != 3 || set.Clinic != 7 || !reflect.DeepEqual(set.Targets, tc.want) {
					t.Fatalf("saved %#v", set)
				}
			}
			if len(logs.lines) != 0 {
				t.Fatalf("logs %v", logs.lines)
			}
		})
	}
}

func TestReplaceLinksHiddenBeforeCrossBoard(t *testing.T) {
	logs := &linkLogs{}
	restore := liblog.SetTestLogHook(logs.hook)
	defer restore()

	db := sameBoard()
	db.rooms[14] = model.LinkNode{ID: 14, BoardID: 8}
	uc := newLinkUC(db)
	ctx := linkCtx()

	hidden := []struct {
		name    string
		side    model.LinkKind
		subject int64
		targets []int64
		lock    bool
	}{
		{name: "hidden room", side: model.LinkRoom, subject: 99, targets: []int64{9}},
		{name: "hidden column", side: model.LinkColumn, subject: 99, targets: []int64{12}},
		{name: "hidden column and cross board", side: model.LinkRoom, subject: 12, targets: []int64{9, 77, 4}, lock: true},
		{name: "hidden room and cross board", side: model.LinkColumn, subject: 9, targets: []int64{12, 77, 14}, lock: true},
	}
	for _, tc := range hidden {
		t.Run(tc.name, func(t *testing.T) {
			logs.reset()
			saves := len(db.saved)
			locks := len(db.locked)
			_, err := uc.ReplaceLinks(ctx, tc.side, tc.subject, tc.targets)
			assertSame404(t, err)
			if len(db.saved) != saves {
				t.Fatalf("saved a refused link")
			}
			if !tc.lock && len(db.locked) != locks {
				t.Fatalf("locked a missing subject")
			}
			if len(logs.lines) != 1 || logs.lines[0] != "room.link_column refused 404" {
				t.Fatalf("logs %v", logs.lines)
			}
		})
	}

	cross := []struct {
		name    string
		side    model.LinkKind
		subject int64
		targets []int64
	}{
		{name: "room to other board", side: model.LinkRoom, subject: 12, targets: []int64{4}},
		{name: "column to other board", side: model.LinkColumn, subject: 9, targets: []int64{14}},
	}
	for _, tc := range cross {
		t.Run(tc.name, func(t *testing.T) {
			logs.reset()
			saves := len(db.saved)
			_, err := uc.ReplaceLinks(ctx, tc.side, tc.subject, tc.targets)
			assertErrorName(t, err, 400, "cross_board")
			if len(db.saved) != saves {
				t.Fatalf("saved a cross-board link")
			}
			if len(logs.lines) != 1 || logs.lines[0] != "room.link_column refused cross_board" {
				t.Fatalf("logs %v", logs.lines)
			}
		})
	}
}

func TestGetColumnRoomsPicker(t *testing.T) {
	db := &linkDB{columns: map[int64]model.LinkNode{
		9: {ID: 9, BoardID: 3},
	}}
	uc := newLinkUC(db)

	_, err := uc.GetColumnRooms(context.Background(), 5)
	assertSame404(t, err)
	if db.listRooms != 0 {
		t.Fatalf("listed rooms for a hidden column")
	}

	db.columnRooms = map[int64][]model.PickerRoom{
		9: {
			{ID: 12, Name: "Poli", IsClinical: true},
			{ID: 13, Name: "Kasir", IsClinical: true},
		},
	}
	two, err := uc.GetColumnRooms(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if !two.PickerRequired || len(two.Items) != 2 || two.Items[0].IsClinical || two.Items[1].IsClinical {
		t.Fatalf("%#v", two)
	}

	db.columnRooms[9] = db.columnRooms[9][:1]
	one, err := uc.GetColumnRooms(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if one.PickerRequired || len(one.Items) != 1 || one.Items[0].ID != 12 || one.Items[0].IsClinical {
		t.Fatalf("%#v", one)
	}

	db.columnRooms[9] = nil
	none, err := uc.GetColumnRooms(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if none.PickerRequired || none.Items == nil || len(none.Items) != 0 {
		t.Fatalf("%#v", none)
	}
	raw, err := json.Marshal(none)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"picker_required":false`) || !strings.Contains(string(raw), `"items":[]`) {
		t.Fatalf("body %s", raw)
	}
}

func TestGetRoomJourneyPointsMatchesRoom(t *testing.T) {
	db := &linkDB{
		row: &model.MstServicePoint{ID: 12, Name: "Poli", IDMstBoard: 3, IDMstInstitution: 7},
		jp:  map[int64][]int64{12: {9, 7}},
	}
	uc := &JourneyUC{JourneyDB: db}
	room, err := uc.GetServicePoint(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
	got, err := uc.GetRoomJourneyPoints(context.Background(), 12)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.JourneyPointIDs, room.JourneyPointIDs) {
		t.Fatalf("mapping %v room %v", got.JourneyPointIDs, room.JourneyPointIDs)
	}

	db.getErr = errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
	_, err = uc.GetRoomJourneyPoints(context.Background(), 12)
	assertSame404(t, err)
}

func assertErrorName(t *testing.T, err error, status int, name string) {
	t.Helper()
	rec := httptest.NewRecorder()
	if setErr := commonwriter.SetError(context.Background(), rec, err); setErr != nil {
		t.Fatal(setErr)
	}
	if rec.Code != status || !strings.Contains(rec.Body.String(), `"error_name":"`+name+`"`) {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
	}
}
