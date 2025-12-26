package odontogram

import (
	"encoding/json"
	"sort"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/pkg/errors"
)

// SnapshotBuilder implements CRDT-based snapshot building from events
type SnapshotBuilder struct {
	events []model.HstOdontogram
	teeth  map[string]*toothState
}

// toothState maintains the state of a single tooth during snapshot building
type toothState struct {
	ID             string
	WholeToothCode *wholeToothCodeState
	GeneralNotes   *textState
	Surfaces       map[string]*surfaceState
}

// wholeToothCodeState implements Set CRDT for whole tooth codes
type wholeToothCodeState struct {
	codes   map[string]bool
	removes map[string]int64 // Track removal timestamps
}

// textState implements Last-Write-Wins CRDT for text fields
type textState struct {
	value            string
	logicalTimestamp int64
	createdByStaffID int64
}

// surfaceState implements Map CRDT for surface data
type surfaceState struct {
	surface          string
	code             string
	notes            string
	logicalTimestamp int64
	createdByStaffID int64
}

// NewSnapshotBuilder creates a new snapshot builder
func NewSnapshotBuilder(events []model.HstOdontogram) *SnapshotBuilder {
	// Filter duplicate events by event_id (only if event_id is not empty)
	filteredEvents := make([]model.HstOdontogram, 0, len(events))
	seenEventIDs := make(map[string]bool)

	for _, event := range events {
		// If event_id is empty, don't filter - include it
		if event.EventID == "" {
			filteredEvents = append(filteredEvents, event)
			continue
		}

		// If event_id is not empty, check for duplicates
		if !seenEventIDs[event.EventID] {
			seenEventIDs[event.EventID] = true
			filteredEvents = append(filteredEvents, event)
		}
		// If duplicate found, skip it (keep first occurrence)
	}

	return &SnapshotBuilder{
		events: filteredEvents,
		teeth:  make(map[string]*toothState),
	}
}

func (sb *SnapshotBuilder) SetTeeth(teeth map[string]*toothState) {
	sb.teeth = teeth
}

// Build constructs the snapshot using CRDT merge rules and sort the events by logical timestamp and created by staff id
func (sb *SnapshotBuilder) Build() (*model.OdontogramSnapshot, error) {
	// Sort events by (logical_timestamp, created_by_staff_id) for deterministic ordering
	sort.Slice(sb.events, func(i, j int) bool {
		if sb.events[i].LogicalTimestamp != sb.events[j].LogicalTimestamp {
			return sb.events[i].LogicalTimestamp < sb.events[j].LogicalTimestamp
		}
		return sb.events[i].CreatedByStaffID < sb.events[j].CreatedByStaffID
	})

	// Apply each event
	for _, event := range sb.events {
		if err := sb.applyEvent(event); err != nil {
			return nil, err
		}
	}

	// Convert to snapshot format
	return sb.toSnapshot(), nil
}

// applyEvent applies a single event using CRDT rules
func (sb *SnapshotBuilder) applyEvent(event model.HstOdontogram) error {
	// Parse event data
	var eventData model.OdontogramEventData
	if err := json.Unmarshal([]byte(event.EventData), &eventData); err != nil {
		return errors.Wrap(err, "failed to parse event data")
	}

	// Ensure tooth exists
	if _, exists := sb.teeth[event.ToothID]; !exists {
		sb.teeth[event.ToothID] = &toothState{
			ID:       event.ToothID,
			Surfaces: make(map[string]*surfaceState),
			WholeToothCode: &wholeToothCodeState{
				codes:   make(map[string]bool),
				removes: make(map[string]int64),
			},
			GeneralNotes: &textState{},
		}
	}

	tooth := sb.teeth[event.ToothID]

	// Apply event based on type
	switch event.EventType {
	case constant.EventTypeToothCodeInsert:
		sb.applyToothCodeInsert(tooth, eventData, event)

	case constant.EventTypeToothCodeRemove:
		sb.applyToothCodeRemove(tooth, eventData, event)

	case constant.EventTypeToothGeneralNoteUpdate:
		sb.applyGeneralNoteUpdate(tooth, eventData, event)

	case constant.EventTypeToothSurfaceCodeSet:
		sb.applySurfaceCodeSet(tooth, eventData, event)

	case constant.EventTypeToothSurfaceCodeRemove:
		sb.applySurfaceCodeRemove(tooth, eventData, event)

	case constant.EventTypeToothSurfaceNoteUpdate:
		sb.applySurfaceNoteUpdate(tooth, eventData, event)

	case constant.EventTypeToothReset:
		sb.applyToothReset(tooth)
	}

	return nil
}

// applyToothCodeInsert implements Set CRDT insert for whole tooth codes
func (sb *SnapshotBuilder) applyToothCodeInsert(tooth *toothState, eventData model.OdontogramEventData, event model.HstOdontogram) {
	for _, code := range eventData.WholeToothCode {
		// Only add if not removed, or if this insert is after the last remove
		if removedAt, wasRemoved := tooth.WholeToothCode.removes[code]; !wasRemoved || event.LogicalTimestamp > removedAt {
			tooth.WholeToothCode.codes[code] = true
		}
	}
}

