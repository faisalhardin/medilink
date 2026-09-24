package compensation

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
)

func TestCommissionRowToResponse_UsesWorksheetUUID(t *testing.T) {
	got := commissionRowToResponse(compensationrepo.VisitCommissionListRow{
		ID:            9,
		WorksheetID:   44,
		WorksheetUUID: "ws-public",
		VisitID:       100,
		RevenueBase:   50000,
	})
	if got.WorksheetUUID != "ws-public" {
		t.Fatalf("WorksheetUUID = %q, want ws-public", got.WorksheetUUID)
	}
	if got.ID != 9 || got.VisitID != 100 {
		t.Fatalf("id/visit = %d/%d", got.ID, got.VisitID)
	}

	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if _, ok := decoded["worksheet_id"]; ok {
		t.Fatal("response must not expose internal worksheet_id")
	}
	if decoded["worksheet_uuid"] != "ws-public" {
		t.Fatalf("worksheet_uuid = %v", decoded["worksheet_uuid"])
	}
	if decoded["commission_type"] != nil || decoded["commission_amount"] != nil {
		t.Fatalf("unapproved mask failed: type=%v amount=%v", decoded["commission_type"], decoded["commission_amount"])
	}
}

func TestCommissionRowToResponse_ApprovedFields(t *testing.T) {
	row := commissionRowToResponse(compensationrepo.VisitCommissionListRow{
		ID:               1,
		WorksheetUUID:    "ws-1",
		VisitID:          2,
		CommissionType:   model.CommissionTypePercent,
		CommissionAmount: 1500,
		ApprovedAt:       sql.NullTime{Time: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Valid: true},
	})
	if !row.CommissionType.Valid || row.CommissionType.String != "percent" {
		t.Fatalf("type = %+v", row.CommissionType)
	}
	if !row.CommissionAmount.Valid || row.CommissionAmount.Int64 != 1500 {
		t.Fatalf("amount = %+v", row.CommissionAmount)
	}
}
