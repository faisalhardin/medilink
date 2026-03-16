package recall

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type RecallUC interface {
	CreateRecall(ctx context.Context, req model.CreateRecallRequest) (model.RecallResponse, error)
	UpdateRecall(ctx context.Context, req model.UpdateRecallRequest) error
	GetNextRecallByPatient(ctx context.Context, patientUUID string) (model.NextRecallResponse, error)
	ListRecalls(ctx context.Context, params model.GetRecallParams) ([]model.RecallResponse, error)
}
