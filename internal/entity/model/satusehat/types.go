package satusehat

// Shared FHIR data types for Satu Sehat integration
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/data-type/general/

// Identifier - Unique identifier for resources
type Identifier struct {
	Use      string           `json:"use,omitempty"`      // usual | official | temp | secondary | old
	Type     *CodeableConcept `json:"type,omitempty"`     // Description of identifier
	System   string           `json:"system,omitempty"`   // Namespace for the identifier value
	Value    string           `json:"value,omitempty"`    // The value that is unique
	Period   *Period          `json:"period,omitempty"`   // Time period when id was valid
	Assigner *Reference       `json:"assigner,omitempty"` // Organization that issued id
}

// Coding - A reference to a code defined by a terminology system
type Coding struct {
	System       string `json:"system,omitempty"`       // Identity of the terminology system
	Version      string `json:"version,omitempty"`      // Version of the system
	Code         string `json:"code,omitempty"`         // Symbol in syntax defined by the system
	Display      string `json:"display,omitempty"`      // Representation defined by the system
	UserSelected *bool  `json:"userSelected,omitempty"` // If this coding was chosen directly by the user
}

// CodeableConcept - A concept that may be defined by a formal reference to a terminology
type CodeableConcept struct {
	Coding []Coding `json:"coding,omitempty"` // Code defined by a terminology system
	Text   string   `json:"text,omitempty"`   // Plain text representation
}

// Reference - A reference from one resource to another
type Reference struct {
	Reference  string      `json:"reference,omitempty"`  // Literal reference, Relative, internal or absolute URL
	Type       string      `json:"type,omitempty"`       // Type the reference refers to
	Identifier *Identifier `json:"identifier,omitempty"` // Logical reference, when literal reference is not known
	Display    string      `json:"display,omitempty"`    // Text alternative for the resource
}

// Extension - Additional content defined by implementations
type Extension struct {
	URL                string           `json:"url"`                          // identifies the meaning of the extension
	ValueString        string           `json:"valueString,omitempty"`        // Value of extension - must be one of these
	ValueCode          string           `json:"valueCode,omitempty"`          // Value as code
	ValueBoolean       *bool            `json:"valueBoolean,omitempty"`       // Value as boolean
	ValueInteger       *int             `json:"valueInteger,omitempty"`       // Value as integer
	ValueDecimal       *float64         `json:"valueDecimal,omitempty"`       // Value as decimal
	ValueAddress       *Address         `json:"valueAddress,omitempty"`       // Value as Address
	ValueCodeableConcept *CodeableConcept `json:"valueCodeableConcept,omitempty"` // Value as CodeableConcept
	ValueCoding        *Coding          `json:"valueCoding,omitempty"`        // Value as Coding
	ValueReference     *Reference       `json:"valueReference,omitempty"`     // Value as Reference
	Extension          []Extension      `json:"extension,omitempty"`          // Nested extensions
}

// Address - An address expressed using postal conventions
type Address struct {
	Use        string      `json:"use,omitempty"`        // home | work | temp | old | billing
	Type       string      `json:"type,omitempty"`       // postal | physical | both
	Text       string      `json:"text,omitempty"`       // Text representation of the address
	Line       []string    `json:"line,omitempty"`       // Street name, number, direction & P.O. Box etc.
	City       string      `json:"city,omitempty"`       // Name of city, town etc.
	District   string      `json:"district,omitempty"`   // District name (aka county)
	State      string      `json:"state,omitempty"`      // Sub-unit of country (abbreviations ok)
	PostalCode string      `json:"postalCode,omitempty"` // Postal code for area
	Country    string      `json:"country,omitempty"`    // Country (e.g. can be ISO 3166 2 or 3 letter code)
	Period     *Period     `json:"period,omitempty"`     // Time period when address was/is in use
	Extension  []Extension `json:"extension,omitempty"`  // AdministrativeCode extension for province, city, district, village, rt, rw
}

// ContactPoint - Details for all kinds of technology-mediated contact points
type ContactPoint struct {
	System string  `json:"system,omitempty"` // phone | fax | email | pager | url | sms | other
	Value  string  `json:"value,omitempty"`  // The actual contact point details
	Use    string  `json:"use,omitempty"`    // home | work | temp | old | mobile
	Rank   *int    `json:"rank,omitempty"`   // Specify preferred order of use (1 = highest)
	Period *Period `json:"period,omitempty"` // Time period when the contact point was/is in use
}

