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
