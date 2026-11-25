package journey

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

func pathInt64(r *http.Request, label string) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, commonerr.SetNewBadRequest(label, "Invalid "+label)
	}
	return id, nil
}

func (h *JourneyHandler) GetRoomJourneyPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := pathInt64(r, "Service Point")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.JourneyUC.GetRoomJourneyPoints(ctx, id)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) PutRoomJourneyPoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := pathInt64(r, "Service Point")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	body := model.PutRoomJourneyPointsRequest{}
	if err = bindingBind(r, &body); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	if body.JourneyPointIDs == nil {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("journey_point_ids", "journey_point_ids is required"))
		return
	}

	ids, err := h.JourneyUC.ReplaceLinks(ctx, model.LinkRoom, id, *body.JourneyPointIDs)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	commonwriter.SetOKWithData(ctx, w, model.RoomJourneyPoints{JourneyPointIDs: ids})
}

func (h *JourneyHandler) GetColumnRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := pathInt64(r, "Journey Point")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.JourneyUC.GetColumnRooms(ctx, id)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) PutColumnRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := pathInt64(r, "Journey Point")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	body := model.PutJourneyPointRoomsRequest{}
	if err = bindingBind(r, &body); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	if body.ServicePointIDs == nil {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("service_point_ids", "service_point_ids is required"))
		return
	}

	ids, err := h.JourneyUC.ReplaceLinks(ctx, model.LinkColumn, id, *body.ServicePointIDs)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	if ids == nil {
		ids = []int64{}
	}
	commonwriter.SetOKWithData(ctx, w, struct {
		ServicePointIDs []int64 `json:"service_point_ids"`
	}{ServicePointIDs: ids})
}
