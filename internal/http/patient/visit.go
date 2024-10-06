package patient

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

func (h *PatientHandler) InsertNewVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	patientUUID := chi.URLParam(r, "id")
	request := model.InsertNewVisitRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	request.PatientUUID = patientUUID
	err = h.VisitUC.InsertNewVisit(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *PatientHandler) GetPatientVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetPatientVisitParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	visits, err := h.VisitUC.GetPatientVisits(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, visits)
}

func (h *PatientHandler) UpdatePatientVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.UpdatePatientVisitRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.VisitUC.UpdatePatientVisit(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
