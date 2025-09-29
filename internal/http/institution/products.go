package institution

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
)

func (h *InstitutionHandler) FindInstitutionProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.FindTrxInstitutionProductParams{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	products, err := h.InstitutionUC.FindInstitutionProductByParams(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, products)
}

func (h *InstitutionHandler) InsertInstitutionProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.InsertInstitutionProductRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	result, err := h.InstitutionUC.InserInstitutionProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, result)
}

func (h *InstitutionHandler) UpdateInstitutionProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.UpdateInstitutionProductRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.InstitutionUC.UpdateInstitutionProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *InstitutionHandler) UpdateInstitutionProductStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.ProductStockResupplyRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.InstitutionUC.UpdateInstitutionProductStock(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
