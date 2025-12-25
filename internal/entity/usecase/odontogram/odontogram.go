package odontogram

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// OdontogramUC defines the interface for odontogram use case operations
type OdontogramUC interface {
	// CreateEvents creates one or more odontogram events
	CreateEvents(ctx context.Context, requests []model.CreateOdontogramEventRequest) (*model.CreateOdontogramEventsResponse, error)

	// GetEvents retrieves events for a patient with filtering
	GetEvents(ctx context.Context, params model.GetOdontogramEventsParams) (*model.GetOdontogramEventsResponse, error)

	// GetSnapshot retrieves the current or historical snapshot for a patient
	GetSnapshot(ctx context.Context, params model.GetOdontogramSnapshotParams) (*model.GetOdontogramSnapshotResponse, error)

	// BuildSnapshot builds a snapshot from events up to a specific sequence number
	BuildSnapshot(ctx context.Context, params model.GetEventsByPatientParams) (*model.OdontogramSnapshot, int64, int64, error)
}
