package odontogram

import (
	"encoding/json"
	"testing"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
)

func TestNewSnapshotBuilder_FilterDuplicates(t *testing.T) {
	tests := []struct {
		name     string
		events   []model.HstOdontogram
		expected int
	}{
		{
			name: "filter duplicate event_ids",
			events: []model.HstOdontogram{
				{EventID: "event-1", SequenceNumber: 1},
				{EventID: "event-2", SequenceNumber: 2},
				{EventID: "event-1", SequenceNumber: 3}, // duplicate
			},
			expected: 2,
		},
		{
			name: "keep events with empty event_id",
			events: []model.HstOdontogram{
				{EventID: "", SequenceNumber: 1},
				{EventID: "", SequenceNumber: 2},
				{EventID: "event-1", SequenceNumber: 3},
			},
			expected: 3, // All kept, empty IDs not filtered
		},
		{
			name: "no duplicates",
			events: []model.HstOdontogram{
				{EventID: "event-1", SequenceNumber: 1},
				{EventID: "event-2", SequenceNumber: 2},
				{EventID: "event-3", SequenceNumber: 3},
			},
			expected: 3,
		},
		{
			name:     "empty events",
			events:   []model.HstOdontogram{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSnapshotBuilder(tt.events)
			if len(builder.events) != tt.expected {
				t.Errorf("expected %d events, got %d", tt.expected, len(builder.events))
			}
		})
	}
}

func TestSnapshotBuilder_Build_ToothCodeInsert(t *testing.T) {
	eventData := model.OdontogramEventData{
		WholeToothCode: []string{"amf", "rct"},
	}
	eventDataJSON, _ := json.Marshal(eventData)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        eventDataJSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if snapshot == nil {
		t.Fatal("snapshot is nil")
	}
	if _, exists := snapshot.Teeth["11"]; !exists {
		t.Fatal("tooth 11 not found in snapshot")
	}
	tooth := snapshot.Teeth["11"]
	if !contains(tooth.WholeToothCode, "amf") {
		t.Error("expected amf in WholeToothCode")
	}
	if !contains(tooth.WholeToothCode, "rct") {
		t.Error("expected rct in WholeToothCode")
	}
}

func TestSnapshotBuilder_Build_ToothCodeRemove(t *testing.T) {
	insertData := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	insertDataJSON, _ := json.Marshal(insertData)

	removeData := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	removeDataJSON, _ := json.Marshal(removeData)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        insertDataJSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeRemove,
			EventData:        removeDataJSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if contains(tooth.WholeToothCode, "amf") {
		t.Error("expected amf to be removed from WholeToothCode")
	}
}

func TestSnapshotBuilder_Build_GeneralNotes_LastWriteWins(t *testing.T) {
	note1Data := model.OdontogramEventData{
		GeneralNotes: "First note",
	}
	note1JSON, _ := json.Marshal(note1Data)

	note2Data := model.OdontogramEventData{
		GeneralNotes: "Second note",
	}
	note2JSON, _ := json.Marshal(note2Data)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        note1JSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        note2JSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200, // Higher timestamp wins
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if tooth.GeneralNotes != "Second note" {
		t.Errorf("expected 'Second note', got '%s'", tooth.GeneralNotes)
	}
}

func TestSnapshotBuilder_Build_SurfaceCodeSet(t *testing.T) {
	surfaceData := model.OdontogramEventData{
		Surface:     "O",
		SurfaceCode: "car",
	}
	surfaceDataJSON, _ := json.Marshal(surfaceData)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothSurfaceCodeSet,
			EventData:        surfaceDataJSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if len(tooth.Surfaces) != 1 {
		t.Fatalf("expected 1 surface, got %d", len(tooth.Surfaces))
	}
	surface := tooth.Surfaces[0]
	if surface.Surface != "O" {
		t.Errorf("expected Surface 'O', got '%s'", surface.Surface)
	}
	if surface.Code != "car" {
		t.Errorf("expected Code 'car', got '%s'", surface.Code)
	}
	if surface.Condition != "Caries" {
		t.Errorf("expected Condition 'Caries', got '%s'", surface.Condition)
	}
	if surface.Color != "#ff0000" {
		t.Errorf("expected Color '#ff0000', got '%s'", surface.Color)
	}
}

