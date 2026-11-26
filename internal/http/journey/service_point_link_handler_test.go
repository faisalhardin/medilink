package journey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyUC "github.com/faisalhardin/medilink/internal/entity/usecase/journey"
	"github.com/go-chi/chi/v5"
)

type linkHandlerUC struct {
	journeyUC.JourneyUC
	calls   int
	side    model.LinkKind
	subject int64
	targets []int64
}

func (u *linkHandlerUC) ReplaceLinks(_ context.Context, side model.LinkKind, subject int64, targets []int64) ([]int64, error) {
	u.calls++
	u.side = side
	u.subject = subject
	u.targets = append([]int64{}, targets...)
	return targets, nil
}

func (u *linkHandlerUC) GetRoomJourneyPoints(context.Context, int64) (model.RoomJourneyPoints, error) {
	return model.RoomJourneyPoints{JourneyPointIDs: []int64{7}}, nil
}

func (u *linkHandlerUC) GetColumnRooms(context.Context, int64) (model.JourneyPointRooms, error) {
	return model.JourneyPointRooms{Items: []model.PickerRoom{}}, nil
}

func callLink(method, id, body string, h http.HandlerFunc) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestLinkHandlersRejectBadIDAndMissingKey(t *testing.T) {
	uc := &linkHandlerUC{}
	h := &JourneyHandler{JourneyUC: uc}

	for _, call := range []struct {
		name string
		fn   http.HandlerFunc
	}{
		{"put room", h.PutRoomJourneyPoints},
		{"put column", h.PutColumnRooms},
		{"get room", h.GetRoomJourneyPoints},
		{"get column", h.GetColumnRooms},
	} {
		rec := callLink(http.MethodPut, "x", `{}`, call.fn)
		if rec.Code != http.StatusBadRequest || uc.calls != 0 {
			t.Fatalf("%s bad id: status %d calls %d body %s", call.name, rec.Code, uc.calls, rec.Body.Bytes())
		}
	}

	roomMiss := callLink(http.MethodPut, "12", `{}`, h.PutRoomJourneyPoints)
	columnMiss := callLink(http.MethodPut, "9", `{}`, h.PutColumnRooms)
	if roomMiss.Code != http.StatusBadRequest || columnMiss.Code != http.StatusBadRequest || uc.calls != 0 {
		t.Fatalf("missing key room %d column %d calls %d", roomMiss.Code, columnMiss.Code, uc.calls)
	}

	roomClear := callLink(http.MethodPut, "12", `{"journey_point_ids":[]}`, h.PutRoomJourneyPoints)
	if roomClear.Code != http.StatusOK || uc.calls != 1 || uc.side != model.LinkRoom || uc.subject != 12 || uc.targets == nil || len(uc.targets) != 0 {
		t.Fatalf("room clear status %d calls %d side %d subject %d targets %#v body %s", roomClear.Code, uc.calls, uc.side, uc.subject, uc.targets, roomClear.Body.Bytes())
	}

	columnClear := callLink(http.MethodPut, "9", `{"service_point_ids":[]}`, h.PutColumnRooms)
	if columnClear.Code != http.StatusOK || uc.calls != 2 || uc.side != model.LinkColumn || uc.subject != 9 || len(uc.targets) != 0 {
		t.Fatalf("column clear status %d calls %d side %d subject %d body %s", columnClear.Code, uc.calls, uc.side, uc.subject, columnClear.Body.Bytes())
	}

	roomGet := callLink(http.MethodGet, "12", "", h.GetRoomJourneyPoints)
	columnGet := callLink(http.MethodGet, "9", "", h.GetColumnRooms)
	if roomGet.Code != http.StatusOK || !strings.Contains(roomGet.Body.String(), `"journey_point_ids":[7]`) {
		t.Fatalf("room get %d %s", roomGet.Code, roomGet.Body.Bytes())
	}
	if columnGet.Code != http.StatusOK || !strings.Contains(columnGet.Body.String(), `"items":[]`) || !strings.Contains(columnGet.Body.String(), `"picker_required":false`) {
		t.Fatalf("column get %d %s", columnGet.Code, columnGet.Body.Bytes())
	}
}
