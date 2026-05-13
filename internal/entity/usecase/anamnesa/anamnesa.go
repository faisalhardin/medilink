package anamnesa

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type AnamnesaUC interface {
	GetByVisitID(ctx context.Context, visitID int64) (*model.AnamnesaResponse, error)
	GetDetailedByVisitID(ctx context.Context, visitID int64) (*model.AnamnesaDetailedResponse, error)
	Upsert(ctx context.Context, visitID int64, req model.UpsertAnamnesaRequest) (model.UpsertAnamnesaResponse, error)
}
