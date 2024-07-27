package patient

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/usecase/patient"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
)

var (
	bindingBind = binding.Bind
)

type PatientHandler struct {
	PatientUC patient.PatientUC
}

func New(handler *PatientHandler) *PatientHandler {
	return handler
}

func (h *PatientHandler) RegisterNewPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.RegisterNewPatientRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.PatientUC.RegisterNewPatient(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
