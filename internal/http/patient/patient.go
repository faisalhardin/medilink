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

func (h *PatientHandler) GetPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetPatientParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	patients, err := h.PatientUC.GetPatients(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, patients)
}

func (h *PatientHandler) UpdatePatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.UpdatePatientRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.PatientUC.UpdatePatient(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
