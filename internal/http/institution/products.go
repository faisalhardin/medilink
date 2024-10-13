package institution

import (
	"net/http"
	"strconv"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

func (h *InstitutionHandler) FindInstitutionProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.FindTrxInstitutionProductDBParams{}
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

	err = h.InstitutionUC.InserInstitutionProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *InstitutionHandler) UpdateInstitutionProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	parsedProductID, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Product ID", "Invalid Product ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.UpdateInstitutionProductRequest{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.ID = parsedProductID
	err = h.InstitutionUC.UpdateInstitutionProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *InstitutionHandler) UpdateInstitutionProductStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productID := chi.URLParam(r, "id")
	parsedProductID, err := strconv.ParseInt(productID, 10, 64)
	if err != nil {
		errMsg := commonerr.SetNewBadRequest("Product ID", "Invalid Product ID")
		commonwriter.SetError(ctx, w, errMsg)
		return
	}

	request := model.DtlInstitutionProductStock{}
	err = bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	request.IDTrxInstitutionProduct = parsedProductID
	err = h.InstitutionUC.UpdateInstitutionProductStock(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
