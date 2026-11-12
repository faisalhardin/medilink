package commonerr

import (
	"errors"
	"fmt"

	liblog "github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
)

type roomKind int

const (
	kindNameRequired roomKind = iota
	kindNameTaken
	kindClearing
	kindHasOpenVisits
	kindCrossBoard
	kindRoomRequired
	kindRoomNotInColumn
)

type roomError struct {
	kind        roomKind
	name        string
	status      int
	description string
	needsCount  bool
}

var roomErrors = []roomError{
	{kind: kindNameRequired, name: "name_required", status: 400, description: "Room name is required."},
	{kind: kindNameTaken, name: "name_taken", status: 409, description: "An active room with this name already exists in this clinic."},
	{kind: kindClearing, name: "clearing", status: 409, description: "The room is being cleared. Try again when the clear finishes."},
	{kind: kindHasOpenVisits, name: "room_has_open_visits", status: 409, description: "The room still has open visits.", needsCount: true},
	{kind: kindCrossBoard, name: "cross_board", status: 400, description: "The room and the column belong to different boards."},
	{kind: kindRoomRequired, name: "room_required", status: 400, description: "This column has more than one room. Choose a room."},
	{kind: kindRoomNotInColumn, name: "room_not_in_column", status: 400, description: "The room is not linked to this column."},
}

func init() {
	seen := make(map[string]struct{}, len(roomErrors))
	needsCount := 0
	for _, row := range roomErrors {
		if row.name == "" || row.status == 0 {
			panic("invalid room error row")
		}
		if _, dup := seen[row.name]; dup {
			panic("duplicate room error name")
		}
		seen[row.name] = struct{}{}
		if row.needsCount {
			needsCount++
		}
	}
	if needsCount != 1 {
		panic("room error table needs exactly one counted row")
	}
}

// shortcut: room_has_open_visits reports 0 here, because this path has no visit count; callers use RoomHasOpenVisits.
func roomByName(name string) *ErrorMessage {
	for i := range roomErrors {
		if roomErrors[i].name != name {
			continue
		}
		if roomErrors[i].needsCount {
			return RoomHasOpenVisits(0)
		}
		return roomMessage(roomErrors[i].kind, nil)
	}
	return nil
}

func roomMessage(kind roomKind, count *int) *ErrorMessage {
	for i := range roomErrors {
		if roomErrors[i].kind != kind {
			continue
		}
		row := roomErrors[i]
		format := &ErrorFormat{
			ErrorName:        row.name,
			ErrorDescription: row.description,
		}
		if row.needsCount {
			format.count = count
		}
		return allocError(row.status, format)
	}
	panic("invalid room error row")
}

func RoomNameRequired() *ErrorMessage {
	return roomMessage(kindNameRequired, nil)
}

func RoomNameTaken() *ErrorMessage {
	return roomMessage(kindNameTaken, nil)
}

func RoomClearing() *ErrorMessage {
	return roomMessage(kindClearing, nil)
}

func RoomHasOpenVisits(openVisits int) *ErrorMessage {
	return roomMessage(kindHasOpenVisits, &openVisits)
}

func RoomCrossBoard() *ErrorMessage {
	return roomMessage(kindCrossBoard, nil)
}

func RoomRequired() *ErrorMessage {
	return roomMessage(kindRoomRequired, nil)
}

// The caller must 404 a missing, other-clinic, orphan, or archived room before this 400, or another clinic's id is revealed.
func RoomNotInColumn() *ErrorMessage {
	return roomMessage(kindRoomNotInColumn, nil)
}

type RoomOp string

const (
	RoomOpCreate     RoomOp = "room.create"
	RoomOpRename     RoomOp = "room.rename"
	RoomOpArchive    RoomOp = "room.archive"
	RoomOpClinical   RoomOp = "room.clinical"
	RoomOpLinkColumn RoomOp = "room.link_column"
	RoomOpMoveVisit  RoomOp = "room.move_visit"
	RoomOpClear      RoomOp = "room.clear"
)

type RoomIDs struct {
	Institution int64
	Board       int64
	Column      int64
	Room        int64
	Job         int64
}

func LogRoom(op RoomOp, ids RoomIDs, err error) {
	if err == nil {
		return
	}
	fields := roomLogFields(ids, err)
	status, _ := fields["status"].(int)
	if status >= 400 && status <= 499 {
		liblog.InfoWithFields(string(op)+" refused", fields)
		return
	}
	liblog.ErrorWithFields(string(op)+" failed", fields)
}

func roomLogFields(ids RoomIDs, err error) logger.KV {
	fields := logger.KV{}
	if ids.Institution != 0 {
		fields["institution_id"] = ids.Institution
	}
	if ids.Board != 0 {
		fields["board_id"] = ids.Board
	}
	if ids.Column != 0 {
		fields["column_id"] = ids.Column
	}
	if ids.Room != 0 {
		fields["room_id"] = ids.Room
	}
	if ids.Job != 0 {
		fields["job_id"] = ids.Job
	}
	var msg *ErrorMessage
	if !errors.As(err, &msg) || msg == nil {
		fields["status"] = 500
		fields["error_type"] = fmt.Sprintf("%T", err)
		return fields
	}
	fields["status"] = msg.Code
	if len(msg.ErrorList) == 0 || msg.ErrorList[0] == nil {
		return fields
	}
	fields["error_name"] = msg.ErrorList[0].ErrorName
	if msg.ErrorList[0].count != nil {
		fields["count"] = *msg.ErrorList[0].count
	}
	return fields
}
