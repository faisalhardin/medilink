package recall

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type RecallDB interface {
	Insert(ctx context.Context, r *model.TrxRecall) error
	Update(ctx context.Context, id int64, institutionID int64, req model.UpdateRecallRequest) error
	GetByID(ctx context.Context, id int64, institutionID int64) (model.TrxRecallJoinPatient, error)
	GetNextByPatient(ctx context.Context, patientUUID string, institutionID int64) (model.TrxRecallJoinPatient, bool, error)
	ListUpcoming(ctx context.Context, params model.GetRecallParams, institutionID int64) ([]model.TrxRecallJoinPatient, error)
}
