package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type InstitutionUC interface {
	InsertInstitution(ctx context.Context, request model.CreateInstitutionRequest) (err error)
	FindInstitutionByParams(ctx context.Context, params model.FindInstitutionParams) (result []model.Institution, err error)
	GetInstitutionByUserContext(ctx context.Context) (result model.Institution, err error)

	InserInstitutionProduct(ctx context.Context, newProduct model.InsertInstitutionProductRequest) (product model.TrxInstitutionProduct, err error)
	FindInstitutionProductByParams(ctx context.Context, params model.FindTrxInstitutionProductParams) (products []model.GetInstitutionProductResponse, err error)
	UpdateInstitutionProduct(ctx context.Context, request model.UpdateInstitutionProductRequest) (err error)
	UpdateInstitutionProductStock(ctx context.Context, product model.DtlInstitutionProductStock) (err error)
}
