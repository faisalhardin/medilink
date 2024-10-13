package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/friendsofgo/errors"
)

func (uc *InstitutionUC) InserInstitutionProduct(ctx context.Context, request model.InsertInstitutionProductRequest) (err error) {
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

	return nil

}

func (uc *InstitutionUC) FindInstitutionProductByParams(ctx context.Context, params model.FindTrxInstitutionProductDBParams) (products []model.GetInstitutionProductResponse, err error) {
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

	product := model.TrxInstitutionProduct{
		ID:           request.ID,
		Name:         request.Name,
		IDMstProduct: request.IDMstProduct,
		Price:        request.Price,
		IsItem:       request.IsItem,
		IsTreatment:  request.IsTreatment,
	}
	err = uc.InstitutionRepo.UpdateTrxInstitutionProduct(ctx, &product)
	if err != nil {
		return err
	}

	if !request.Quantity.Valid || !request.UnitType.Valid {
		return nil
	}

	productStock := model.DtlInstitutionProductStock{
		IDTrxInstitutionProduct: product.ID,
		Quantity:                request.Quantity.Int64,
		UnitType:                request.UnitType.String,
	}

	err = uc.InstitutionRepo.UpdateDtlInstitutionProductStock(ctx, &productStock)
	if err != nil {
		return err
	}

	return nil
}

func (uc *InstitutionUC) UpdateInstitutionProductStock(ctx context.Context, product model.DtlInstitutionProductStock) (err error) {
	return uc.InstitutionRepo.UpdateDtlInstitutionProductStock(ctx, &product)
}
