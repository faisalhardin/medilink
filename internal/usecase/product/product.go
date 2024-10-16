package product

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/entity/repo/product"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	WrapMsgPrefix                 = "ProductUC."
	WrapMsgInsertMstProduct       = WrapMsgPrefix + "InsertMstProduct"
	WrapMsgListMstProductByParams = WrapMsgPrefix + "ListMstProductByParams"
	WrapMsgUpdateMstProduct       = WrapMsgPrefix + "UpdateMstProduct"
	WrapMsgDeleteMstProduct       = WrapMsgPrefix + "DeleteMstProduct"
)

type ProductUC struct {
	ProductDB product.ProductDB
}

func NewProductUC(u *ProductUC) *ProductUC {
	return u
}

func (u *ProductUC) InsertMstProduct(ctx context.Context, req model.MstProduct) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	req.AddedBy = userDetail.Email

	err = u.ProductDB.InsertMstProduct(ctx, &req)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertMstProduct)
	}
	return nil
}

func (u *ProductUC) ListMstProductByParams(ctx context.Context, request model.MstProduct) (products []model.MstProduct, err error) {

	products, err = u.ProductDB.FindMstProductByParams(ctx, request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListMstProductByParams)
		return
	}

	return
}

func (u *ProductUC) UpdateMstProduct(ctx context.Context, request model.MstProduct) (err error) {
	err = u.ProductDB.UpdateMstProduct(ctx, &request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateMstProduct)
		return
	}

	return
}

func (u *ProductUC) DeleteMstProduct(ctx context.Context, request model.MstProduct) (err error) {
	err = u.ProductDB.DeleteMstProduct(ctx, &model.MstProduct{
		ID: request.ID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgDeleteMstProduct)
		return
	}

	return
}
