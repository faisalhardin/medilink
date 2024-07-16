package usecase

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type InstitutionUC interface {
	InsertInstitution(ctx context.Context, request model.CreateInstitutionRequest) (err error)
	FindInstitutionByParams(ctx context.Context, params model.FindInstitutionParams) (result []model.Institution, err error)
}
