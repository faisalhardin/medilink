package icd10

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	icd10uc "github.com/faisalhardin/medilink/internal/entity/usecase/icd10"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
)

var bindingBind = binding.Bind

// ICD10Handler exposes the GET /v1/icd10/search endpoint.
type ICD10Handler struct {
	ICD10UC icd10uc.ICD10UC
}

func New(handler *ICD10Handler) *ICD10Handler {
	return handler
}

// Search handles GET /v1/icd10/search?q=...&limit=...
func (h *ICD10Handler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.ICD10SearchRequest
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.ICD10UC.Search(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}