// HumanName - A name of a human with text, parts and usage information
type HumanName struct {
	Use    string   `json:"use,omitempty"`    // usual | official | temp | nickname | anonymous | old | maiden
	Text   string   `json:"text,omitempty"`   // Text representation of the full name
	Family string   `json:"family,omitempty"` // Family name (surname)
	Given  []string `json:"given,omitempty"`  // Given names (not always 'first')
	Prefix []string `json:"prefix,omitempty"` // Parts that come before the name (titles)
	Suffix []string `json:"suffix,omitempty"` // Parts that come after the name
	Period *Period  `json:"period,omitempty"` // Time period when name was/is in use
}

// Period - Time range defined by start and end date/time
type Period struct {
	Start string `json:"start,omitempty"` // Starting time with inclusive boundary (YYYY-MM-DD or YYYY-MM-DDThh:mm:ss+zz:zz)
	End   string `json:"end,omitempty"`   // End time with inclusive boundary
}

// Narrative - Human-readable summary of the resource
type Narrative struct {
	Status string `json:"status"` // generated | extensions | additional | empty
	Div    string `json:"div"`    // Limited xhtml content
}

// Meta - Metadata about a resource
type Meta struct {
	VersionId   string   `json:"versionId,omitempty"`   // Version specific identifier
	LastUpdated string   `json:"lastUpdated,omitempty"` // When the resource version last changed
	Profile     []string `json:"profile,omitempty"`     // Profiles this resource claims to conform to
	Security    []Coding `json:"security,omitempty"`    // Security Labels applied to this resource
	Tag         []Coding `json:"tag,omitempty"`         // Tags applied to this resource
}

// Quantity - A measured or measurable amount
type Quantity struct {
	Value      *float64 `json:"value,omitempty"`      // Numerical value (with implicit precision)
	Comparator string   `json:"comparator,omitempty"` // < | <= | >= | > - how to understand the value
	Unit       string   `json:"unit,omitempty"`       // Unit representation
	System     string   `json:"system,omitempty"`     // System that defines coded unit form
	Code       string   `json:"code,omitempty"`       // Coded form of the unit
}

// Duration - A length of time
type Duration Quantity

// Age - A duration of time during which an organism (or a process) has existed
type Age Quantity

// Distance - A length - a value with a unit that is a physical distance
type Distance Quantity

// Count - A measured amount (or an amount that can potentially be measured)
type Count Quantity

// MoneyQuantity - An amount of money
type MoneyQuantity struct {
	Value    *float64 `json:"value,omitempty"`    // Numerical value
	Currency string   `json:"currency,omitempty"` // ISO 4217 Currency Code
}

// SimpleQuantity - A fixed quantity (no comparator)
type SimpleQuantity Quantity

// Range - Set of values bounded by low and high
type Range struct {
	Low  *SimpleQuantity `json:"low,omitempty"`  // Low limit
	High *SimpleQuantity `json:"high,omitempty"` // High limit
}

// Ratio - A ratio of two Quantity values - a numerator and a denominator
type Ratio struct {
	Numerator   *Quantity `json:"numerator,omitempty"`   // Numerator value
	Denominator *Quantity `json:"denominator,omitempty"` // Denominator value
}

// RatioRange - A range of ratios
type RatioRange struct {
	LowNumerator  *SimpleQuantity `json:"lowNumerator,omitempty"`  // Low Numerator limit
	HighNumerator *SimpleQuantity `json:"highNumerator,omitempty"` // High Numerator limit
	Denominator   *SimpleQuantity `json:"denominator,omitempty"`   // Denominator value
}

// Annotation - Text node with attribution
type Annotation struct {
	AuthorReference *Reference `json:"authorReference,omitempty"` // Individual responsible for the annotation
	AuthorString    string     `json:"authorString,omitempty"`    // Individual responsible for the annotation (string)
	Time            string     `json:"time,omitempty"`            // When the annotation was made
	Text            string     `json:"text"`                      // The annotation - text content (REQUIRED)
}

