package satusehat

// Prerequisites resources for Satu Sehat FHIR
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/prerequisites/

// Organization - A grouping of people or organizations with a common purpose
type Organization struct {
	DomainResource
	Identifier []Identifier            `json:"identifier,omitempty"` // Identifies this organization across multiple systems
	Active     *bool                   `json:"active,omitempty"`     // Whether the organization's record is still in active use
	Type       []CodeableConcept       `json:"type,omitempty"`       // Kind of organization
	Name       string                  `json:"name,omitempty"`       // Name used for the organization
	Alias      []string                `json:"alias,omitempty"`      // A list of alternate names that the organization is known as
	Telecom    []ContactPoint          `json:"telecom,omitempty"`    // A contact detail for the organization
	Address    []Address               `json:"address,omitempty"`    // An address for the organization
	PartOf     *Reference              `json:"partOf,omitempty"`     // The organization of which this organization forms a part
	Contact    []OrganizationContact   `json:"contact,omitempty"`    // Contact for the organization for a certain purpose
	Endpoint   []Reference             `json:"endpoint,omitempty"`   // Technical endpoints providing access to services operated for the organization
}

// OrganizationContact - Contact for the organization for a certain purpose
type OrganizationContact struct {
	BackboneElement
	Purpose *CodeableConcept `json:"purpose,omitempty"` // The type of contact (billing, admin, hr, etc.)
	Name    *HumanName       `json:"name,omitempty"`    // A name associated with the contact
	Telecom []ContactPoint   `json:"telecom,omitempty"` // Contact details (telephone, email, etc.) for a contact
	Address *Address         `json:"address,omitempty"` // Visiting or postal addresses for the contact
}

// Location - Details and position information for a physical place
type Location struct {
	DomainResource
	Identifier              []Identifier          `json:"identifier,omitempty"`              // Unique code or number identifying the location
	Status                  string                `json:"status,omitempty"`                  // active | suspended | inactive
	OperationalStatus       *Coding               `json:"operationalStatus,omitempty"`       // The operational status (e.g. bed status)
	Name                    string                `json:"name,omitempty"`                    // Name of the location
	Alias                   []string              `json:"alias,omitempty"`                   // A list of alternate names
	Description             string                `json:"description,omitempty"`             // Additional details about the location
	Mode                    string                `json:"mode,omitempty"`                    // instance | kind
	Type                    []CodeableConcept     `json:"type,omitempty"`                    // Type of function performed
	Telecom                 []ContactPoint        `json:"telecom,omitempty"`                 // Contact details of the location
	Address                 *Address              `json:"address,omitempty"`                 // Physical location
	PhysicalType            *CodeableConcept      `json:"physicalType,omitempty"`            // Physical form of the location
	Position                *LocationPosition     `json:"position,omitempty"`                // The absolute geographic location
	ManagingOrganization    *Reference            `json:"managingOrganization,omitempty"`    // Organization responsible for provisioning and upkeep
	PartOf                  *Reference            `json:"partOf,omitempty"`                  // Another Location this one is physically a part of
	HoursOfOperation        []LocationHoursOfOperation `json:"hoursOfOperation,omitempty"` // What days/times during a week is this location usually open
	AvailabilityExceptions  string                `json:"availabilityExceptions,omitempty"`  // Description of availability exceptions
	Endpoint                []Reference           `json:"endpoint,omitempty"`                // Technical endpoints providing access to services
}

// LocationPosition - The absolute geographic location
type LocationPosition struct {
	BackboneElement
	Longitude *float64 `json:"longitude"` // Longitude with WGS84 datum (REQUIRED)
	Latitude  *float64 `json:"latitude"`  // Latitude with WGS84 datum (REQUIRED)
	Altitude  *float64 `json:"altitude,omitempty"` // Altitude with WGS84 datum
}

// LocationHoursOfOperation - What days/times during a week is this location usually open
type LocationHoursOfOperation struct {
	BackboneElement
	DaysOfWeek  []string `json:"daysOfWeek,omitempty"`  // mon | tue | wed | thu | fri | sat | sun
	AllDay      *bool    `json:"allDay,omitempty"`      // The Location is open all day
	OpeningTime string   `json:"openingTime,omitempty"` // Time that the Location opens
	ClosingTime string   `json:"closingTime,omitempty"` // Time that the Location closes
}