// applyToothCodeRemove implements Set CRDT remove for whole tooth codes
func (sb *SnapshotBuilder) applyToothCodeRemove(tooth *toothState, eventData model.OdontogramEventData, event model.HstOdontogram) {
	for _, code := range eventData.WholeToothCode {
		tooth.WholeToothCode.removes[code] = event.LogicalTimestamp
		delete(tooth.WholeToothCode.codes, code)
	}
}

// applyGeneralNoteUpdate implements Last-Write-Wins CRDT for general notes
func (sb *SnapshotBuilder) applyGeneralNoteUpdate(tooth *toothState, eventData model.OdontogramEventData, event model.HstOdontogram) {
	// Update if this event is newer (higher timestamp or same timestamp but higher staff ID)
	if event.LogicalTimestamp > tooth.GeneralNotes.logicalTimestamp ||
		(event.LogicalTimestamp == tooth.GeneralNotes.logicalTimestamp &&
			event.CreatedByStaffID > tooth.GeneralNotes.createdByStaffID) {
		tooth.GeneralNotes.value = eventData.GeneralNotes
		tooth.GeneralNotes.logicalTimestamp = event.LogicalTimestamp
		tooth.GeneralNotes.createdByStaffID = event.CreatedByStaffID
	}
}

// applySurfaceCodeSet implements Map CRDT for surface codes
func (sb *SnapshotBuilder) applySurfaceCodeSet(tooth *toothState, eventData model.OdontogramEventData, event model.HstOdontogram) {
	surface := eventData.Surface
	existing, exists := tooth.Surfaces[surface]

	// Update if this event is newer or doesn't exist
	if !exists || event.LogicalTimestamp > existing.logicalTimestamp ||
		(event.LogicalTimestamp == existing.logicalTimestamp &&
			event.CreatedByStaffID > existing.createdByStaffID) {

		existingNotes := ""
		if exists {
			existingNotes = existing.notes
		}
		tooth.Surfaces[surface] = &surfaceState{
			surface:          surface,
			code:             eventData.SurfaceCode,
			notes:            existingNotes, // Preserve notes if they exist
			logicalTimestamp: event.LogicalTimestamp,
			createdByStaffID: event.CreatedByStaffID,
		}
		// Preserve notes timestamp if it's newer
		if exists && existing.logicalTimestamp > event.LogicalTimestamp {
			tooth.Surfaces[surface].notes = existing.notes
		}
	}
}

// applySurfaceCodeRemove removes a surface from the map
func (sb *SnapshotBuilder) applySurfaceCodeRemove(tooth *toothState, eventData model.OdontogramEventData, event model.HstOdontogram) {
	surface := eventData.Surface
	existing, exists := tooth.Surfaces[surface]

	// Only remove if this event is newer than existing
	if !exists || event.LogicalTimestamp > existing.logicalTimestamp ||
		(event.LogicalTimestamp == existing.logicalTimestamp &&
			event.CreatedByStaffID > existing.createdByStaffID) {
		delete(tooth.Surfaces, surface)
	}
}

// applySurfaceNoteUpdate implements Last-Write-Wins CRDT for surface notes
func (sb *SnapshotBuilder) applySurfaceNoteUpdate(tooth *toothState, eventData model.OdontogramEventData, event model.HstOdontogram) {
	surface := eventData.Surface
	existing, exists := tooth.Surfaces[surface]

	if !exists {
		// Create new surface entry with just notes
		tooth.Surfaces[surface] = &surfaceState{
			surface:          surface,
			notes:            eventData.SurfaceNotes,
			logicalTimestamp: event.LogicalTimestamp,
			createdByStaffID: event.CreatedByStaffID,
		}
	} else if event.LogicalTimestamp > existing.logicalTimestamp ||
		(event.LogicalTimestamp == existing.logicalTimestamp &&
			event.CreatedByStaffID > existing.createdByStaffID) {
		// Update notes if this event is newer
		existing.notes = eventData.SurfaceNotes
		// Don't update timestamp/staff if code was set more recently
	}
}

// applyToothReset clears all tooth data
func (sb *SnapshotBuilder) applyToothReset(tooth *toothState) {
	tooth.WholeToothCode.codes = make(map[string]bool)
	tooth.WholeToothCode.removes = make(map[string]int64)
	tooth.GeneralNotes = &textState{}
	tooth.Surfaces = make(map[string]*surfaceState)
}

