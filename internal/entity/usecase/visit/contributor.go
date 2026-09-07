package visit

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// VisitContributorUC lists and manually attaches staff attributed to a visit for compensation.
type VisitContributorUC interface {
	ListVisitContributors(ctx context.Context, visitID int64) (model.ListVisitContributorsResponse, error)
	AddVisitContributor(ctx context.Context, visitID int64, staffID string) (model.AddVisitContributorResponse, error)
}
