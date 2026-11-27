package journey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyUC "github.com/faisalhardin/medilink/internal/entity/usecase/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/go-chi/chi/v5"
)

type roomArchiveUC struct {
	journeyUC.JourneyUC
	id      int64
	deletes int
	result  model.ArchiveRoomResult
	err     error
}

func (u *roomArchiveUC) ArchiveServicePoint(_ context.Context, id int64) (model.ArchiveRoomResult, error) {
	u.deletes++
	u.id = id
	return u.result, u.err
}

func TestArchiveServicePointHandlerReturnsArchived(t *testing.T) {
	uc := &roomArchiveUC{result: model.ArchiveRoomResult{Archived: true}}
	h := &JourneyHandler{JourneyUC: uc}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	req := httptest.NewRequest(http.MethodDelete, "/v1/journey/service-point/42", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ArchiveServicePoint(rec, req)

	if rec.Code != http.StatusOK || uc.deletes != 1 || uc.id != 42 {
		t.Fatalf("status %d calls %d id %d body %s", rec.Code, uc.deletes, uc.id, rec.Body.Bytes())
	}
	if !strings.Contains(rec.Body.String(), `"archived":true`) {
		t.Fatalf("body %s", rec.Body.Bytes())
	}
}

func TestArchiveServicePointHandlerOpenVisits(t *testing.T) {
	uc := &roomArchiveUC{err: commonerr.RoomHasOpenVisits(2)}
	h := &JourneyHandler{JourneyUC: uc}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "9")
	req := httptest.NewRequest(http.MethodDelete, "/v1/journey/service-point/9", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.ArchiveServicePoint(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
	}
	body := rec.Body.String()
	if !strings.Contains(body, `"room_has_open_visits"`) || !strings.Contains(body, `"count":2`) {
		t.Fatalf("body %s", body)
	}
	for _, leak := range []string{"visit_id", "patient", "nik", "320101"} {
		if strings.Contains(strings.ToLower(body), leak) {
			t.Fatalf("leak %q in %s", leak, body)
		}
	}
}
