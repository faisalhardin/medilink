package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type InstitutionDB interface {
	InsertNewInstitution(ctx context.Context, institution *model.Institution) (err error)
	FindInstitutionByParams(ctx context.Context, request model.FindInstitutionParams) (institutions []model.Institution, err error)

	InsertInstitutionProduct(ctx context.Context, product *model.TrxInstitutionProduct) (err error)
	FindTrxInstitutionProductByParams(ctx context.Context, request model.FindTrxInstitutionProductParams) (products []model.TrxInstitutionProduct, err error)
	UpdateTrxInstitutionProduct(ctx context.Context, request *model.TrxInstitutionProduct) (err error)
	InsertInstitutionProductStock(ctx context.Context, product *model.DtlInstitutionProductStock) (err error)
	FindTrxInstitutionProductStockByParams(ctx context.Context, request model.DtlInstitutionProductStock) (stock []model.DtlInstitutionProductStock, err error)
	FindTrxInstitutionProductJoinStockByParams(ctx context.Context, request model.FindTrxInstitutionProductParams) (products []model.GetInstitutionProductResponse, err error)
	UpdateDtlInstitutionProductStock(ctx context.Context, request *model.DtlInstitutionProductStock) (err error)
}
