package journey

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	journeyusecase "github.com/faisalhardin/medilink/internal/usecase/journey"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

// visibleRoomDB returns room 12 of clinic 7. Every other id gets the
// ErrorNoAffectedRow that UserJourneyDB and xorm produce for a missing,
// other-clinic, orphan, or archived row.
type visibleRoomDB struct{ journeyRepo.JourneyDB }

func (visibleRoomDB) GetServicePoint(_ context.Context, sp model.MstServicePoint) (*model.MstServicePoint, error) {
	if sp.ID == 12 {
		return &model.MstServicePoint{ID: 12, Name: "Poli", IDMstBoard: 3, IDMstInstitution: 7}, nil
	}
	return nil, errors.Wrap(constant.ErrorNoAffectedRow, "hidden")
}

func (visibleRoomDB) ListJourneyPointIDs(context.Context, []int64) (map[int64][]int64, error) {
	return map[int64][]int64{}, nil
}

func getRoom(id string) *httptest.ResponseRecorder {
	h := &JourneyHandler{JourneyUC: &journeyusecase.JourneyUC{JourneyDB: visibleRoomDB{}}}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req := httptest.NewRequest(http.MethodGet, "/v1/journey/service-point/"+id, nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.GetServicePoint(rec, req)
	return rec
}

func TestGetServicePointReturnsRoomObject(t *testing.T) {
	rec := getRoom("12")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.Bytes())
	}
	want := `{"id":12,"name":"Poli","board_id":3,"is_clinical":false,"journey_point_ids":[],"location":null}`
	if !strings.Contains(rec.Body.String(), want) {
		t.Fatalf("body %s missing %s", rec.Body.Bytes(), want)
	}
}

func TestGetServicePointInvalidIDIs400(t *testing.T) {
	if rec := getRoom("x"); rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.Bytes())
	}
}

func TestGetServicePointHiddenRoomsShareSet404(t *testing.T) {
	want := httptest.NewRecorder()
	if err := commonwriter.SetError(context.Background(), want, commonerr.Set404()); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"99", "100", "101"} {
		rec := getRoom(id)
		if rec.Code != want.Code || !bytes.Equal(rec.Body.Bytes(), want.Body.Bytes()) {
			t.Fatalf("id %s: got %d %s want %d %s", id, rec.Code, rec.Body.Bytes(), want.Code, want.Body.Bytes())
		}
	}
}
