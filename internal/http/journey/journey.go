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

func (h *JourneyHandler) GetJourneyPoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	journeyPointIDStr := chi.URLParam(r, "id")
	journeyPointID, err := strconv.ParseInt(journeyPointIDStr, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Journey Point", "Invalid Journey Point ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := &model.GetJourneyPointParams{}
	err = bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.ID = journeyPointID

	resp, err := h.JourneyUC.ListJourneyPoints(ctx, model.GetJourneyPointParams{})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) InsertNewJourneyPoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := &model.MstJourneyPoint{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.InsertNewJourneyPoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, request)
}

func (h *JourneyHandler) UpdateJourneyPoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	journeyIDStr := chi.URLParam(r, "id")
	request := &model.MstJourneyPoint{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.ShortID = journeyIDStr

	err = h.JourneyUC.UpdateJourneyPoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, request)
}

func (h *JourneyHandler) RenameJourneyPoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := &model.RenameJourneyPointRequest{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.UpdateJourneyPoint(ctx, &model.MstJourneyPoint{
		ID:   request.ID,
		Name: request.Name,
	})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, request)
}

func (h *JourneyHandler) ArchiveJourneyPoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := &model.ArchiveJourneyPointRequest{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.ArchiveJourneyPoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *JourneyHandler) GetServicePoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	servicePointIDStr := chi.URLParam(r, "id")
	servicePointID, err := strconv.ParseInt(servicePointIDStr, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Service Point", "Invalid Service Point ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	resp, err := h.JourneyUC.GetServicePoint(ctx, servicePointID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) ListServicePoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetServicePointParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.JourneyUC.ListServicePoints(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *JourneyHandler) InsertNewServicePoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := &model.MstServicePoint{}
	err := bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.JourneyUC.InsertNewServicePoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, request)
}

func (h *JourneyHandler) UpdateServicePoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	servicePointIDStr := chi.URLParam(r, "id")
	servicePointID, err := strconv.ParseInt(servicePointIDStr, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Service Point", "Invalid Service Point ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := &model.MstServicePoint{}
	err = bindingBind(r, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.ID = servicePointID

	err = h.JourneyUC.UpdateServicePoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *JourneyHandler) ArchiveServicePoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	servicePointIDStr := chi.URLParam(r, "id")
	servicePointID, err := strconv.ParseInt(servicePointIDStr, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Service Point", "Invalid Service Point ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	err = h.JourneyUC.ArchiveServicePoint(ctx, &model.MstServicePoint{
		ID: servicePointID,
	})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
