package customerjourney

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyIface "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
)

type recordingRoomDB struct {
	journeyIface.JourneyDB
	calls       int
	institution int64
}

func (r *recordingRoomDB) GetServicePoint(_ context.Context, sp model.MstServicePoint) (*model.MstServicePoint, error) {
	r.calls++
	r.institution = sp.IDMstInstitution
	return &sp, nil
}

func (r *recordingRoomDB) UpdateServicePoint(_ context.Context, sp *model.MstServicePoint) error {
	r.calls++
	r.institution = sp.IDMstInstitution
	return nil
}

func (r *recordingRoomDB) DeleteServicePoint(_ context.Context, sp *model.MstServicePoint) error {
	r.calls++
	r.institution = sp.IDMstInstitution
	return nil
}

func (r *recordingRoomDB) InserNewServicePoint(_ context.Context, sp *model.MstServicePoint) error {
	r.calls++
	r.institution = sp.IDMstInstitution
	return nil
}

func (r *recordingRoomDB) ListLinkNodes(_ context.Context, query model.LinkQuery) ([]model.LinkNode, error) {
	r.calls++
	r.institution = query.Clinic
	return []model.LinkNode{}, nil
}

func (r *recordingRoomDB) ListColumnRooms(_ context.Context, clinic, _ int64) ([]model.PickerRoom, error) {
	r.calls++
	r.institution = clinic
	return []model.PickerRoom{}, nil
}

func (r *recordingRoomDB) LockJourneyBoard(context.Context, int64) error {
	r.calls++
	return nil
}

func (r *recordingRoomDB) LockRoom(_ context.Context, clinicID, _ int64) (*model.MstServicePoint, error) {
	r.calls++
	r.institution = clinicID
	return &model.MstServicePoint{ID: 9, IDMstInstitution: clinicID}, nil
}

func (r *recordingRoomDB) LockRoomShare(_ context.Context, clinicID, _ int64) (*model.MstServicePoint, error) {
	r.calls++
	r.institution = clinicID
	return &model.MstServicePoint{ID: 9, IDMstInstitution: clinicID}, nil
}

func (r *recordingRoomDB) DetachRoomLinks(_ context.Context, clinicID, _ int64) error {
	r.calls++
	r.institution = clinicID
	return nil
}

func (r *recordingRoomDB) ReplaceLinks(_ context.Context, set model.LinkSet) error {
	r.calls++
	r.institution = set.Clinic
	return nil
}

func roomCalls(c *UserJourneyDB) map[string]func(context.Context) error {
	return map[string]func(context.Context) error{
		"get": func(ctx context.Context) error {
			_, err := c.GetServicePoint(ctx, model.MstServicePoint{ID: 99})
			return err
		},
		"update": func(ctx context.Context) error {
			return c.UpdateServicePoint(ctx, &model.MstServicePoint{ID: 99, Name: "x"})
		},
		"archive": func(ctx context.Context) error {
			return c.DeleteServicePoint(ctx, &model.MstServicePoint{ID: 99})
		},
	}
}

func roomLockCalls(c *UserJourneyDB) map[string]func(context.Context) error {
	return map[string]func(context.Context) error{
		"lock": func(ctx context.Context) error {
			_, err := c.LockRoom(ctx, 99, 9)
			return err
		},
		"share": func(ctx context.Context) error {
			_, err := c.LockRoomShare(ctx, 99, 9)
			return err
		},
		"detach": func(ctx context.Context) error {
			return c.DetachRoomLinks(ctx, 99, 9)
		},
	}
}

func TestRoomLookupWithoutClinicIs404AndSkipsRepo(t *testing.T) {
	for _, clinic := range []int64{0, -1} {
		repo := &recordingRoomDB{}
		c := &UserJourneyDB{JourneyDB: repo}
		ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: clinic})

		for verb, call := range roomCalls(c) {
			err := call(ctx)

			got := httptest.NewRecorder()
			want := httptest.NewRecorder()
			if setErr := commonwriter.SetError(context.Background(), got, err); setErr != nil {
				t.Fatalf("clinic %d %s: SetError: %v", clinic, verb, setErr)
			}
			if setErr := commonwriter.SetError(context.Background(), want, commonerr.Set404()); setErr != nil {
				t.Fatalf("clinic %d %s: SetError want: %v", clinic, verb, setErr)
			}
			if got.Code != want.Code || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
				t.Fatalf("clinic %d %s: got %d %s want %d %s", clinic, verb, got.Code, got.Body.Bytes(), want.Code, want.Body.Bytes())
			}
		}
		if repo.calls != 0 {
			t.Fatalf("clinic %d reached the repo %d time(s)", clinic, repo.calls)
		}
	}
}

