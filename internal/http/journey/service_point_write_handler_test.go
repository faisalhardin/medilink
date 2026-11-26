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

type roomWriteUC struct {
	journeyUC.JourneyUC
	created model.CreateRoomRequest
	renamed model.RenameRoomRequest
	id      int64
	posts   int
	patches int
	room    *model.Room
}

func (u *roomWriteUC) InsertNewServicePoint(_ context.Context, req model.CreateRoomRequest) (*model.Room, error) {
	u.posts++
	u.created = req
	return u.room, nil
}

func (u *roomWriteUC) UpdateServicePoint(_ context.Context, id int64, req model.RenameRoomRequest) (*model.Room, error) {
	u.patches++
	u.id = id
	u.renamed = req
	return u.room, nil
}

func TestServicePointWriteHandlersReturnRoom(t *testing.T) {
	uc := &roomWriteUC{room: &model.Room{
		ID: 42, Name: "Poli", BoardID: 3, JourneyPointIDs: []int64{},
	}}
	h := &JourneyHandler{JourneyUC: uc}

	post := httptest.NewRequest(http.MethodPost, "/v1/journey/service-point", strings.NewReader(`{"name":" Poli A ","board_id":3,"is_clinical":true}`))
	post.Header.Set("Content-Type", "application/json")
	postRec := httptest.NewRecorder()
	h.InsertNewServicePoint(postRec, post)
	if postRec.Code != http.StatusCreated || uc.posts != 1 || uc.created.Name != " Poli A " || uc.created.BoardID != 3 || !uc.created.IsClinical {
		t.Fatalf("post status %d calls %d req %#v body %s", postRec.Code, uc.posts, uc.created, postRec.Body.Bytes())
	}
	if !strings.Contains(postRec.Body.String(), `"id":42`) || !strings.Contains(postRec.Body.String(), `"name":"Poli"`) {
		t.Fatalf("post body %s", postRec.Body.Bytes())
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "42")
	patch := httptest.NewRequest(http.MethodPatch, "/v1/journey/service-point/42", strings.NewReader(`{"name":"X","is_clinical":false}`))
	patch.Header.Set("Content-Type", "application/json")
	patch = patch.WithContext(context.WithValue(patch.Context(), chi.RouteCtxKey, rctx))
	patchRec := httptest.NewRecorder()
	h.UpdateServicePoint(patchRec, patch)
	if patchRec.Code != http.StatusOK || uc.patches != 1 || uc.id != 42 || uc.renamed.Name == nil || *uc.renamed.Name != "X" || uc.renamed.IsClinical {
		t.Fatalf("patch status %d calls %d id %d req %#v body %s", patchRec.Code, uc.patches, uc.id, uc.renamed, patchRec.Body.Bytes())
	}
	if !strings.Contains(patchRec.Body.String(), `"name":"Poli"`) {
		t.Fatalf("patch body %s", patchRec.Body.Bytes())
	}

	bad := httptest.NewRequest(http.MethodPatch, "/v1/journey/service-point/x", strings.NewReader(`{"name":"X"}`))
	bad.Header.Set("Content-Type", "application/json")
	badCtx := chi.NewRouteContext()
	badCtx.URLParams.Add("id", "x")
	bad = bad.WithContext(context.WithValue(bad.Context(), chi.RouteCtxKey, badCtx))
	badRec := httptest.NewRecorder()
	h.UpdateServicePoint(badRec, bad)
	if badRec.Code != http.StatusBadRequest || uc.patches != 1 {
		t.Fatalf("bad id status %d calls %d body %s", badRec.Code, uc.patches, badRec.Body.Bytes())
	}
}
