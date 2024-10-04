package institution

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/usecase/institution"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
)

var (
	bindingBind = binding.Bind
)

type InstitutionHandler struct {
	InstitutionUC institution.InstitutionUC
}

func New(handler *InstitutionHandler) *InstitutionHandler {
	return handler
}

func (h *InstitutionHandler) InsertNewInstitution(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.CreateInstitutionRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.InstitutionUC.InsertInstitution(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *InstitutionHandler) FindInstitutions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.FindInstitutionParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	institutions, err := h.InstitutionUC.FindInstitutionByParams(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, institutions)
}
