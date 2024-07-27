package patient

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type PatientUC interface {
	RegisterNewPatient(ctx context.Context, req model.RegisterNewPatientRequest) (err error)
}
