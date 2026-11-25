package model

import (
	"encoding/json"
	"testing"
)

func TestParseSyncStatusAcceptsOnlyFiveStatuses(t *testing.T) {
	for _, s := range []string{"pending", "processing", "synced", "failed", "inactive"} {
		got, ok := ParseSyncStatus(s)
		if !ok || string(got) != s {
			t.Fatalf("ParseSyncStatus(%q) = %q, %v", s, got, ok)
		}
	}
	for _, s := range []string{"", "PENDING", "done", "synced ", "queued"} {
		if got, ok := ParseSyncStatus(s); ok || got != "" {
			t.Fatalf("ParseSyncStatus(%q) = %q, %v, want reject", s, got, ok)
		}
	}
}

func TestRoomJSONShape(t *testing.T) {
	failure := "timeout"
	for _, tc := range []struct {
		name string
		room Room
		want string
	}{
		{
			name: "no location no columns",
			room: Room{ID: 4, Name: "Poli", BoardID: 2, JourneyPointIDs: []int64{}},
			want: `{"id":4,"name":"Poli","board_id":2,"is_clinical":false,"archived":false,"journey_point_ids":[],"location":null}`,
		},
		{
			name: "archived room",
			room: Room{ID: 4, Name: "Poli", BoardID: 2, Archived: true, JourneyPointIDs: []int64{}},
			want: `{"id":4,"name":"Poli","board_id":2,"is_clinical":false,"archived":true,"journey_point_ids":[],"location":null}`,
		},
		{
			name: "pending location",
			room: Room{ID: 4, Name: "Poli", BoardID: 2, JourneyPointIDs: []int64{7, 9},
				Location: &RoomLocation{SyncStatus: SyncPending}},
			want: `{"id":4,"name":"Poli","board_id":2,"is_clinical":false,"archived":false,"journey_point_ids":[7,9],"location":{"sync_status":"pending","last_job_type":"","failure_kind":null,"from_clear":false}}`,
		},
		{
			name: "failed location",
			room: Room{ID: 4, Name: "Poli", BoardID: 2, JourneyPointIDs: []int64{},
				Location: &RoomLocation{SyncStatus: SyncFailed, LastJobType: "create", FailureKind: &failure, FromClear: true}},
			want: `{"id":4,"name":"Poli","board_id":2,"is_clinical":false,"archived":false,"journey_point_ids":[],"location":{"sync_status":"failed","last_job_type":"create","failure_kind":"timeout","from_clear":true}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.room)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got  %s\nwant %s", got, tc.want)
			}
		})
	}
}
