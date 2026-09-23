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

// VisitCommissionHandler implements HTTP handlers for visit commissions.
type VisitCommissionHandler struct {
	VisitCommissionUC compensationuc.VisitCommissionUC
}

func NewVisitCommissionHandler(h *VisitCommissionHandler) *VisitCommissionHandler {
	return h
}

func (h *VisitCommissionHandler) GenerateVisitCommissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.GenerateVisitCommissionsRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.VisitCommissionUC.GenerateVisitCommissions(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	status := http.StatusOK
	if req.IncludeGenerate {
		status = http.StatusAccepted
	}
	_, _ = commonwriter.WriteJSONAPIData(w, nil, status, resp)
}

func (h *VisitCommissionHandler) ListVisitCommissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.ListVisitCommissionsRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.VisitCommissionUC.ListVisitCommissions(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *VisitCommissionHandler) PatchCommissionItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid_parameter", "id must be a positive integer"))
		return
	}

	req := model.PatchCommissionItemRequest{ID: id}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.VisitCommissionUC.PatchCommissionItem(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *VisitCommissionHandler) ArchiveVisitCommission(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid_parameter", "id must be a positive integer"))
		return
	}

	resp, err := h.VisitCommissionUC.ArchiveVisitCommission(ctx, id)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}
