package patient

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/usecase/visit"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

type errVisitUC struct {
	visit.VisitUC
	err error
}

func (e errVisitUC) InsertNewVisit(context.Context, model.InsertNewVisitRequest) error {
	return e.err
}

func (e errVisitUC) UpdatePatientVisit(context.Context, model.UpdatePatientVisitRequest) error {
	return e.err
}

func callCreate(h *PatientHandler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/v1/patient/visit", strings.NewReader(`{"patient_uuid":"p","journey_point_id":"jp","service_point_id":900}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.InsertNewVisit(rec, req)
	return rec
}

func callMove(h *PatientHandler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/v1/patient/visit/100", strings.NewReader(`{"journey_point_id":"jp","service_point_id":900}`))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "100")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	h.UpdatePatientVisit(rec, req)
	return rec
}

func TestVisitRoomHandlerErrorsShareWire(t *testing.T) {
	notFound := httptest.NewRecorder()
	if err := commonwriter.SetError(context.Background(), notFound, commonerr.Set404()); err != nil {
		t.Fatal(err)
	}

	rows := []struct {
		name string
		err  error
		miss bool
	}{
		{name: "missing", err: commonerr.Set404(), miss: true},
		{name: "other_clinic", err: commonerr.Set404(), miss: true},
		{name: "orphan", err: commonerr.Set404(), miss: true},
		{name: "archived", err: commonerr.Set404(), miss: true},
		{name: "room_not_in_column", err: commonerr.RoomNotInColumn()},
		{name: "room_required", err: commonerr.RoomRequired()},
	}

	var hidden [][]byte
	for _, row := range rows {
		row := row
		for _, endpoint := range []string{"create", "move"} {
			t.Run(endpoint+"/"+row.name, func(t *testing.T) {
				h := &PatientHandler{VisitUC: errVisitUC{err: row.err}}
				rec := callMove(h)
				if endpoint == "create" {
					rec = callCreate(h)
				}
				want := httptest.NewRecorder()
				if err := commonwriter.SetError(context.Background(), want, row.err); err != nil {
					t.Fatal(err)
				}
				if rec.Code != want.Code || !bytes.Equal(rec.Body.Bytes(), want.Body.Bytes()) {
					t.Fatalf("got %d %s want %d %s", rec.Code, rec.Body.Bytes(), want.Code, want.Body.Bytes())
				}
				if row.miss {
					if rec.Code != notFound.Code || !bytes.Equal(rec.Body.Bytes(), notFound.Body.Bytes()) {
						t.Fatalf("body %s", rec.Body.Bytes())
					}
					hidden = append(hidden, append([]byte(nil), rec.Body.Bytes()...))
					return
				}
				if bytes.Equal(rec.Body.Bytes(), notFound.Body.Bytes()) {
					t.Fatal("400 body matched 404")
				}
			})
		}
	}
	if len(hidden) != 8 {
		t.Fatalf("404 fixtures %d", len(hidden))
	}
	for i := 1; i < len(hidden); i++ {
		if !bytes.Equal(hidden[0], hidden[i]) {
			t.Fatal("404 bodies differ")
		}
	}
}
