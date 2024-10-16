package product

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type ProductDB interface {
	InsertMstProduct(ctx context.Context, institution *model.MstProduct) (err error)
	FindMstProductByParams(ctx context.Context, request model.MstProduct) (products []model.MstProduct, err error)
	UpdateMstProduct(ctx context.Context, mstProduct *model.MstProduct) (err error)
	DeleteMstProduct(ctx context.Context, mstProduct *model.MstProduct) (err error)
}
