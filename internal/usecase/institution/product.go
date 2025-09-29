package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/friendsofgo/errors"
)

func (uc *InstitutionUC) InserInstitutionProduct(ctx context.Context, request model.InsertInstitutionProductRequest) (resp model.TrxInstitutionProduct, err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	session, err := uc.Transaction.Begin(ctx)
	defer uc.Transaction.Finish(session, &err)

	product := model.TrxInstitutionProduct{
		Name:             request.Name,
		IDMstProduct:     request.IDMstProduct,
		IDMstInstitution: userDetail.InstitutionID,
		Price:            request.Price,
		IsItem:           request.IsItem,
		IsTreatment:      request.IsTreatment,
	}

	if !request.IsItem {
		request.Quantity = 1
	}

	err = uc.InstitutionRepo.InsertInstitutionProduct(ctx, &product)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInserInstitutionProduct)
		return
	}

	productStock := model.DtlInstitutionProductStock{
		IDTrxInstitutionProduct: product.ID,
		Quantity:                request.Quantity,
		UnitType:                request.UnitType,
	}
	err = uc.InstitutionRepo.InsertInstitutionProductStock(ctx, &productStock)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInserInstitutionProduct)
		return
	}

	return product, nil

}

func (uc *InstitutionUC) FindInstitutionProductByParams(ctx context.Context, params model.FindTrxInstitutionProductParams) (products []model.GetInstitutionProductResponse, err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}
	params.IDMstInstitution = userDetail.InstitutionID
	return uc.InstitutionRepo.FindTrxInstitutionProductJoinStockByParams(ctx, params)
}

func (uc *InstitutionUC) UpdateInstitutionProduct(ctx context.Context, request model.UpdateInstitutionProductRequest) (err error) {

	session, err := uc.Transaction.Begin(ctx)
	defer uc.Transaction.Finish(session, &err)

	_, err = uc.InstitutionRepo.UpdateTrxInstitutionProduct(ctx, &request)
	if err != nil {
		return err
	}

	if !request.UnitType.Valid {
		return nil
	}

	productStock := model.DtlInstitutionProductStock{
		IDTrxInstitutionProduct: request.ID,
		UnitType:                request.UnitType.String,
	}

	err = uc.InstitutionRepo.UpdateDtlInstitutionProduct(ctx, &productStock)
	if err != nil {
		return err
	}

	return nil
}

func (uc *InstitutionUC) UpdateInstitutionProductStock(ctx context.Context, request model.ProductStockResupplyRequest) (err error) {

	// check product ownership
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	productIDs := make([]int64, len(request.Products))
	for _, product := range request.Products {
		productIDs = append(productIDs, product.IDTrxInstitutionProduct)
	}

	productStocks, err := uc.InstitutionRepo.FindTrxInstitutionProductByParams(ctx, model.FindTrxInstitutionProductParams{
		IDMstProducts:    productIDs,
		IDMstInstitution: userDetail.InstitutionID,
	})

	if len(productStocks) == 0 {
		err = commonerr.SetNewBadRequest("product invalid", "One or more products are invalid")
		return
	}

	// update product stock
	session, err := uc.Transaction.Begin(ctx)
	defer uc.Transaction.Finish(session, &err)
	ctx = xorm.SetDBSession(ctx, session)

	for _, product := range request.Products {
		productStock := model.DtlInstitutionProductStock{
			IDTrxInstitutionProduct: product.IDTrxInstitutionProduct,
			Quantity:                product.Quantity,
		}
		err = uc.InstitutionRepo.RestockDtlInstitutionProductStock(ctx, &productStock)
		if err != nil {
			return err
		}
	}

	return
}
