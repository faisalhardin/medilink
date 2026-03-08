package satusehat

// Encounter resource for patient visits
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/encounter/

// Encounter - An interaction between a patient and healthcare provider(s)
type Encounter struct {
	DomainResource
	Identifier      []Identifier              `json:"identifier,omitempty"`      // Identifier(s) by which this encounter is known (REQUIRED)
	Status          string                    `json:"status"`                    // planned | arrived | triaged | in-progress | onleave | finished | cancelled (REQUIRED)
	StatusHistory   []EncounterStatusHistory  `json:"statusHistory,omitempty"`   // List of past encounter statuses (REQUIRED)
	Class           *Coding                   `json:"class"`                     // Classification of patient encounter (REQUIRED)
	ClassHistory    []EncounterClassHistory   `json:"classHistory,omitempty"`    // List of past encounter classes (REQUIRED)
	Type            []CodeableConcept         `json:"type,omitempty"`            // Specific type of encounter
	ServiceType     *CodeableConcept          `json:"serviceType,omitempty"`     // Specific type of service
	Priority        *CodeableConcept          `json:"priority,omitempty"`        // Indicates the urgency of the encounter
	Subject         *Reference                `json:"subject"`                   // The patient present at the encounter (REQUIRED)
	EpisodeOfCare   []Reference               `json:"episodeOfCare,omitempty"`   // Episode(s) of care that this encounter should be recorded against
	BasedOn         []Reference               `json:"basedOn,omitempty"`         // The ServiceRequest that initiated this encounter
	Participant     []EncounterParticipant    `json:"participant,omitempty"`     // List of participants involved in the encounter
	Appointment     []Reference               `json:"appointment,omitempty"`     // The appointment that scheduled this encounter
	Period          *Period                   `json:"period"`                    // The start and end time of the encounter (REQUIRED)
	Length          *Duration                 `json:"length,omitempty"`          // Quantity of time the encounter lasted
	ReasonCode      []CodeableConcept         `json:"reasonCode,omitempty"`      // Coded reason the encounter takes place (REQUIRED)
	ReasonReference []Reference               `json:"reasonReference,omitempty"` // Reason the encounter takes place (reference)
	Diagnosis       []EncounterDiagnosis      `json:"diagnosis,omitempty"`       // The list of diagnosis relevant to this encounter (REQUIRED)
	Account         []Reference               `json:"account,omitempty"`         // The set of accounts that may be used for billing
	Hospitalization *EncounterHospitalization `json:"hospitalization,omitempty"` // Details about the admission to a healthcare service
	Location        []EncounterLocation       `json:"location,omitempty"`        // List of locations where the patient has been (REQUIRED)
	ServiceProvider *Reference                `json:"serviceProvider,omitempty"` // The organization (facility) responsible for this encounter (REQUIRED)
	PartOf          *Reference                `json:"partOf,omitempty"`          // Another Encounter this encounter is part of
}

// EncounterStatusHistory - List of past encounter statuses
type EncounterStatusHistory struct {
	BackboneElement
	Status string  `json:"status"` // planned | arrived | triaged | in-progress | onleave | finished | cancelled (REQUIRED)
	Period *Period `json:"period"` // The time that the episode was in the specified status (REQUIRED)
}

// EncounterClassHistory - List of past encounter classes
type EncounterClassHistory struct {
	BackboneElement
	Class  *Coding `json:"class"`  // inpatient | outpatient | ambulatory | emergency (REQUIRED)
	Period *Period `json:"period"` // The time that the episode was in the specified class (REQUIRED)
}

// EncounterParticipant - List of participants involved in the encounter
type EncounterParticipant struct {
	BackboneElement
	Type       []CodeableConcept `json:"type,omitempty"`       // Role of participant in encounter
	Period     *Period           `json:"period,omitempty"`     // Period of time during the encounter that the participant participated
	Individual *Reference        `json:"individual,omitempty"` // Persons involved in the encounter other than the patient
}

// EncounterDiagnosis - The list of diagnosis relevant to this encounter
type EncounterDiagnosis struct {
	BackboneElement
	Condition *Reference       `json:"condition"` // The diagnosis or procedure relevant to the encounter (REQUIRED)
	Use       *CodeableConcept `json:"use"`       // Role that this diagnosis has within the encounter (e.g. admission, billing, discharge) (REQUIRED)
	Rank      *int             `json:"rank"`      // Ranking of the diagnosis (for each role type) (REQUIRED)
}

// EncounterHospitalization - Details about the admission to a healthcare service
type EncounterHospitalization struct {
	BackboneElement
	PreAdmissionIdentifier *Identifier       `json:"preAdmissionIdentifier,omitempty"` // Pre-admission identifier
	Origin                 *Reference        `json:"origin,omitempty"`                 // The location/organization from which the patient came before admission
	AdmitSource            *CodeableConcept  `json:"admitSource,omitempty"`            // From where patient was admitted (physician referral, transfer)
	ReAdmission            *CodeableConcept  `json:"reAdmission,omitempty"`            // The type of hospital re-admission that has occurred
	DietPreference         []CodeableConcept `json:"dietPreference,omitempty"`         // Diet preferences reported by the patient
	SpecialCourtesy        []CodeableConcept `json:"specialCourtesy,omitempty"`        // Special courtesies (VIP, board member)
	SpecialArrangement     []CodeableConcept `json:"specialArrangement,omitempty"`     // Wheelchair, translator, stretcher, etc.
	Destination            *Reference        `json:"destination,omitempty"`            // Location/organization to which the patient is discharged
	DischargeDisposition   *CodeableConcept  `json:"dischargeDisposition,omitempty"`   // Category or kind of location after discharge
}

// EncounterLocation - List of locations where the patient has been
type EncounterLocation struct {
	BackboneElement
	Location      *Reference `json:"location"`           // Location the encounter takes place (REQUIRED)
	Status        string     `json:"status,omitempty"`   // planned | active | reserved | completed
	PhysicalType  *CodeableConcept `json:"physicalType,omitempty"` // The physical type of the location
	Period        *Period    `json:"period,omitempty"`   // Time period during which the patient was present at the location
}