func TestInserNewServicePointRejectsEmptyClinicAndStampsCaller(t *testing.T) {
	for _, clinic := range []int64{0, -1} {
		repo := &recordingRoomDB{}
		c := &UserJourneyDB{JourneyDB: repo}
		ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: clinic})
		err := c.InserNewServicePoint(ctx, &model.MstServicePoint{Name: "room"})

		got := httptest.NewRecorder()
		want := httptest.NewRecorder()
		if setErr := commonwriter.SetError(context.Background(), got, err); setErr != nil {
			t.Fatalf("clinic %d: SetError: %v", clinic, setErr)
		}
		if setErr := commonwriter.SetError(context.Background(), want, commonerr.Set404()); setErr != nil {
			t.Fatalf("clinic %d: SetError want: %v", clinic, setErr)
		}
		if got.Code != want.Code || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
			t.Fatalf("clinic %d: got %d %s want %d %s", clinic, got.Code, got.Body.Bytes(), want.Code, want.Body.Bytes())
		}
		if repo.calls != 0 {
			t.Fatalf("clinic %d reached the repo %d time(s)", clinic, repo.calls)
		}
	}

	repo := &recordingRoomDB{}
	c := &UserJourneyDB{JourneyDB: repo}
	ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: 7})
	room := &model.MstServicePoint{Name: "room"}
	if err := c.InserNewServicePoint(ctx, room); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 1 || repo.institution != 7 || room.IDMstInstitution != 7 {
		t.Fatalf("calls=%d institution=%d room=%d, want 1 and 7", repo.calls, repo.institution, room.IDMstInstitution)
	}
}

func TestLinkReadsUseCallerClinicAndWritesPassThrough(t *testing.T) {
	for _, clinic := range []int64{0, -1} {
		repo := &recordingRoomDB{}
		c := &UserJourneyDB{JourneyDB: repo}
		ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: clinic})
		_, listErr := c.ListLinkNodes(ctx, model.LinkQuery{Kind: model.LinkRoom, IDs: []int64{9}})
		_, roomErr := c.ListColumnRooms(ctx, 99, 4)
		assertWire(t, listErr, commonerr.Set404())
		assertWire(t, roomErr, commonerr.Set404())
		if repo.calls != 0 {
			t.Fatalf("clinic %d reached the repo %d time(s)", clinic, repo.calls)
		}
	}

	repo := &recordingRoomDB{}
	c := &UserJourneyDB{JourneyDB: repo}
	ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: 7})
	if _, err := c.ListLinkNodes(ctx, model.LinkQuery{Kind: model.LinkColumn, Clinic: 1, IDs: []int64{4}}); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 1 || repo.institution != 7 {
		t.Fatalf("list nodes calls=%d institution=%d", repo.calls, repo.institution)
	}
	if _, err := c.ListColumnRooms(ctx, 1, 4); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 2 || repo.institution != 7 {
		t.Fatalf("list rooms calls=%d institution=%d", repo.calls, repo.institution)
	}
	if err := c.LockJourneyBoard(ctx, 3); err != nil {
		t.Fatal(err)
	}
	set := model.LinkSet{Side: model.LinkRoom, Clinic: 4, Board: 3, Subject: 9, Targets: []int64{}}
	if err := c.ReplaceLinks(ctx, set); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 4 || repo.institution != 4 {
		t.Fatalf("calls=%d institution=%d, want the set clinic left unchanged", repo.calls, repo.institution)
	}
}

func assertWire(t *testing.T, err error, wantErr error) {
	t.Helper()
	got := httptest.NewRecorder()
	want := httptest.NewRecorder()
	if setErr := commonwriter.SetError(context.Background(), got, err); setErr != nil {
		t.Fatalf("SetError: %v", setErr)
	}
	if setErr := commonwriter.SetError(context.Background(), want, wantErr); setErr != nil {
		t.Fatalf("SetError want: %v", setErr)
	}
	if got.Code != want.Code || !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
		t.Fatalf("got %d %s want %d %s", got.Code, got.Body.Bytes(), want.Code, want.Body.Bytes())
	}
}

func TestRoomLookupWithClinicStampsRepo(t *testing.T) {
	repo := &recordingRoomDB{}
	c := &UserJourneyDB{JourneyDB: repo}
	ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: 7})

	for verb, call := range roomCalls(c) {
		repo.calls, repo.institution = 0, 0
		if err := call(ctx); err != nil {
			t.Fatalf("%s: %v", verb, err)
		}
		if repo.calls != 1 || repo.institution != 7 {
			t.Fatalf("%s: calls=%d institution=%d, want 1 and 7", verb, repo.calls, repo.institution)
		}
	}
}

func TestRoomLockWrappersWithoutClinicAre404AndSkipRepo(t *testing.T) {
	for _, clinic := range []int64{0, -1} {
		repo := &recordingRoomDB{}
		c := &UserJourneyDB{JourneyDB: repo}
		ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: clinic})

		for verb, call := range roomLockCalls(c) {
			t.Run(verb, func(t *testing.T) {
				assertWire(t, call(ctx), commonerr.Set404())
			})
		}
		if repo.calls != 0 {
			t.Fatalf("clinic %d reached the repo %d time(s)", clinic, repo.calls)
		}
	}
}

func TestRoomLockWrappersStampCallerClinic(t *testing.T) {
	repo := &recordingRoomDB{}
	c := &UserJourneyDB{JourneyDB: repo}
	ctx := auth.SetUserDetailToCtx(context.Background(), model.UserJWTPayload{InstitutionID: 7})

	for verb, call := range roomLockCalls(c) {
		repo.calls, repo.institution = 0, 0
		if err := call(ctx); err != nil {
			t.Fatalf("%s: %v", verb, err)
		}
		// Caller passed clinic 99; wrapper must stamp the JWT clinic (7).
		if repo.calls != 1 || repo.institution != 7 {
			t.Fatalf("%s: calls=%d institution=%d, want 1 and 7", verb, repo.calls, repo.institution)
		}
	}
}