// toSnapshot converts the internal state to the snapshot format
func (sb *SnapshotBuilder) toSnapshot() *model.OdontogramSnapshot {
	snapshot := &model.OdontogramSnapshot{
		Teeth: make(map[string]model.ToothData),
	}

	for toothID, tooth := range sb.teeth {
		// Convert whole tooth codes to slice
		codes := make([]string, 0, len(tooth.WholeToothCode.codes))
		for code := range tooth.WholeToothCode.codes {
			codes = append(codes, code)
		}
		sort.Strings(codes) // Sort for consistency

		// Convert surfaces to slice
		surfaces := make([]model.SurfaceData, 0, len(tooth.Surfaces))
		for _, surface := range tooth.Surfaces {
			surfaceData := model.SurfaceData{
				Surface:          surface.surface,
				Code:             surface.code,
				Notes:            surface.notes,
				LogicalTimestamp: surface.logicalTimestamp,
				CreatedByStaffID: surface.createdByStaffID,
			}
			// Add display properties based on code
			surfaceData.Condition = getConditionName(surface.code)
			surfaceData.Color = getConditionColor(surface.code)
			surfaceData.Pattern = getConditionPattern(surface.code)

			surfaces = append(surfaces, surfaceData)
		}

		// Sort surfaces for consistency
		sort.Slice(surfaces, func(i, j int) bool {
			return surfaces[i].Surface < surfaces[j].Surface
		})

		snapshot.Teeth[toothID] = model.ToothData{
			ID:             toothID,
			Surfaces:       surfaces,
			WholeToothCode: codes,
			GeneralNotes:   tooth.GeneralNotes.value,
		}
	}

	return snapshot
}

// FromSnapshot converts a public snapshot model to internal state
// This is useful for loading existing snapshots and continuing to build on them
func (sb *SnapshotBuilder) FromSnapshot(snapshot *model.OdontogramSnapshot) {
	sb.teeth = make(map[string]*toothState)

	for toothID, toothData := range snapshot.Teeth {
		// Create tooth state
		tooth := &toothState{
			ID:       toothID,
			Surfaces: make(map[string]*surfaceState),
			WholeToothCode: &wholeToothCodeState{
				codes:   make(map[string]bool),
				removes: make(map[string]int64),
			},
			GeneralNotes: &textState{},
		}

		// Convert whole tooth codes from slice to map
		for _, code := range toothData.WholeToothCode {
			tooth.WholeToothCode.codes[code] = true
		}

		// Convert general notes from string to textState
		if toothData.GeneralNotes != "" {
			tooth.GeneralNotes.value = toothData.GeneralNotes
			// Note: We don't have timestamp/staffID from snapshot, so we use defaults
			// This means the next update will overwrite regardless of timestamp
			tooth.GeneralNotes.logicalTimestamp = 0
			tooth.GeneralNotes.createdByStaffID = 0
		}

		// Convert surfaces from slice to map
		for _, surfaceData := range toothData.Surfaces {
			tooth.Surfaces[surfaceData.Surface] = &surfaceState{
				surface:          surfaceData.Surface,
				code:             surfaceData.Code,
				notes:            surfaceData.Notes,
				logicalTimestamp: surfaceData.LogicalTimestamp,
				createdByStaffID: surfaceData.CreatedByStaffID,
			}
		}

		sb.teeth[toothID] = tooth
	}
}

// getConditionName returns a human-readable name for a condition code
func getConditionName(code string) string {
	conditionNames := map[string]string{
		"car": "Caries",
		"amf": "Amalgam Filling",
		"gif": "Glass Ionomer",
		"rcf": "Resin Composite Filling",
		"cof": "Composite Filling",
		"tmf": "Temporary Filling",
		"sea": "Sealant",
		"abr": "Abrasion",
		"att": "Attrition",
		"ero": "Erosion",
		"fra": "Fracture",
	}
	if name, exists := conditionNames[code]; exists {
		return name
	}
	return code
}

// getConditionColor returns a color for a condition code
func getConditionColor(code string) string {
	colorMap := map[string]string{
		"car": "#ff0000", // Red for caries
		"amf": "#000000", // Black for amalgam
		"gif": "#16a34a", // Green for glass ionomer
		"rcf": "#ffffff", // White for resin composite
		"cof": "#ffffff", // White for composite
		"tmf": "#ffff00", // Yellow for temporary
		"sea": "#0000ff", // Blue for sealant
		"abr": "#ff9800", // Orange for abrasion
		"att": "#ff9800", // Orange for attrition
		"ero": "#ff9800", // Orange for erosion
		"fra": "#ff0000", // Red for fracture
	}
	if color, exists := colorMap[code]; exists {
		return color
	}
	return "#ffffff"
}

// getConditionPattern returns a pattern for a condition code
func getConditionPattern(code string) string {
	patternMap := map[string]string{
		"car": "outline",
		"amf": "solid",
		"gif": "solid",
		"rcf": "solid",
		"cof": "solid",
		"tmf": "diagonal",
		"sea": "dots",
		"abr": "diagonal",
		"att": "diagonal",
		"ero": "diagonal",
		"fra": "crosshatch",
	}
	if pattern, exists := patternMap[code]; exists {
		return pattern
	}
	return "solid"
}
