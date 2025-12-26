package odontogram

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// OdontogramRepo defines the interface for odontogram repository operations
type OdontogramRepo interface {
	// Event operations
	InsertEvent(ctx context.Context, event *model.HstOdontogram) error
	InsertEventsBatch(ctx context.Context, events []*model.HstOdontogram) error
	GetEventsByPatient(ctx context.Context, params model.GetEventsByPatientParams) ([]model.HstOdontogram, error)
	GetEventsByPatientFiltered(ctx context.Context, params model.GetOdontogramEventsParams, patientID int64) ([]model.HstOdontogram, error)
	GetEventByID(ctx context.Context, institutionID int64, eventID string) (*model.HstOdontogram, error)
	GetMaxSequenceNumber(ctx context.Context, institutionID, patientID int64) (int64, error)
	GetMaxLogicalTimestamp(ctx context.Context, institutionID, patientID int64) (int64, error)

	// Snapshot operations
	GetSnapshot(ctx context.Context, institutionID, patientID int64) (*model.MstPatientOdontogram, error)
	UpsertSnapshot(ctx context.Context, snapshot *model.MstPatientOdontogram) error

	// Visit operations
	HasEventsForVisit(ctx context.Context, institutionID, visitID int64) (bool, error)
}
