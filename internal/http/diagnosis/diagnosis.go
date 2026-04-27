package diagnosis

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	diagnosisuc "github.com/faisalhardin/medilink/internal/entity/usecase/diagnosis"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var bindingBind = binding.Bind

type DiagnosisHandler struct {
	DiagnosisUC diagnosisuc.DiagnosisUC
}

func New(handler *DiagnosisHandler) *DiagnosisHandler {
	return handler
}

func (h *DiagnosisHandler) GetByVisitID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	rows, ucErr := h.DiagnosisUC.GetByVisitID(ctx, visitID)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, rows)
}

func (h *DiagnosisHandler) Save(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	var req model.SaveDiagnosesRequest
	if err = bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, ucErr := h.DiagnosisUC.Save(ctx, visitID, req)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *DiagnosisHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	diagnosisID, err := parseInt64Param(r, "diagnosis_id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	if ucErr := h.DiagnosisUC.Delete(ctx, visitID, diagnosisID); ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, "ok")
}

func parseInt64Param(r *http.Request, key string) (int64, error) {
	raw := chi.URLParam(r, key)
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, commonerr.SetNewBadRequest("invalid_parameter", key+" must be an integer")
	}
	return v, nil
}
