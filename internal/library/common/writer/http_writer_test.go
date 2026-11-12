package httpwriter

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	liblog "github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/common/log/logger"
	"github.com/faisalhardin/medilink/internal/library/util/requestinfo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
)

const piiText = "Budi Santoso 3201012345670001"

func TestSetErrorRoomWire(t *testing.T) {
	three := 3
	var logs []string
	restore := liblog.SetTestLogHook(func(level, msg string, fields logger.KV) {
		logs = append(logs, level+" "+msg+" "+fmt.Sprint(fields))
	})
	t.Cleanup(restore)

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
			logs = nil
			rec := httptest.NewRecorder()
			wrapped := errors.Wrap(row.err, piiText)
			if err := SetError(context.Background(), rec, wrapped); err != nil {
				t.Fatal(err)
			}
			if row.err.Code != row.wantCode {
				t.Fatalf("ErrorMessage.Code %d want %d", row.err.Code, row.wantCode)
			}
			if rec.Code != row.err.Code {
				t.Fatalf("status %d ErrorMessage.Code %d body %s", rec.Code, row.err.Code, rec.Body.Bytes())
			}
			assertRoomItem(t, rec.Body.Bytes(), row.wantName, row.err, row.count)
			assertNoPII(t, "body", rec.Body.String())
			commonerr.LogRoom(commonerr.RoomOpCreate, commonerr.RoomIDs{Room: 1}, wrapped)
			logged := strings.Join(logs, "\n")
			assertNoPII(t, "log", logged)
			if !strings.Contains(logged, row.wantName) {
				t.Fatalf("log missing %s: %s", row.wantName, logged)
			}
		})
	}

	t.Run("percent_w", func(t *testing.T) {
		logs = nil
		rec := httptest.NewRecorder()
		wrapped := fmt.Errorf("%s: %w", piiText, commonerr.RoomNameTaken())
		if err := SetError(context.Background(), rec, wrapped); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusConflict {
			t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
		}
		assertRoomItem(t, rec.Body.Bytes(), "name_taken", commonerr.RoomNameTaken(), nil)
		assertNoPII(t, "body", rec.Body.String())
		commonerr.LogRoom(commonerr.RoomOpRename, commonerr.RoomIDs{Room: 1}, wrapped)
		logged := strings.Join(logs, "\n")
		assertNoPII(t, "log", logged)
		if !strings.Contains(logged, "name_taken") {
			t.Fatalf("log: %s", logged)
		}
	})

	t.Run("internal", func(t *testing.T) {
		logs = nil
		rec := httptest.NewRecorder()
		secretErr := stderrors.New(piiText)
		if err := SetError(roomCtx(), rec, secretErr); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
		}
		assertRoomItem(t, rec.Body.Bytes(), "internal_server_error", commonerr.SetNewInternalError(), nil)
		assertNoPII(t, "body", rec.Body.String())
		commonerr.LogRoom(commonerr.RoomOpCreate, commonerr.RoomIDs{}, secretErr)
		logged := strings.Join(logs, "\n")
		assertNoPII(t, "log", logged)
		for _, want := range []string{"request failed", "*errors.errorString", "req-77", "POST", "/v1/rooms/{id}"} {
			if !strings.Contains(logged, want) {
				t.Fatalf("log missing %s: %s", want, logged)
			}
		}
	})

	t.Run("pg", func(t *testing.T) {
		logs = nil
		rec := httptest.NewRecorder()
		pgErr := &pgconn.PgError{
			Severity:       "ERROR",
			Code:           "23505",
			Message:        piiText,
			Detail:         piiText,
			Where:          piiText,
			TableName:      "patients",
			ConstraintName: "rooms_name_key",
		}
		wrapped := fmt.Errorf("%s: %w", piiText, pgErr)
		if err := SetError(roomCtx(), rec, wrapped); err != nil {
			t.Fatal(err)
		}
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status %d body %s", rec.Code, rec.Body.Bytes())
		}
		assertNoPII(t, "body", rec.Body.String())
		commonerr.LogRoom(commonerr.RoomOpCreate, commonerr.RoomIDs{}, wrapped)
		logged := strings.Join(logs, "\n")
		assertNoPII(t, "log", logged)
		for _, want := range []string{"23505", "rooms_name_key", "*fmt.wrapError", "req-77"} {
			if !strings.Contains(logged, want) {
				t.Fatalf("log missing %s: %s", want, logged)
			}
		}
	})
}

func roomCtx() context.Context {
	ctx := requestinfo.SetRequestContext(context.Background(), httptest.NewRequest(http.MethodPost, "/v1/rooms/9", nil))
	rctx := chi.NewRouteContext()
	rctx.RoutePatterns = []string{"/v1/rooms/{id}"}
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	return context.WithValue(ctx, middleware.RequestIDKey, "req-77")
}

func assertRoomItem(t *testing.T, body []byte, wantName string, msg *commonerr.ErrorMessage, count *int) {
	t.Helper()
	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		t.Fatal(err)
	}
	if len(top) != 1 {
		t.Fatalf("object keys %#v", top)
	}
	rawItems, ok := top["error_messages"]
	if !ok {
		t.Fatalf("body %s", body)
	}
	var items []map[string]interface{}
	if err := json.Unmarshal(rawItems, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("items %#v", items)
	}
	item := items[0]
	if item["error_name"] != wantName || item["error_name"] != msg.ErrorList[0].ErrorName {
		t.Fatalf("name %#v", item["error_name"])
	}
	got, has := item["count"]
	if count == nil {
		if has || len(item) != 2 {
			t.Fatalf("item %#v", item)
		}
		return
	}
	if !has || got != float64(*count) || len(item) != 3 {
		t.Fatalf("count %#v", item)
	}
}

func assertNoPII(t *testing.T, label, text string) {
	t.Helper()
	if strings.Contains(text, "Budi Santoso") || strings.Contains(text, "3201012345670001") {
		t.Fatalf("%s contains patient text: %s", label, text)
	}
}
