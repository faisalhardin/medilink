package product

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type ProductUC interface {
	InsertMstProduct(ctx context.Context, req model.MstProduct) (err error)
	ListMstProductByParams(ctx context.Context, request model.MstProduct) (products []model.MstProduct, err error)
	UpdateMstProduct(ctx context.Context, request model.MstProduct) (err error)
	DeleteMstProduct(ctx context.Context, request model.MstProduct) (err error)
}
