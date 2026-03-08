package satusehat

// Composition resource for clinical documents
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/composition/

// Composition - A set of healthcare-related information that is assembled together
type Composition struct {
	DomainResource
	Identifier      *Identifier            `json:"identifier,omitempty"`      // Version-independent identifier for the Composition
	Status          string                 `json:"status"`                    // preliminary | final | amended | entered-in-error (REQUIRED)
	Type            *CodeableConcept       `json:"type"`                      // Kind of composition (LOINC if possible) (REQUIRED)
	Category        []CodeableConcept      `json:"category,omitempty"`        // Categorization of Composition
	Subject         *Reference             `json:"subject"`                   // Who and/or what the composition is about (REQUIRED)
	Encounter       *Reference             `json:"encounter,omitempty"`       // Context of the Composition
	Date            string                 `json:"date"`                      // Composition editing time (REQUIRED)
	Author          []Reference            `json:"author"`                    // Who and/or what authored the composition (REQUIRED)
	Title           string                 `json:"title"`                     // Human Readable name/title (REQUIRED)
	Confidentiality string                 `json:"confidentiality,omitempty"` // As defined by affinity domain
	Attester        []CompositionAttester  `json:"attester,omitempty"`        // Attests to accuracy of composition
	Custodian       *Reference             `json:"custodian,omitempty"`       // Organization which maintains the composition
	RelatesTo       []CompositionRelatesTo `json:"relatesTo,omitempty"`       // Relationships to other compositions/documents
	Event           []CompositionEvent     `json:"event,omitempty"`           // The clinical service(s) being documented
	Section         []CompositionSection   `json:"section,omitempty"`         // Composition is broken into sections
}

// CompositionAttester - Attests to accuracy of composition
type CompositionAttester struct {
	BackboneElement
	Mode  string     `json:"mode"`            // personal | professional | legal | official (REQUIRED)
	Time  string     `json:"time,omitempty"`  // When the composition was attested
	Party *Reference `json:"party,omitempty"` // Who attested the composition
}

// CompositionRelatesTo - Relationships to other compositions/documents
type CompositionRelatesTo struct {
	BackboneElement
	Code             string      `json:"code"`                       // replaces | transforms | signs | appends (REQUIRED)
	TargetIdentifier *Identifier `json:"targetIdentifier,omitempty"` // Target of the relationship
	TargetReference  *Reference  `json:"targetReference,omitempty"`  // Target of the relationship
}

// CompositionEvent - The clinical service(s) being documented
type CompositionEvent struct {
	BackboneElement
	Code   []CodeableConcept `json:"code,omitempty"`   // Code(s) that apply to the event being documented
	Period *Period           `json:"period,omitempty"` // The period covered by the documentation
	Detail []Reference       `json:"detail,omitempty"` // The event(s) being documented
}

// CompositionSection - Composition is broken into sections
type CompositionSection struct {
	BackboneElement
	Title       string                `json:"title,omitempty"`       // Label for section (e.g. for ToC)
	Code        *CodeableConcept      `json:"code,omitempty"`        // Classification of section (recommended)
	Author      []Reference           `json:"author,omitempty"`      // Who and/or what authored the section
	Focus       *Reference            `json:"focus,omitempty"`       // Who/what the section is about, when it is not about the subject
	Text        *Narrative            `json:"text,omitempty"`        // Text summary of the section, for human interpretation
	Mode        string                `json:"mode,omitempty"`        // working | snapshot | changes
	OrderedBy   *CodeableConcept      `json:"orderedBy,omitempty"`   // Order of section entries
	Entry       []Reference           `json:"entry,omitempty"`       // A reference to data that supports this section
	EmptyReason *CodeableConcept      `json:"emptyReason,omitempty"` // Why the section is empty
	Section     []CompositionSection  `json:"section,omitempty"`     // Nested Section
}
