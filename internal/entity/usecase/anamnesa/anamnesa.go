package anamnesa

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type AnamnesaUC interface {
	GetByVisitID(ctx context.Context, visitID int64) (*model.TrxAnamnesa, error)
	Upsert(ctx context.Context, visitID int64, req model.UpsertAnamnesaRequest) (model.UpsertAnamnesaResponse, error)
}
