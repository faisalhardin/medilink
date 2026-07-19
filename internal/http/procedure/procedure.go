package procedure

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	procedureuc "github.com/faisalhardin/medilink/internal/entity/usecase/procedure"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var bindingBind = binding.Bind

type ProcedureHandler struct {
	ProcedureUC procedureuc.ProcedureUC
}

func New(h *ProcedureHandler) *ProcedureHandler {
	return h
}

// SearchICD9CM handles GET /v1/icd9cm/search
func (h *ProcedureHandler) SearchICD9CM(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.ICD9CMSearchRequest
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	results, ucErr := h.ProcedureUC.SearchICD9CM(ctx, req.Query, req.Limit)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, results)
}

// GetByVisitID handles GET /v1/visit/:id/procedure
func (h *ProcedureHandler) GetByVisitID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	rows, ucErr := h.ProcedureUC.GetByVisitID(ctx, visitID)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, rows)
}

// Save handles POST /v1/visit/:id/procedure (atomic replace)
func (h *ProcedureHandler) Save(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	var req model.SaveProceduresRequest
	if err = bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	summary, ucErr := h.ProcedureUC.Save(ctx, visitID, req)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}

	commonwriter.WriteJSON(w, http.StatusOK, model.SaveProceduresResult{
		Message: "Procedures saved successfully",
		Data:    summary,
	})
}

// Delete handles DELETE /v1/visit/:id/procedure/:procedure_id
func (h *ProcedureHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	procedureID, err := parseInt64Param(r, "procedure_id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	if ucErr := h.ProcedureUC.Delete(ctx, visitID, procedureID); ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetPatientHistory handles GET /v1/patient/:uuid/procedure/history
func (h *ProcedureHandler) GetPatientHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid_parameter", "uuid is required"))
		return
	}

	limit := parseIntQuery(r, "limit", 20)
	offset := parseIntQuery(r, "offset", 0)

	histResp, ucErr := h.ProcedureUC.GetPatientHistory(ctx, uuid, limit, offset)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.WriteJSON(w, http.StatusOK, histResp)
}

func parseInt64Param(r *http.Request, key string) (int64, error) {
	raw := chi.URLParam(r, key)
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, commonerr.SetNewBadRequest("invalid_parameter", key+" must be an integer")
	}
	return v, nil
}

func parseIntQuery(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
