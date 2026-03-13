package satusehat

// Procedure resource for actions performed on patients
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/procedure/

// Procedure - An action that is being or was performed on a patient
type Procedure struct {
	DomainResource
	Identifier            []Identifier          `json:"identifier,omitempty"`            // External Identifiers for this procedure
	InstantiatesCanonical []string              `json:"instantiatesCanonical,omitempty"` // Instantiates FHIR protocol or definition
	InstantiatesURI       []string              `json:"instantiatesUri,omitempty"`       // Instantiates external protocol or definition
	BasedOn               []Reference           `json:"basedOn,omitempty"`               // A request for this procedure
	PartOf                []Reference           `json:"partOf,omitempty"`                // Part of referenced event
	Status                string                `json:"status"`                          // preparation | in-progress | not-done | suspended | aborted | completed | entered-in-error | unknown (REQUIRED)
	StatusReason          *CodeableConcept      `json:"statusReason,omitempty"`          // Reason for current status
	Category              *CodeableConcept      `json:"category,omitempty"`              // Classification of the procedure
	Code                  *CodeableConcept      `json:"code,omitempty"`                  // Identification of the procedure
	Subject               *Reference            `json:"subject"`                         // Who the procedure was performed on (REQUIRED)
	Encounter             *Reference            `json:"encounter,omitempty"`             // Encounter created as part of
	PerformedDateTime     string                `json:"performedDateTime,omitempty"`     // When the procedure was performed
	PerformedPeriod       *Period               `json:"performedPeriod,omitempty"`       // When the procedure was performed
	PerformedString       string                `json:"performedString,omitempty"`       // When the procedure was performed
	PerformedAge          *Age                  `json:"performedAge,omitempty"`          // When the procedure was performed
	PerformedRange        *Range                `json:"performedRange,omitempty"`        // When the procedure was performed
	Recorder              *Reference            `json:"recorder,omitempty"`              // Who recorded the procedure
	Asserter              *Reference            `json:"asserter,omitempty"`              // Person who asserts this procedure
	Performer             []ProcedurePerformer  `json:"performer,omitempty"`             // The people who performed the procedure
	Location              *Reference            `json:"location,omitempty"`              // Where the procedure happened
	ReasonCode            []CodeableConcept     `json:"reasonCode,omitempty"`            // Coded reason procedure performed
	ReasonReference       []Reference           `json:"reasonReference,omitempty"`       // The justification that the procedure was performed
	BodySite              []CodeableConcept     `json:"bodySite,omitempty"`              // Target body sites
	Outcome               *CodeableConcept      `json:"outcome,omitempty"`               // The result of procedure
	Report                []Reference           `json:"report,omitempty"`                // Any report resulting from the procedure
	Complication          []CodeableConcept     `json:"complication,omitempty"`          // Complication following the procedure
	ComplicationDetail    []Reference           `json:"complicationDetail,omitempty"`    // A condition that is a result of the procedure
	FollowUp              []CodeableConcept     `json:"followUp,omitempty"`              // Instructions for follow up
	Note                  []Annotation          `json:"note,omitempty"`                  // Additional information about the procedure
	FocalDevice           []ProcedureFocalDevice `json:"focalDevice,omitempty"`          // Manipulated, implanted, or removed device
	UsedReference         []Reference           `json:"usedReference,omitempty"`         // Items used during procedure
	UsedCode              []CodeableConcept     `json:"usedCode,omitempty"`              // Coded items used during procedure
}

// ProcedurePerformer - The people who performed the procedure
type ProcedurePerformer struct {
	BackboneElement
	Function    *CodeableConcept `json:"function,omitempty"`    // Type of performance
	Actor       *Reference       `json:"actor"`                 // The reference to the practitioner (REQUIRED)
	OnBehalfOf  *Reference       `json:"onBehalfOf,omitempty"`  // Organization the device or practitioner was acting for
}

// ProcedureFocalDevice - Manipulated, implanted, or removed device
type ProcedureFocalDevice struct {
	BackboneElement
	Action      *CodeableConcept `json:"action,omitempty"`      // Kind of change to device
	Manipulated *Reference       `json:"manipulated"`           // Device that was changed (REQUIRED)
}
