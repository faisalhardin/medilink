package recall

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	recalluc "github.com/faisalhardin/medilink/internal/entity/usecase/recall"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var bindingBind = binding.Bind

// RecallHandler handles HTTP requests for recall (scheduled control/appointment) operations
type RecallHandler struct {
	RecallUC recalluc.RecallUC
}

// New creates a new recall HTTP handler
func New(handler *RecallHandler) *RecallHandler {
	return handler
}

// CreateRecall handles POST /recall - schedule a control or future appointment
func (h *RecallHandler) CreateRecall(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.CreateRecallRequest
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.RecallUC.CreateRecall(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

// UpdateRecall handles PATCH /recall/{id} - update scheduled_at, recall_type, and/or notes only
func (h *RecallHandler) UpdateRecall(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req model.UpdateRecallRequest
	if err := bindingBind(r, &req); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err := h.RecallUC.UpdateRecall(ctx, req)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

// GetNextRecallByPatient handles GET /recall/patient/{uuid}/next - next scheduled recall for a patient (doctor reminder)
func (h *RecallHandler) GetNextRecallByPatient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	patientUUID := chi.URLParam(r, "uuid")
	if patientUUID == "" {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("missing patient", "patient_uuid is required"))
		return
	}

	resp, err := h.RecallUC.GetNextRecallByPatient(ctx, patientUUID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

// ListRecalls handles GET /recall - list upcoming recalls for the doctor (optional filters: patient_uuid, from_time, to_time, recall_type)
func (h *RecallHandler) ListRecalls(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var params model.GetRecallParams
	if err := bindingBind(r, &params); err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	list, err := h.RecallUC.ListRecalls(ctx, params)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, list)
}
