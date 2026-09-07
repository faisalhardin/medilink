package visit

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	visituc "github.com/faisalhardin/medilink/internal/entity/usecase/visit"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
)

type VisitContributorHandler struct {
	VisitContributorUC visituc.VisitContributorUC
}

func New(handler *VisitContributorHandler) *VisitContributorHandler {
	return handler
}

func (h *VisitContributorHandler) ListVisitContributors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid", "Invalid Visit ID"))
		return
	}

	response, err := h.VisitContributorUC.ListVisitContributors(ctx, visitID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *VisitContributorHandler) AddVisitContributor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	visitID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		commonwriter.SetError(ctx, w, commonerr.SetNewBadRequest("invalid", "Invalid Visit ID"))
		return
	}

	request := model.AddVisitContributorRequest{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	response, err := h.VisitContributorUC.AddVisitContributor(ctx, visitID, request.StaffID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}
