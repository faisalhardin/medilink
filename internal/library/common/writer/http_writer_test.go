package httpwriter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/pkg/errors"
)

func TestSetErrorRoomWire(t *testing.T) {
	three := 3
	pii := []string{"Budi Santoso", "3201012345670001", "patient note", "sekrit-token", "987654321"}
	rows := []struct {
		name     string
		err      *commonerr.ErrorMessage
		wantCode int
		wantName string
		count    *int
	}{
		{name: "name_required", err: commonerr.RoomNameRequired(), wantCode: 400, wantName: "name_required"},
		{name: "name_taken", err: commonerr.RoomNameTaken(), wantCode: 409, wantName: "name_taken"},
		{name: "clearing", err: commonerr.RoomClearing(), wantCode: 409, wantName: "clearing"},
		{name: "room_has_open_visits", err: commonerr.RoomHasOpenVisits(3), wantCode: 409, wantName: "room_has_open_visits", count: &three},
		{name: "cross_board", err: commonerr.RoomCrossBoard(), wantCode: 400, wantName: "cross_board"},
		{name: "room_required", err: commonerr.RoomRequired(), wantCode: 400, wantName: "room_required"},
		{name: "room_not_in_column", err: commonerr.RoomNotInColumn(), wantCode: 400, wantName: "room_not_in_column"},
		{name: "not_found", err: commonerr.Set404(), wantCode: 404, wantName: "404"},
		{name: "missing", err: commonerr.Set404(), wantCode: 404, wantName: "404"},
		{name: "other_clinic", err: commonerr.Set404(), wantCode: 404, wantName: "404"},
		{name: "orphan", err: commonerr.Set404(), wantCode: 404, wantName: "404"},
		{name: "archived", err: commonerr.Set404(), wantCode: 404, wantName: "404"},
		{name: "room_not_in_column_after_lookup", err: commonerr.RoomNotInColumn(), wantCode: 400, wantName: "room_not_in_column"},
		{name: "room_required_choose", err: commonerr.RoomRequired(), wantCode: 400, wantName: "room_required"},
	}

	for _, row := range rows {
		row := row
		t.Run(row.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if err := SetError(context.Background(), rec, errors.Wrap(row.err, row.name)); err != nil {
				t.Fatal(err)
			}
			if row.err.Code != row.wantCode {
				t.Fatalf("ErrorMessage.Code %d want %d", row.err.Code, row.wantCode)
			}
			if rec.Code != row.err.Code {
				t.Fatalf("status %d ErrorMessage.Code %d body %s", rec.Code, row.err.Code, rec.Body.Bytes())
			}
			var top map[string]json.RawMessage
			if err := json.Unmarshal(rec.Body.Bytes(), &top); err != nil {
				t.Fatal(err)
			}
			if len(top) != 1 {
				t.Fatalf("object keys %#v", top)
			}
			rawItems, ok := top["error_messages"]
			if !ok {
				t.Fatalf("body %s", rec.Body.Bytes())
			}
			var items []map[string]interface{}
			if err := json.Unmarshal(rawItems, &items); err != nil {
				t.Fatal(err)
			}
			if len(items) != 1 {
				t.Fatalf("items %#v", items)
			}
			item := items[0]
			if item["error_name"] != row.wantName || item["error_name"] != row.err.ErrorList[0].ErrorName {
				t.Fatalf("name %#v", item["error_name"])
			}
			got, has := item["count"]
			if row.count == nil {
				if has || len(item) != 2 {
					t.Fatalf("item %#v", item)
				}
			} else if !has || got != float64(*row.count) || len(item) != 3 {
				t.Fatalf("count %#v", item)
			}
			body := rec.Body.String()
			for _, secret := range pii {
				if strings.Contains(body, secret) {
					t.Fatalf("body contains %q: %s", secret, body)
				}
			}
		})
	}

	rec := httptest.NewRecorder()
	if err := SetError(context.Background(), rec, fmt.Errorf("x: %w", commonerr.RoomNameTaken())); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &top); err != nil {
		t.Fatal(err)
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(top["error_messages"], &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0]["error_name"] != "internal_server_error" {
		t.Fatalf("body %s", rec.Body.Bytes())
	}
	body := rec.Body.String()
	for _, secret := range pii {
		if strings.Contains(body, secret) {
			t.Fatalf("body contains %q: %s", secret, body)
		}
	}
}
