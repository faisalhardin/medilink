package anamnesa

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	anamnesauc "github.com/faisalhardin/medilink/internal/entity/usecase/anamnesa"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var bindingBind = binding.Bind

type AnamnesaHandler struct {
	AnamnesaUC anamnesauc.AnamnesaUC
}

func New(handler *AnamnesaHandler) *AnamnesaHandler {
	return handler
}

func (h *AnamnesaHandler) GetByVisitID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, ucErr := h.AnamnesaUC.GetByVisitID(ctx, visitID)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *AnamnesaHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	visitID, err := parseInt64Param(r, "id")
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	var req model.UpsertAnamnesaRequest
	if err = bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, ucErr := h.AnamnesaUC.Upsert(ctx, visitID, req)
	if ucErr != nil {
		commonwriter.SetError(ctx, w, ucErr)
		return
	}
	commonwriter.SetOKWithData(ctx, w, resp)
}

func parseInt64Param(r *http.Request, key string) (int64, error) {
	raw := chi.URLParam(r, key)
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, commonerr.SetNewBadRequest("invalid_parameter", key+" must be an integer")
	}
	return v, nil
}
