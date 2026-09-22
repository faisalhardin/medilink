package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// VisitCommissionUC is the visit-commission lifecycle contract.
type VisitCommissionUC interface {
	// GenerateVisitCommissions triggers async generation of commission rows for a worksheet.
	// Returns 202-style response with generate_status=running.
	GenerateVisitCommissions(ctx context.Context, req model.GenerateVisitCommissionsRequest) (model.GenerateVisitCommissionsResponse, error)

	// ListVisitCommissions returns paginated commission rows for a worksheet.
	ListVisitCommissions(ctx context.Context, req model.ListVisitCommissionsRequest) (model.ListVisitCommissionsResponse, error)

	// PatchCommissionItem updates commission_type/amounts on a live commission row.
	PatchCommissionItem(ctx context.Context, req model.PatchCommissionItemRequest) (model.PatchCommissionItemResponse, error)

	// ArchiveVisitCommission soft-deletes a commission row.
	ArchiveVisitCommission(ctx context.Context, id int64) (model.ArchiveVisitCommissionResponse, error)
}