func TestSnapshotBuilder_Build_SurfaceNoteUpdate(t *testing.T) {
	codeData := model.OdontogramEventData{
		Surface:     "O",
		SurfaceCode: "car",
	}
	codeDataJSON, _ := json.Marshal(codeData)

	noteData := model.OdontogramEventData{
		Surface:      "O",
		SurfaceNotes: "Small cavity",
	}
	noteDataJSON, _ := json.Marshal(noteData)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothSurfaceCodeSet,
			EventData:        codeDataJSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothSurfaceNoteUpdate,
			EventData:        noteDataJSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if len(tooth.Surfaces) != 1 {
		t.Fatalf("expected 1 surface, got %d", len(tooth.Surfaces))
	}
	surface := tooth.Surfaces[0]
	if surface.Notes != "Small cavity" {
		t.Errorf("expected Notes 'Small cavity', got '%s'", surface.Notes)
	}
	if surface.Code != "car" {
		t.Errorf("expected Code 'car', got '%s'", surface.Code)
	}
}

func TestSnapshotBuilder_Build_ToothReset(t *testing.T) {
	insertData := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	insertDataJSON, _ := json.Marshal(insertData)

	noteData := model.OdontogramEventData{
		GeneralNotes: "Some notes",
	}
	noteDataJSON, _ := json.Marshal(noteData)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        insertDataJSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        noteDataJSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-3",
			ToothID:          "11",
			EventType:        constant.EventTypeToothReset,
			EventData:        json.RawMessage("{}"),
			SequenceNumber:   3,
			LogicalTimestamp: 300,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if len(tooth.WholeToothCode) != 0 {
		t.Errorf("expected empty WholeToothCode, got %v", tooth.WholeToothCode)
	}
	if tooth.GeneralNotes != "" {
		t.Errorf("expected empty GeneralNotes, got '%s'", tooth.GeneralNotes)
	}
	if len(tooth.Surfaces) != 0 {
		t.Errorf("expected empty Surfaces, got %d", len(tooth.Surfaces))
	}
}

