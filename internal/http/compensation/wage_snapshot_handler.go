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

// WageSnapshotHandler implements HTTP for generated wage snapshots.
type WageSnapshotHandler struct {
	WageSnapshotUC compensationuc.WageSnapshotUC
}

func NewWageSnapshotHandler(h *WageSnapshotHandler) *WageSnapshotHandler {
	return h
}

func (h *WageSnapshotHandler) Generate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.GenerateStaffWageSnapshotsRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WageSnapshotUC.Generate(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WageSnapshotHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp, err := h.WageSnapshotUC.List(ctx)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WageSnapshotHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid_parameter", "id must be a positive integer"))
		return
	}
	req := model.UpdateStaffWageSnapshotRequest{}
	if err = bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WageSnapshotUC.Update(ctx, id, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WageSnapshotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid_parameter", "id must be a positive integer"))
		return
	}
	resp, err := h.WageSnapshotUC.Delete(ctx, id)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}
