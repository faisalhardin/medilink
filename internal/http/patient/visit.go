package patient

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

func (h *PatientHandler) InsertNewVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.InsertNewVisitRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.VisitUC.InsertNewVisit(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *PatientHandler) ListPatientVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetPatientVisitParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	visits, err := h.VisitUC.ListPatientVisits(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, visits)
}

func (h *PatientHandler) ListPatientVisitsByPatientUUID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	patientUUID := chi.URLParam(r, "uuid")
	request := model.GetPatientVisitParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.PatientUUID = patientUUID

	visits, err := h.VisitUC.ListPatientVisits(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, visits)
}

func (h *PatientHandler) GetPatientVisits(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID := chi.URLParam(r, "id")
	parsedVisitID, err := strconv.ParseInt(visitID, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("invalid", "Invalid Visit ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.GetPatientVisitParams{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.IDPatientVisit = parsedVisitID

	visits, err := h.VisitUC.GetPatientVisitDetail(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, visits)
}

func (h *PatientHandler) ListPatientVisitsDetailed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetPatientVisitParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	visits, err := h.VisitUC.ListPatientVisitDetailed(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, visits)
}

func (h *PatientHandler) UpdatePatientVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitIDStr := chi.URLParam(r, "id")
	visitID, err := strconv.ParseInt(visitIDStr, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Visit", "Invalid Visit ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.UpdatePatientVisitRequest{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.ID = visitID

	err = h.VisitUC.UpdatePatientVisit(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *PatientHandler) ArchivePatientVisit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.ArchivePatientVisitRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.VisitUC.ArchivePatientVisit(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *PatientHandler) ListVisitTouchpoints(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID := chi.URLParam(r, "id")
	parsedVisitID, err := strconv.ParseInt(visitID, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("invalid", "Invalid Visit ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.DtlPatientVisitRequest{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.IDTrxPatientVisit = parsedVisitID

	touchpoint, err := h.VisitUC.GetVisitTouchpoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, touchpoint)
}

func (h *PatientHandler) GetVisitTouchpoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitDetailID := chi.URLParam(r, "id")
	parsedVisitDetailID, err := strconv.ParseInt(visitDetailID, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("invalid", "Invalid Visit ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.DtlPatientVisitRequest{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.ID = parsedVisitDetailID

	touchpoint, err := h.VisitUC.GetVisitTouchpoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, touchpoint)
}

func (h *PatientHandler) UpsertVisitTouchpoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.DtlPatientVisitRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.VisitUC.UpsertVisitTouchpoint(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *PatientHandler) InsertVisitProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.UpsertTrxVisitProductRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.VisitUC.UpsertVisitProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *PatientHandler) ListVisitProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.GetVisitProductRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	response, err := h.VisitUC.ListVisitProducts(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *PatientHandler) UpdateVisitProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID := chi.URLParam(r, "id")
	parsedVisitID, err := strconv.ParseInt(visitID, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("invalid", "Invalid Visit ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.InsertTrxVisitProductRequest{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.IDTrxPatientVisit = parsedVisitID

	err = h.VisitUC.UpdateVisitProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
