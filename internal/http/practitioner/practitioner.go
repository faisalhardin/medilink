package practitioner

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	practitioneruc "github.com/faisalhardin/medilink/internal/entity/usecase/practitioner"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
)

var bindingBind = binding.Bind

// PractitionerHandler exposes the GET /v1/doctor/search and GET /v1/nurse/search endpoints.
type PractitionerHandler struct {
	PractitionerUC practitioneruc.PractitionerUC
}

func New(handler *PractitionerHandler) *PractitionerHandler {
	return handler
}

// SearchDoctors handles GET /v1/doctor/search?q=...&limit=...
func (h *PractitionerHandler) SearchDoctors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.DoctorSearchRequest
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.PractitionerUC.SearchDoctors(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

// SearchNurses handles GET /v1/nurse/search?q=...&role=...&limit=...
func (h *PractitionerHandler) SearchNurses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.NurseSearchRequest
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.PractitionerUC.SearchNurses(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}
