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
