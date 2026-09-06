package compensation

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationuc "github.com/faisalhardin/medilink/internal/entity/usecase/compensation"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
)

type CompensationPeriodHandler struct {
	CompensationPeriodUC compensationuc.CompensationPeriodUC
}

func New(handler *CompensationPeriodHandler) *CompensationPeriodHandler {
	return handler
}

func (h *CompensationPeriodHandler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.CreateCompensationPeriodRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	response, err := h.CompensationPeriodUC.CreatePeriod(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *CompensationPeriodHandler) ListPeriods(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.ListCompensationPeriodsRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	response, err := h.CompensationPeriodUC.ListPeriods(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *CompensationPeriodHandler) GetPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "periodId")

	response, err := h.CompensationPeriodUC.GetPeriod(ctx, periodID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *CompensationPeriodHandler) DraftPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "periodId")

	response, err := h.CompensationPeriodUC.DraftPeriod(ctx, periodID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *CompensationPeriodHandler) FinalizePeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "periodId")

	response, err := h.CompensationPeriodUC.FinalizePeriod(ctx, periodID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *CompensationPeriodHandler) ReopenPeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "periodId")

	response, err := h.CompensationPeriodUC.ReopenPeriod(ctx, periodID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *CompensationPeriodHandler) DeletePeriod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	periodID := chi.URLParam(r, "periodId")

	response, err := h.CompensationPeriodUC.DeletePeriod(ctx, periodID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}
