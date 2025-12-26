package odontogram

import (
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
)

// ValidateEventRequest validates a single event request
func ValidateEventRequest(req model.CreateOdontogramEventRequest) error {
	// Validate required fields
	if req.PatientUUID == "" {
		return commonerr.SetNewBadRequest("patient_uuid is required", "Patient UUID must be provided")
	}

	if req.ToothID == "" {
		return commonerr.SetNewBadRequest("tooth_id is required", "Tooth ID must be provided")
	}

	if req.EventType == "" {
		return commonerr.SetNewBadRequest("event_type is required", "Event type must be provided")
	}

	// Validate tooth ID
	if !constant.ValidToothNumbers[req.ToothID] {
		return commonerr.SetNewBadRequest(
			"invalid_tooth_id",
			fmt.Sprintf("Invalid tooth ID: %s. Must be a valid FDI tooth number (11-18, 21-28, 31-38, 41-48)", req.ToothID),
		)
	}

	// Validate event type and corresponding data
	switch req.EventType {
	case constant.EventTypeToothCodeInsert, constant.EventTypeToothCodeRemove:
		if len(req.EventData.WholeToothCode) == 0 {
			return commonerr.SetNewBadRequest(
				"invalid_event_data",
				fmt.Sprintf("Event type %s requires whole_tooth_code in event_data", req.EventType),
			)
		}
		// Validate tooth codes
		// for _, code := range req.EventData.WholeToothCode {
		// 	if !constant.ValidWholeToothCodes[code] {
		// 		return commonerr.SetNewBadRequest(
		// 			"invalid_tooth_code",
		// 			fmt.Sprintf("Invalid whole tooth code: %s", code),
		// 		)
		// 	}
		// }

	case constant.EventTypeToothGeneralNoteUpdate:
		// No specific validation needed, empty string is allowed to clear notes

	case constant.EventTypeToothSurfaceCodeSet:
		if req.EventData.Surface == "" {
			return commonerr.SetNewBadRequest(
				"invalid_event_data",
				"Event type tooth_surface_code_set requires surface in event_data",
			)
		}
		if req.EventData.SurfaceCode == "" {
			return commonerr.SetNewBadRequest(
				"invalid_event_data",
				"Event type tooth_surface_code_set requires surface_code in event_data",
			)
		}
		// Validate surface
		if !constant.ValidSurfaceCodes[req.EventData.Surface] {
			return commonerr.SetNewBadRequest(
				"invalid_surface",
				fmt.Sprintf("Invalid surface code: %s. Must be one of M, D, L, O, V", req.EventData.Surface),
			)
		}
		// Validate surface code
		// if !constant.ValidSurfaceTreatmentCodes[req.EventData.SurfaceCode] {
		// 	return commonerr.SetNewBadRequest(
		// 		"invalid_surface_code",
		// 		fmt.Sprintf("Invalid surface treatment code: %s", req.EventData.SurfaceCode),
		// 	)
		// }

	case constant.EventTypeToothSurfaceCodeRemove:
		if req.EventData.Surface == "" {
			return commonerr.SetNewBadRequest(
				"invalid_event_data",
				"Event type tooth_surface_code_remove requires surface in event_data",
			)
		}
		// Validate surface
		if !constant.ValidSurfaceCodes[req.EventData.Surface] {
			return commonerr.SetNewBadRequest(
				"invalid_surface",
				fmt.Sprintf("Invalid surface code: %s. Must be one of M, D, B, L, O, I, V", req.EventData.Surface),
			)
		}

	case constant.EventTypeToothSurfaceNoteUpdate:
		if req.EventData.Surface == "" {
			return commonerr.SetNewBadRequest(
				"invalid_event_data",
				"Event type tooth_surface_note_update requires surface in event_data",
			)
		}
		// Validate surface
		if !constant.ValidSurfaceCodes[req.EventData.Surface] {
			return commonerr.SetNewBadRequest(
				"invalid_surface",
				fmt.Sprintf("Invalid surface code: %s. Must be one of M, D, B, L, O, I, V", req.EventData.Surface),
			)
		}

	case constant.EventTypeToothReset:
		// No specific validation needed

	default:
		return commonerr.SetNewBadRequest(
			"invalid_event_type",
			fmt.Sprintf("Invalid event type: %s", req.EventType),
		)
	}

	return nil
}

// ValidateGetEventsParams validates parameters for getting events
func ValidateGetEventsParams(params model.GetOdontogramEventsParams) error {
	if params.PatientUUID == "" {
		return commonerr.SetNewBadRequest("patient_uuid is required", "Patient UUID must be provided")
	}

	// Validate tooth ID if provided
	if params.ToothID != "" && !constant.ValidToothNumbers[params.ToothID] {
		return commonerr.SetNewBadRequest(
			"invalid tooth id",
			fmt.Sprintf("Invalid tooth ID: %s", params.ToothID),
		)
	}

	return nil
}
