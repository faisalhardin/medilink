package visit

import (
	"net/http"
	"strconv"

	visituc "github.com/faisalhardin/medilink/internal/entity/usecase/visit"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
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
