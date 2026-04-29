package diagnosis

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// DiagnosisUC is the diagnosis endpoint orchestration contract.
type DiagnosisUC interface {
	GetByVisitID(ctx context.Context, visitID int64) ([]model.DiagnosisResponse, error)
	Save(ctx context.Context, visitID int64, req model.SaveDiagnosesRequest) (model.SaveDiagnosesResponse, error)
	Delete(ctx context.Context, visitID, diagnosisID int64) error
}
