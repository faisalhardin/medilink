package satusehat

// Condition resource for diagnoses and health conditions
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/condition/

// Condition - A clinical condition, problem, diagnosis, or other event
type Condition struct {
	DomainResource
	Identifier         []Identifier        `json:"identifier,omitempty"`         // External Ids for this condition
	ClinicalStatus     *CodeableConcept    `json:"clinicalStatus,omitempty"`     // active | recurrence | relapse | inactive | remission | resolved
	VerificationStatus *CodeableConcept    `json:"verificationStatus,omitempty"` // unconfirmed | provisional | differential | confirmed | refuted | entered-in-error
	Category           []CodeableConcept   `json:"category,omitempty"`           // problem-list-item | encounter-diagnosis
	Severity           *CodeableConcept    `json:"severity,omitempty"`           // Subjective severity of condition
	Code               *CodeableConcept    `json:"code"`                         // Identification of the condition, problem or diagnosis (REQUIRED)
	BodySite           []CodeableConcept   `json:"bodySite,omitempty"`           // Anatomical location, if relevant
	Subject            *Reference          `json:"subject"`                      // Who has the condition? (REQUIRED)
	Encounter          *Reference          `json:"encounter"`                    // Encounter created as part of (REQUIRED)
	OnsetDateTime      string              `json:"onsetDateTime,omitempty"`      // Estimated or actual date, date-time, or age
	OnsetAge           *Age                `json:"onsetAge,omitempty"`           // Estimated or actual date, date-time, or age
	OnsetPeriod        *Period             `json:"onsetPeriod,omitempty"`        // Estimated or actual date, date-time, or age
	OnsetRange         *Range              `json:"onsetRange,omitempty"`         // Estimated or actual date, date-time, or age
	OnsetString        string              `json:"onsetString,omitempty"`        // Estimated or actual date, date-time, or age
	AbatementDateTime  string              `json:"abatementDateTime,omitempty"`  // When in resolution/remission
	AbatementAge       *Age                `json:"abatementAge,omitempty"`       // When in resolution/remission
	AbatementPeriod    *Period             `json:"abatementPeriod,omitempty"`    // When in resolution/remission
	AbatementRange     *Range              `json:"abatementRange,omitempty"`     // When in resolution/remission
	AbatementString    string              `json:"abatementString,omitempty"`    // When in resolution/remission
	RecordedDate       string              `json:"recordedDate,omitempty"`       // Date record was first recorded
	Recorder           *Reference          `json:"recorder,omitempty"`           // Who recorded the condition
	Asserter           *Reference          `json:"asserter,omitempty"`           // Person who asserts this condition
	Stage              []ConditionStage    `json:"stage,omitempty"`              // Stage/grade, usually assessed formally
	Evidence           []ConditionEvidence `json:"evidence,omitempty"`           // Supporting evidence
	Note               []Annotation        `json:"note,omitempty"`               // Additional information about the Condition
}

// ConditionStage - Stage/grade, usually assessed formally
type ConditionStage struct {
	BackboneElement
	Summary    *CodeableConcept `json:"summary,omitempty"`    // Simple summary (disease specific)
	Assessment []Reference      `json:"assessment,omitempty"` // Formal record of assessment
	Type       *CodeableConcept `json:"type,omitempty"`       // Kind of staging
}

// ConditionEvidence - Supporting evidence
type ConditionEvidence struct {
	BackboneElement
	Code   []CodeableConcept `json:"code,omitempty"`   // Manifestation/symptom
	Detail []Reference       `json:"detail,omitempty"` // Supporting information found elsewhere
}
