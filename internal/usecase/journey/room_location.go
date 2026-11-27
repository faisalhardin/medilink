package journey

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// RoomLocationPort is the SH-1 seam. BE-3 ships a no-op; SH-1/SH-2 fill Read and Enqueue.
// Clearing is the SH-2 lock check. roomID 0 is a create.
type RoomLocationPort interface {
	Read(ctx context.Context, clinicID int64, roomIDs []int64) (map[int64]*model.RoomLocation, error)
	Enqueue(ctx context.Context, clinicID, roomID int64, jobType string) error
	Clearing(ctx context.Context, clinicID, roomID int64) (bool, error)
}

const (
	jobLocationCreate     = "location_create"
	jobLocationInactivate = "location_inactivate"
)

type NoopRoomLocation struct{}

func (NoopRoomLocation) Read(context.Context, int64, []int64) (map[int64]*model.RoomLocation, error) {
	return nil, nil
}

func (NoopRoomLocation) Enqueue(context.Context, int64, int64, string) error { return nil }

func (NoopRoomLocation) Clearing(context.Context, int64, int64) (bool, error) { return false, nil }

func roomLocationPort(port RoomLocationPort) RoomLocationPort {
	if port == nil {
		return NoopRoomLocation{}
	}
	return port
}
