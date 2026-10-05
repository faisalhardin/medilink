package compensation

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

// WageHandler implements the HTTP handler for staff wage configuration.
type WageHandler struct {
	WageUC compensationuc.WageUC
}

func NewWageHandler(h *WageHandler) *WageHandler {
	return h
}

func (h *WageHandler) ListWages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.ListStaffWagesRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WageUC.ListWages(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WageHandler) UpsertWage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.UpsertStaffWageRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WageUC.UpsertWage(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WageHandler) DeleteWage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.ParseInt(chi.URLParam(r, "wageId"), 10, 64)
	if err != nil || id <= 0 {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid_parameter", "wageId must be a positive integer"))
		return
	}
	resp, err := h.WageUC.DeleteWage(ctx, id)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}