// Practitioner - A person with a formal responsibility in the provisioning of healthcare
type Practitioner struct {
	DomainResource
	Identifier    []Identifier                `json:"identifier,omitempty"`    // An identifier for the person as this agent
	Active        *bool                       `json:"active,omitempty"`        // Whether this practitioner's record is in active use
	Name          []HumanName                 `json:"name,omitempty"`          // The name(s) associated with the practitioner
	Telecom       []ContactPoint              `json:"telecom,omitempty"`       // A contact detail for the practitioner
	Address       []Address                   `json:"address,omitempty"`       // Address(es) of the practitioner
	Gender        string                      `json:"gender,omitempty"`        // male | female | other | unknown
	BirthDate     string                      `json:"birthDate,omitempty"`     // The date of birth for the practitioner
	Photo         []Attachment                `json:"photo,omitempty"`         // Image of the person
	Qualification []PractitionerQualification `json:"qualification,omitempty"` // Certification, licenses, or training pertaining to the provision of care
	Communication []CodeableConcept           `json:"communication,omitempty"` // A language the practitioner can use in patient communication
}

// PractitionerQualification - Certification, licenses, or training
type PractitionerQualification struct {
	BackboneElement
	Identifier []Identifier     `json:"identifier,omitempty"` // An identifier for this qualification for the practitioner
	Code       *CodeableConcept `json:"code"`                 // Coded representation of the qualification (REQUIRED)
	Period     *Period          `json:"period,omitempty"`     // Period during which the qualification is valid
	Issuer     *Reference       `json:"issuer,omitempty"`     // Organization that regulates and issues the qualification
}

// Patient - Information about an individual receiving health care services
type Patient struct {
	DomainResource
	Identifier           []Identifier           `json:"identifier,omitempty"`           // An identifier for this patient
	Active               *bool                  `json:"active,omitempty"`               // Whether this patient's record is in active use
	Name                 []HumanName            `json:"name,omitempty"`                 // A name associated with the patient (REQUIRED)
	Telecom              []ContactPoint         `json:"telecom,omitempty"`              // A contact detail for the individual
	Gender               string                 `json:"gender,omitempty"`               // male | female | other | unknown
	BirthDate            string                 `json:"birthDate,omitempty"`            // The date of birth for the individual
	DeceasedBoolean      *bool                  `json:"deceasedBoolean,omitempty"`      // Indicates if the individual is deceased or not
	DeceasedDateTime     string                 `json:"deceasedDateTime,omitempty"`     // Indicates if the individual is deceased or not
	Address              []Address              `json:"address,omitempty"`              // An address for the individual
	MaritalStatus        *CodeableConcept       `json:"maritalStatus,omitempty"`        // Marital (civil) status of a patient
	MultipleBirthBoolean *bool                  `json:"multipleBirthBoolean,omitempty"` // Whether patient is part of a multiple birth (REQUIRED)
	MultipleBirthInteger *int                   `json:"multipleBirthInteger,omitempty"` // Whether patient is part of a multiple birth
	Photo                []Attachment           `json:"photo,omitempty"`                // Image of the patient
	Contact              []PatientContact       `json:"contact,omitempty"`              // A contact party (e.g. guardian, partner, friend) for the patient
	Communication        []PatientCommunication `json:"communication,omitempty"`        // A language which may be used to communicate with the patient
	GeneralPractitioner  []Reference            `json:"generalPractitioner,omitempty"`  // Patient's nominated primary care provider
	ManagingOrganization *Reference             `json:"managingOrganization,omitempty"` // Organization that is the custodian of the patient record
	Link                 []PatientLink          `json:"link,omitempty"`                 // Link to another patient resource that concerns the same actual person
}

// PatientContact - A contact party for the patient
type PatientContact struct {
	BackboneElement
	Relationship []CodeableConcept `json:"relationship,omitempty"` // The kind of relationship
	Name         *HumanName        `json:"name,omitempty"`         // A name associated with the contact person
	Telecom      []ContactPoint    `json:"telecom,omitempty"`      // A contact detail for the person
	Address      *Address          `json:"address,omitempty"`      // Address for the contact person
	Gender       string            `json:"gender,omitempty"`       // male | female | other | unknown
	Organization *Reference        `json:"organization,omitempty"` // Organization that is associated with the contact
	Period       *Period           `json:"period,omitempty"`       // The period during which this contact person or organization is valid
}

// PatientCommunication - A language which may be used to communicate with the patient
type PatientCommunication struct {
	BackboneElement
	Language  *CodeableConcept `json:"language"`            // The language which can be used to communicate (REQUIRED)
	Preferred *bool            `json:"preferred,omitempty"` // Language preference indicator
}

// PatientLink - Link to another patient resource
type PatientLink struct {
	BackboneElement
	Other *Reference `json:"other"` // The other patient or related person resource (REQUIRED)
	Type  string     `json:"type"`  // replaced-by | replaces | refer | seealso (REQUIRED)
}
