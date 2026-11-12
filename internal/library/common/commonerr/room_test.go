package commonerr

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestSetNewBadRequestOmitsCount(t *testing.T) {
	raw, err := json.Marshal(SetNewBadRequest("a", "b"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"count"`) {
		t.Fatalf("raw JSON has count: %s", raw)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		t.Fatal(err)
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(top["error_list"], &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items: %#v", items)
	}
	item := items[0]
	if len(item) != 2 || item["error_name"] != "a" || item["error_description"] != "b" {
		t.Fatalf("item keys: %#v", item)
	}
}

func TestRoomConstructorsMatchSpec(t *testing.T) {
	three := 3
	spec := []struct {
		call        func() *ErrorMessage
		name        string
		status      int
		description string
		count       *int
	}{
		{RoomNameRequired, "name_required", 400, "Room name is required.", nil},
		{RoomNameTaken, "name_taken", 409, "An active room with this name already exists in this clinic.", nil},
		{RoomClearing, "clearing", 409, "The room is being cleared. Try again when the clear finishes.", nil},
		{func() *ErrorMessage { return RoomHasOpenVisits(3) }, "room_has_open_visits", 409, "The room still has open visits.", &three},
		{RoomCrossBoard, "cross_board", 400, "The room and the column belong to different boards.", nil},
		{RoomRequired, "room_required", 400, "This column has more than one room. Choose a room.", nil},
		{RoomNotInColumn, "room_not_in_column", 400, "The room is not linked to this column.", nil},
	}
	if len(spec) != len(roomErrors) {
		t.Fatalf("spec %d table %d", len(spec), len(roomErrors))
	}
	for i, row := range spec {
		msg := row.call()
		again := row.call()
		if msg == again {
			t.Fatalf("row %d returned the same pointer", i)
		}
		if msg.Code != row.status || roomErrors[i].name != row.name || roomErrors[i].status != row.status {
			t.Fatalf("row %d code %d table %#v", i, msg.Code, roomErrors[i])
		}
		if len(msg.ErrorList) != 1 || msg.ErrorList[0] == nil {
			t.Fatalf("row %d list %#v", i, msg.ErrorList)
		}
		item := msg.ErrorList[0]
		if item.ErrorName != row.name || item.ErrorDescription != row.description {
			t.Fatalf("row %d item %#v", i, item)
		}
		raw, err := json.Marshal(item)
		if err != nil {
			t.Fatal(err)
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatal(err)
		}
		got, has := decoded["count"]
		if row.count == nil {
			if has || item.count != nil || roomErrors[i].needsCount {
				t.Fatalf("row %d count present: %s", i, raw)
			}
			continue
		}
		if !roomErrors[i].needsCount || !has || got != float64(*row.count) || item.count == nil || *item.count != *row.count {
			t.Fatalf("row %d count %v field %v", i, got, item.count)
		}
	}
}

func TestRoomHasOpenVisitsStoresZero(t *testing.T) {
	msg := RoomHasOpenVisits(0)
	if msg.ErrorList[0].count == nil || *msg.ErrorList[0].count != 0 {
		t.Fatalf("count %#v", msg.ErrorList[0].count)
	}
	raw, err := json.Marshal(msg.ErrorList[0])
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["count"] != float64(0) {
		t.Fatalf("json %#v", decoded)
	}
}

func TestSet404IsFresh(t *testing.T) {
	first := Set404()
	second := Set404()
	if first == second {
		t.Fatal("Set404 returned the same pointer")
	}
	first.ErrorList[0].ErrorDescription = "changed"
	if second.ErrorList[0].ErrorDescription != "404 not found" {
		t.Fatalf("second description %q", second.ErrorList[0].ErrorDescription)
	}
	if second.Code != 404 || second.ErrorList[0].ErrorName != "404" {
		t.Fatalf("second %#v", second)
	}
}

func TestSetNewErrorUsesRoomConstructor(t *testing.T) {
	pii := "Budi Santoso 3201012345670001"
	msg := SetNewError(500, "name_taken", pii)
	if msg.Code != 409 || msg.ErrorList[0].ErrorName != "name_taken" || strings.Contains(msg.Error(), "Budi") {
		t.Fatalf("%s", msg.Error())
	}
	counted := SetNewError(400, "room_has_open_visits", pii)
	if counted.Code != 409 || counted.ErrorList[0].count == nil || *counted.ErrorList[0].count != 0 || strings.Contains(counted.Error(), "Budi") {
		t.Fatalf("%s", counted.Error())
	}
	bad := SetNewBadRequestByFormat(&ErrorFormat{ErrorName: "clearing", ErrorDescription: pii})
	if bad.Code != 409 || bad.ErrorList[0].ErrorName != "clearing" || strings.Contains(bad.Error(), "Budi") {
		t.Fatalf("%s", bad.Error())
	}
	appended := SetNewBadRequest("a", "b").Append("room_required", pii)
	if appended.Code != 400 || appended.ErrorList[1].ErrorDescription != "This column has more than one room. Choose a room." || strings.Contains(appended.Error(), "Budi") {
		t.Fatalf("%s", appended.Error())
	}
	kept := SetNewBadRequest("a", "b").Append("field", "must be set")
	if kept.ErrorList[1].ErrorName != "field" || kept.ErrorList[1].ErrorDescription != "must be set" {
		t.Fatalf("%#v", kept.ErrorList)
	}
}

func TestSetNewErrorAllowsOtherNames(t *testing.T) {
	msg := SetNewError(400, "a", "b")
	if msg.Code != 400 || msg.ErrorList[0].ErrorName != "a" || msg.ErrorList[0].ErrorDescription != "b" {
		t.Fatalf("%#v", msg)
	}
	if strings.Contains(string(msg.Marshal()), `"count"`) {
		t.Fatalf("count leaked: %s", msg.Marshal())
	}
}

func TestRoomLogFieldsOpenVisits(t *testing.T) {
	fields := roomLogFields(RoomIDs{Institution: 9, Room: 4}, pkgerrors.Wrap(RoomHasOpenVisits(3), "room repo"))
	if fields["institution_id"] != int64(9) || fields["room_id"] != int64(4) {
		t.Fatalf("ids %#v", fields)
	}
	if fields["status"] != 409 || fields["error_name"] != "room_has_open_visits" || fields["count"] != 3 {
		t.Fatalf("error %#v", fields)
	}
	if _, ok := fields["error_description"]; ok {
		t.Fatalf("description copied: %#v", fields)
	}
	for _, key := range []string{"board_id", "column_id", "job_id"} {
		if _, ok := fields[key]; ok {
			t.Fatalf("zero id %s in %#v", key, fields)
		}
	}
}

func TestRoomLogFieldsPercentW(t *testing.T) {
	pii := "Budi Santoso 3201012345670001"
	taken := roomLogFields(RoomIDs{}, fmt.Errorf("%s: %w", pii, RoomNameTaken()))
	if taken["status"] != 409 || taken["error_name"] != "name_taken" {
		t.Fatalf("%#v", taken)
	}
	open := roomLogFields(RoomIDs{}, fmt.Errorf("%s: %w", pii, RoomHasOpenVisits(3)))
	if open["status"] != 409 || open["error_name"] != "room_has_open_visits" || open["count"] != 3 {
		t.Fatalf("%#v", open)
	}
	for _, fields := range []map[string]interface{}{taken, open} {
		for _, value := range fields {
			text := strings.ToLower(stringify(value))
			if strings.Contains(text, "budi") || strings.Contains(text, "3201012345670001") {
				t.Fatalf("sentinel in %#v", fields)
			}
		}
	}
}

func TestRoomLogFieldsDropsErrorText(t *testing.T) {
	fields := roomLogFields(RoomIDs{}, pkgerrors.Wrap(errors.New("visit 987654321 Budi"), "x"))
	if fields["status"] != 500 {
		t.Fatalf("status %#v", fields)
	}
	if _, ok := fields["error_type"]; !ok {
		t.Fatalf("missing error_type %#v", fields)
	}
	for _, value := range fields {
		text := strings.ToLower(stringify(value))
		if strings.Contains(text, "987654321") || strings.Contains(text, "budi") {
			t.Fatalf("sentinel in %#v", fields)
		}
	}
}

func TestLogRoomNil(t *testing.T) {
	LogRoom(RoomOpCreate, RoomIDs{Institution: 1}, nil)
}

func stringify(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		raw, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(raw)
	}
}