func TestSnapshotBuilder_Build_EventOrdering(t *testing.T) {
	// Events with different logical timestamps - should be sorted
	note1Data := model.OdontogramEventData{
		GeneralNotes: "First",
	}
	note1JSON, _ := json.Marshal(note1Data)

	note2Data := model.OdontogramEventData{
		GeneralNotes: "Second",
	}
	note2JSON, _ := json.Marshal(note2Data)

	events := []model.HstOdontogram{
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        note2JSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200, // Higher timestamp
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        note1JSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100, // Lower timestamp
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Should apply in timestamp order, so "Second" wins
	tooth := snapshot.Teeth["11"]
	if tooth.GeneralNotes != "Second" {
		t.Errorf("expected 'Second', got '%s'", tooth.GeneralNotes)
	}
}

func TestSnapshotBuilder_Build_SameTimestamp_DifferentStaffID(t *testing.T) {
	note1Data := model.OdontogramEventData{
		GeneralNotes: "Staff 1 note",
	}
	note1JSON, _ := json.Marshal(note1Data)

	note2Data := model.OdontogramEventData{
		GeneralNotes: "Staff 2 note",
	}
	note2JSON, _ := json.Marshal(note2Data)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        note1JSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothGeneralNoteUpdate,
			EventData:        note2JSON,
			SequenceNumber:   2,
			LogicalTimestamp: 100, // Same timestamp
			CreatedByStaffID: 2,   // Higher staff ID wins
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if tooth.GeneralNotes != "Staff 2 note" {
		t.Errorf("expected 'Staff 2 note', got '%s'", tooth.GeneralNotes)
	}
}

func TestSnapshotBuilder_Build_MultipleTeeth(t *testing.T) {
	tooth11Data := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	tooth11JSON, _ := json.Marshal(tooth11Data)

	tooth12Data := model.OdontogramEventData{
		WholeToothCode: []string{"rct"},
	}
	tooth12JSON, _ := json.Marshal(tooth12Data)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        tooth11JSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "12",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        tooth12JSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if _, exists := snapshot.Teeth["11"]; !exists {
		t.Error("tooth 11 not found")
	}
	if _, exists := snapshot.Teeth["12"]; !exists {
		t.Error("tooth 12 not found")
	}
	if !contains(snapshot.Teeth["11"].WholeToothCode, "amf") {
		t.Error("expected amf in tooth 11")
	}
	if !contains(snapshot.Teeth["12"].WholeToothCode, "rct") {
		t.Error("expected rct in tooth 12")
	}
}

func TestSnapshotBuilder_FromSnapshot(t *testing.T) {
	snapshot := &model.OdontogramSnapshot{
		Teeth: map[string]model.ToothData{
			"11": {
				ID:             "11",
				WholeToothCode: []string{"amf", "rct"},
				GeneralNotes:   "Test notes",
				Surfaces: []model.SurfaceData{
					{
						Surface:          "O",
						Code:             "car",
						Notes:            "Cavity",
						LogicalTimestamp: 100,
						CreatedByStaffID: 1,
					},
				},
			},
		},
	}

	builder := NewSnapshotBuilder([]model.HstOdontogram{})
	builder.FromSnapshot(snapshot)

	if _, exists := builder.teeth["11"]; !exists {
		t.Fatal("tooth 11 not found")
	}
	tooth := builder.teeth["11"]
	if !tooth.WholeToothCode.codes["amf"] {
		t.Error("expected amf in codes")
	}
	if !tooth.WholeToothCode.codes["rct"] {
		t.Error("expected rct in codes")
	}
	if tooth.GeneralNotes.value != "Test notes" {
		t.Errorf("expected 'Test notes', got '%s'", tooth.GeneralNotes.value)
	}
	if _, exists := tooth.Surfaces["O"]; !exists {
		t.Error("surface O not found")
	}
	if tooth.Surfaces["O"].code != "car" {
		t.Errorf("expected code 'car', got '%s'", tooth.Surfaces["O"].code)
	}
	if tooth.Surfaces["O"].notes != "Cavity" {
		t.Errorf("expected notes 'Cavity', got '%s'", tooth.Surfaces["O"].notes)
	}
}

func TestSnapshotBuilder_FromSnapshot_ThenBuild(t *testing.T) {
	// Load existing snapshot
	snapshot := &model.OdontogramSnapshot{
		Teeth: map[string]model.ToothData{
			"11": {
				ID:             "11",
				WholeToothCode: []string{"amf"},
				GeneralNotes:   "Original note",
			},
		},
	}

	// New event to add
	noteData := model.OdontogramEventData{
		GeneralNotes: "Updated note",
	}
	noteDataJSON, _ := json.Marshal(noteData)

	newEvent := model.HstOdontogram{
		EventID:          "event-new",
		ToothID:          "11",
		EventType:        constant.EventTypeToothGeneralNoteUpdate,
		EventData:        noteDataJSON,
		SequenceNumber:   10,
		LogicalTimestamp: 1000,
		CreatedByStaffID: 1,
	}

	builder := NewSnapshotBuilder([]model.HstOdontogram{newEvent})
	builder.FromSnapshot(snapshot)
	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := result.Teeth["11"]
	if !contains(tooth.WholeToothCode, "amf") {
		t.Error("expected amf preserved from snapshot")
	}
	if tooth.GeneralNotes != "Updated note" {
		t.Errorf("expected 'Updated note', got '%s'", tooth.GeneralNotes)
	}
}

func TestSnapshotBuilder_Build_InvalidEventData(t *testing.T) {
	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        json.RawMessage("invalid json"),
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	_, err := builder.Build()
	if err == nil {
		t.Error("expected error for invalid event data")
	}
	if err != nil && err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}

func TestGetConditionName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"car", "Caries"},
		{"amf", "Amalgam Filling"},
		{"rct", "rct"}, // Not in map, returns code
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := getConditionName(tt.code)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetConditionColor(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"car", "#ff0000"},
		{"amf", "#000000"},
		{"unknown", "#ffffff"}, // Default
		{"", "#ffffff"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := getConditionColor(tt.code)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetConditionPattern(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"car", "outline"},
		{"amf", "solid"},
		{"tmf", "diagonal"},
		{"unknown", "solid"}, // Default
		{"", "solid"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := getConditionPattern(tt.code)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestSnapshotBuilder_Build_RemoveAfterInsert(t *testing.T) {
	// Test that insert after remove works correctly
	insert1Data := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	insert1JSON, _ := json.Marshal(insert1Data)

	removeData := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	removeJSON, _ := json.Marshal(removeData)

	insert2Data := model.OdontogramEventData{
		WholeToothCode: []string{"amf"},
	}
	insert2JSON, _ := json.Marshal(insert2Data)

	events := []model.HstOdontogram{
		{
			EventID:          "event-1",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        insert1JSON,
			SequenceNumber:   1,
			LogicalTimestamp: 100,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-2",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeRemove,
			EventData:        removeJSON,
			SequenceNumber:   2,
			LogicalTimestamp: 200,
			CreatedByStaffID: 1,
		},
		{
			EventID:          "event-3",
			ToothID:          "11",
			EventType:        constant.EventTypeToothCodeInsert,
			EventData:        insert2JSON,
			SequenceNumber:   3,
			LogicalTimestamp: 300, // After remove, should be added
			CreatedByStaffID: 1,
		},
	}

	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	tooth := snapshot.Teeth["11"]
	if !contains(tooth.WholeToothCode, "amf") {
		t.Error("expected amf to be present after re-insert")
	}
}

func TestSnapshotBuilder_Build_EmptyEvents(t *testing.T) {
	builder := NewSnapshotBuilder([]model.HstOdontogram{})
	snapshot, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if snapshot == nil {
		t.Fatal("snapshot is nil")
	}
	if len(snapshot.Teeth) != 0 {
		t.Errorf("expected empty Teeth, got %d", len(snapshot.Teeth))
	}
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
