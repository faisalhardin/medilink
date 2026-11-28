package visit

import (
	"context"
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

type stepRoomMode int

const (
	stepRoomMove stepRoomMode = iota
	stepRoomNote
)

// resolveStepRoom returns the room id stored on the new step. Zero means SQL NULL.
// Move with no id: zero rooms stays empty, one room is that id, more than one is RoomRequired.
// Move with an id locks it, then returns RoomNotInColumn or Set404 for a room the clinic cannot use.
// Note turns a bad id into zero and never returns RoomRequired when the id is omitted.
func (u *VisitUC) resolveStepRoom(ctx context.Context, clinicID, columnID int64, requested null.Int64, mode stepRoomMode) (int64, error) {
	if requested.Valid && requested.Int64 > 0 {
		if err := u.shareStepRoom(ctx, clinicID, requested.Int64); err != nil {
			if mode == stepRoomNote && isNotFound(err) {
				return 0, nil
			}
			return 0, err
		}
		rooms, err := u.JourneyDB.ListColumnRooms(ctx, clinicID, columnID)
		if err != nil {
			return 0, err
		}
		if !columnHasRoom(rooms, requested.Int64) {
			if mode == stepRoomNote {
				return 0, nil
			}
			return 0, commonerr.RoomNotInColumn()
		}
		return requested.Int64, nil
	}
	if mode == stepRoomNote {
		return 0, nil
	}

	rooms, err := u.JourneyDB.ListColumnRooms(ctx, clinicID, columnID)
	if err != nil {
		return 0, err
	}
	switch len(rooms) {
	case 0:
		return 0, nil
	case 1:
		if err := u.shareStepRoom(ctx, clinicID, rooms[0].ID); err != nil {
			return 0, err
		}
		return rooms[0].ID, nil
	default:
		return 0, commonerr.RoomRequired()
	}
}

func (u *VisitUC) shareStepRoom(ctx context.Context, clinicID, roomID int64) error {
	room, err := u.JourneyDB.LockRoomShare(ctx, clinicID, roomID)
	if err != nil {
		if errors.Is(err, constant.ErrorNoAffectedRow) {
			return commonerr.Set404()
		}
		return err
	}
	if room == nil || room.ID == 0 {
		return commonerr.Set404()
	}
	return nil
}

func columnHasRoom(rooms []model.PickerRoom, roomID int64) bool {
	for _, room := range rooms {
		if room.ID == roomID {
			return true
		}
	}
	return false
}

func isNotFound(err error) bool {
	var msg *commonerr.ErrorMessage
	return errors.As(err, &msg) && msg != nil && msg.Code == http.StatusNotFound
}
