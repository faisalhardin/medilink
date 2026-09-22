package compensation

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

// WorksheetHandler implements the HTTP handler for worksheets.
type WorksheetHandler struct {
	WorksheetUC         compensationuc.WorksheetUC
	VisitCommissionUC   compensationuc.VisitCommissionUC
}

func NewWorksheetHandler(h *WorksheetHandler) *WorksheetHandler {
	return h
}

func (h *WorksheetHandler) CreateWorksheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.CreateWorksheetRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WorksheetUC.CreateWorksheet(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WorksheetHandler) ListWorksheets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.ListWorksheetsRequest{}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WorksheetUC.ListWorksheets(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WorksheetHandler) GetWorksheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "id")
	resp, err := h.WorksheetUC.GetWorksheet(ctx, uuid)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WorksheetHandler) PatchWorksheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.PatchWorksheetRequest{UUID: chi.URLParam(r, "id")}
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	resp, err := h.WorksheetUC.PatchWorksheet(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WorksheetHandler) DeleteWorksheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "id")
	resp, err := h.WorksheetUC.DeleteWorksheet(ctx, uuid)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WorksheetHandler) FinalizeWorksheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "id")
	resp, err := h.WorksheetUC.FinalizeWorksheet(ctx, uuid)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *WorksheetHandler) ListWorksheetCommissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.ListVisitCommissionsRequest{
		WorksheetUUID: chi.URLParam(r, "id"),
	}
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

func (h *WorksheetHandler) GenerateWorksheetCommissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := model.GenerateVisitCommissionsRequest{
		WorksheetID:     chi.URLParam(r, "id"),
		IncludeGenerate: true,
	}
	resp, err := h.VisitCommissionUC.GenerateVisitCommissions(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	_, _ = commonwriter.WriteJSONAPIData(w, nil, http.StatusAccepted, resp)
}
