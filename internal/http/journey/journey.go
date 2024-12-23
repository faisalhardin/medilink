package journey

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyUC "github.com/faisalhardin/medilink/internal/entity/usecase/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
)

type JourneyHandler struct {
	JourneyUC journeyUC.JourneyUC
}

func New(handler *JourneyHandler) *JourneyHandler {
	return handler
}

func (h *JourneyHandler) InsertNewJourneyBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.MstJourneyBoard{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.InsertNewJourneyBoard(ctx, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *JourneyHandler) ListJourneyBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetJourneyBoardParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.JourneyUC.ListJourneyBoard(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) GetJourneyBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	boardIDStr := chi.URLParam(r, "id")
	boardID, err := strconv.ParseInt(boardIDStr, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Board ID", "Invalid Board ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	resp, err := h.JourneyUC.GetJourneyBoardDetail(ctx, model.GetJourneyBoardParams{
		ID: []int64{boardID},
	})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) UpdateJourneyBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := &model.MstJourneyBoard{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.UpdateJourneyBoard(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *JourneyHandler) DeleteJourneyBoard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := &model.MstJourneyBoard{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.DeleteJourneyBoard(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