// Attachment - Content in a format defined elsewhere
type Attachment struct {
	ContentType string `json:"contentType,omitempty"` // Mime type of the content
	Language    string `json:"language,omitempty"`    // Human language of the content (BCP-47)
	Data        string `json:"data,omitempty"`        // Data inline, base64ed
	URL         string `json:"url,omitempty"`         // Uri where the data can be found
	Size        *int   `json:"size,omitempty"`        // Number of bytes of content
	Hash        string `json:"hash,omitempty"`        // Hash of the data (sha-1, base64ed)
	Title       string `json:"title,omitempty"`       // Label to display in place of the data
	Creation    string `json:"creation,omitempty"`    // Date attachment was first created
}

// Signature - Digital Signature
type Signature struct {
	Type         []Coding    `json:"type"`                   // Indication of the reason the entity signed the object(s)
	When         string      `json:"when"`                   // When the signature was created
	Who          *Reference  `json:"who"`                    // Who signed
	OnBehalfOf   *Reference  `json:"onBehalfOf,omitempty"`   // The party represented
	TargetFormat string      `json:"targetFormat,omitempty"` // The technical format of the signed resources
	SigFormat    string      `json:"sigFormat,omitempty"`    // The technical format of the signature
	Data         string      `json:"data,omitempty"`         // The actual signature content (XML DigSig, JWS, picture, etc.)
}

// Timing - A timing schedule that specifies an event that may occur multiple times
type Timing struct {
	Event  []string      `json:"event,omitempty"`  // When the event occurs
	Repeat *TimingRepeat `json:"repeat,omitempty"` // When the event is to occur
	Code   *CodeableConcept `json:"code,omitempty"` // Code for a known / defined timing pattern
}

// TimingRepeat - When the event is to occur
type TimingRepeat struct {
	BoundsDuration *Duration `json:"boundsDuration,omitempty"` // Length/Range of lengths
	BoundsRange    *Range    `json:"boundsRange,omitempty"`    // Length/Range of lengths
	BoundsPeriod   *Period   `json:"boundsPeriod,omitempty"`   // Length/Range of lengths
	Count          *int      `json:"count,omitempty"`          // Number of times to repeat
	CountMax       *int      `json:"countMax,omitempty"`       // Maximum number of times to repeat
	Duration       *float64  `json:"duration,omitempty"`       // How long when it happens
	DurationMax    *float64  `json:"durationMax,omitempty"`    // How long when it happens (Max)
	DurationUnit   string    `json:"durationUnit,omitempty"`   // s | min | h | d | wk | mo | a - unit of time (UCUM)
	Frequency      *int      `json:"frequency,omitempty"`      // Event occurs frequency times per period
	FrequencyMax   *int      `json:"frequencyMax,omitempty"`   // Event occurs up to frequencyMax times per period
	Period         *float64  `json:"period,omitempty"`         // Event occurs frequency times per period
	PeriodMax      *float64  `json:"periodMax,omitempty"`      // Upper limit of period (3-4 hours)
	PeriodUnit     string    `json:"periodUnit,omitempty"`     // s | min | h | d | wk | mo | a - unit of time (UCUM)
	DayOfWeek      []string  `json:"dayOfWeek,omitempty"`      // mon | tue | wed | thu | fri | sat | sun
	TimeOfDay      []string  `json:"timeOfDay,omitempty"`      // Time of day for action
	When           []string  `json:"when,omitempty"`           // Code for time period of occurrence
	Offset         *int      `json:"offset,omitempty"`         // Minutes from event (before or after)
}

// SampledData - A series of measurements taken by a device
type SampledData struct {
	Origin     *SimpleQuantity `json:"origin"`               // Zero value and units
	Period     *float64        `json:"period"`               // Number of milliseconds between samples
	Factor     *float64        `json:"factor,omitempty"`     // Multiply data by this before adding to origin
	LowerLimit *float64        `json:"lowerLimit,omitempty"` // Lower limit of detection
	UpperLimit *float64        `json:"upperLimit,omitempty"` // Upper limit of detection
	Dimensions *int            `json:"dimensions"`           // Number of sample points at each time point
	Data       string          `json:"data,omitempty"`       // Decimal values with spaces, or "E" | "U" | "L"
}

// BackboneElement - Base definition for complex types defined in resources
// This is used as an embedded type for nested structures within resources
type BackboneElement struct {
	ID                string      `json:"id,omitempty"`                // Unique id for inter-element referencing
	Extension         []Extension `json:"extension,omitempty"`         // Additional content defined by implementations
	ModifierExtension []Extension `json:"modifierExtension,omitempty"` // Extensions that cannot be ignored even if unrecognized
}
