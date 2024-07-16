package institution

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type InstitutionDB interface {
	InsertNewInstitution(ctx context.Context, institution *model.Institution) (err error)
	FindInstitutionByParams(ctx context.Context, request model.FindInstitutionParams) (institutions []model.Institution, err error)
}
