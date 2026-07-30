package patient

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/usecase/patient"
	"github.com/faisalhardin/medilink/internal/entity/usecase/visit"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
)

type PatientHandler struct {
	PatientUC patient.PatientUC
	VisitUC   visit.VisitUC
}

func New(handler *PatientHandler) *PatientHandler {
	return handler
}

func (h *PatientHandler) RegisterNewPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("idempotency_key_required", "Idempotency-Key header is required"))
		return
	}

	request := model.RegisterNewPatientRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	newPatientResponse, err := h.PatientUC.RegisterNewPatient(ctx, request, idempotencyKey)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, newPatientResponse)
}

func (h *PatientHandler) ListPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetPatientParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	patients, err := h.PatientUC.ListPatients(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, patients)
}

func (h *PatientHandler) GetPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	patientUUID := chi.URLParam(r, "uuid")
	patients, err := h.PatientUC.GetPatients(ctx, patientUUID)
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
