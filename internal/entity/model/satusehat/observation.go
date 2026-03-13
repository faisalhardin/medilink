package satusehat

// Observation resource for measurements and assessments
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/observation/

// Observation - Measurements and simple assertions made about a patient
type Observation struct {
	DomainResource
	Identifier          []Identifier              `json:"identifier,omitempty"`          // Business Identifier for observation
	BasedOn             []Reference               `json:"basedOn,omitempty"`             // Fulfills plan, proposal or order
	PartOf              []Reference               `json:"partOf,omitempty"`              // Part of referenced event
	Status              string                    `json:"status"`                        // registered | preliminary | final | amended (REQUIRED)
	Category            []CodeableConcept         `json:"category,omitempty"`            // Classification of type of observation
	Code                *CodeableConcept          `json:"code"`                          // Type of observation (code / type) (REQUIRED)
	Subject             *Reference                `json:"subject"`                       // Who and/or what the observation is about (REQUIRED)
	Focus               []Reference               `json:"focus,omitempty"`               // What the observation is about, when it is not about the subject
	Encounter           *Reference                `json:"encounter"`                     // Healthcare event during which this observation is made (REQUIRED)
	EffectiveDateTime   string                    `json:"effectiveDateTime,omitempty"`   // Clinically relevant time/time-period for observation
	EffectivePeriod     *Period                   `json:"effectivePeriod,omitempty"`     // Clinically relevant time/time-period for observation
	EffectiveTiming     *Timing                   `json:"effectiveTiming,omitempty"`     // Clinically relevant time/time-period for observation
	EffectiveInstant    string                    `json:"effectiveInstant,omitempty"`    // Clinically relevant time/time-period for observation
	Issued              string                    `json:"issued,omitempty"`              // Date/Time this version was made available
	Performer           []Reference               `json:"performer,omitempty"`           // Who is responsible for the observation
	ValueQuantity       *Quantity                 `json:"valueQuantity,omitempty"`       // Actual result - Quantity
	ValueCodeableConcept *CodeableConcept         `json:"valueCodeableConcept,omitempty"` // Actual result - CodeableConcept
	ValueString         string                    `json:"valueString,omitempty"`         // Actual result - string
	ValueBoolean        *bool                     `json:"valueBoolean,omitempty"`        // Actual result - boolean
	ValueInteger        *int                      `json:"valueInteger,omitempty"`        // Actual result - integer
	ValueRange          *Range                    `json:"valueRange,omitempty"`          // Actual result - Range
	ValueRatio          *Ratio                    `json:"valueRatio,omitempty"`          // Actual result - Ratio
	ValueSampledData    *SampledData              `json:"valueSampledData,omitempty"`    // Actual result - SampledData
	ValueTime           string                    `json:"valueTime,omitempty"`           // Actual result - time
	ValueDateTime       string                    `json:"valueDateTime,omitempty"`       // Actual result - dateTime
	ValuePeriod         *Period                   `json:"valuePeriod,omitempty"`         // Actual result - Period
	DataAbsentReason    *CodeableConcept          `json:"dataAbsentReason,omitempty"`    // Why the result is missing
	Interpretation      []CodeableConcept         `json:"interpretation,omitempty"`      // High, low, normal, etc.
	Note                []Annotation              `json:"note,omitempty"`                // Comments about the observation
	BodySite            *CodeableConcept          `json:"bodySite,omitempty"`            // Observed body part
	Method              *CodeableConcept          `json:"method,omitempty"`              // How it was done
	Specimen            *Reference                `json:"specimen,omitempty"`            // Specimen used for this observation
	Device              *Reference                `json:"device,omitempty"`              // (Measurement) Device
	ReferenceRange      []ObservationReferenceRange `json:"referenceRange,omitempty"`   // Provides guide for interpretation
	HasMember           []Reference               `json:"hasMember,omitempty"`           // Related resource that belongs to the Observation group
	DerivedFrom         []Reference               `json:"derivedFrom,omitempty"`         // Related measurements the observation is made from
	Component           []ObservationComponent    `json:"component,omitempty"`           // Component results
}

// ObservationReferenceRange - Provides guide for interpretation
type ObservationReferenceRange struct {
	BackboneElement
	Low       *SimpleQuantity    `json:"low,omitempty"`       // Low Range, if relevant
	High      *SimpleQuantity    `json:"high,omitempty"`      // High Range, if relevant
	Type      *CodeableConcept   `json:"type,omitempty"`      // Reference range qualifier
	AppliesTo []CodeableConcept  `json:"appliesTo,omitempty"` // Reference range population
	Age       *Range             `json:"age,omitempty"`       // Applicable age range, if relevant
	Text      string             `json:"text,omitempty"`      // Text based reference range in an observation
}

// ObservationComponent - Component results
type ObservationComponent struct {
	BackboneElement
	Code                 *CodeableConcept             `json:"code"`                           // Type of component observation (code / type) (REQUIRED)
	ValueQuantity        *Quantity                    `json:"valueQuantity,omitempty"`        // Actual component result - Quantity
	ValueCodeableConcept *CodeableConcept             `json:"valueCodeableConcept,omitempty"` // Actual component result - CodeableConcept
	ValueString          string                       `json:"valueString,omitempty"`          // Actual component result - string
	ValueBoolean         *bool                        `json:"valueBoolean,omitempty"`         // Actual component result - boolean
	ValueInteger         *int                         `json:"valueInteger,omitempty"`         // Actual component result - integer
	ValueRange           *Range                       `json:"valueRange,omitempty"`           // Actual component result - Range
	ValueRatio           *Ratio                       `json:"valueRatio,omitempty"`           // Actual component result - Ratio
	ValueSampledData     *SampledData                 `json:"valueSampledData,omitempty"`     // Actual component result - SampledData
	ValueTime            string                       `json:"valueTime,omitempty"`            // Actual component result - time
	ValueDateTime        string                       `json:"valueDateTime,omitempty"`        // Actual component result - dateTime
	ValuePeriod          *Period                      `json:"valuePeriod,omitempty"`          // Actual component result - Period
	DataAbsentReason     *CodeableConcept             `json:"dataAbsentReason,omitempty"`     // Why the component result is missing
	Interpretation       []CodeableConcept            `json:"interpretation,omitempty"`       // High, low, normal, etc.
	ReferenceRange       []ObservationReferenceRange  `json:"referenceRange,omitempty"`       // Provides guide for interpretation of component result
}
