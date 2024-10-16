package product

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/usecase/product"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
)

var (
	bindingBind = binding.Bind
)

type ProductHandler struct {
	ProductUC product.ProductUC
}

func New(handler *ProductHandler) *ProductHandler {
	return handler
}

func (h *ProductHandler) InsertMstProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.MstProduct{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.ProductUC.InsertMstProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *ProductHandler) ListMstProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.MstProduct{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.ProductUC.ListMstProductByParams(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *ProductHandler) UpdateMstProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.MstProduct{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.ProductUC.UpdateMstProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}

func (h *ProductHandler) DeleteMstProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.MstProduct{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.ProductUC.DeleteMstProduct(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, "ok")
}
