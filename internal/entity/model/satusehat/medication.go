package satusehat

// Medication resources for prescriptions and dispensing
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/

// Medication - Definition of a medication
type Medication struct {
	DomainResource
	Identifier []Identifier          `json:"identifier,omitempty"` // Business identifier for this medication
	Code       *CodeableConcept      `json:"code,omitempty"`       // Codes that identify this medication
	Status     string                `json:"status,omitempty"`     // active | inactive | entered-in-error
	Manufacturer *Reference          `json:"manufacturer,omitempty"` // Manufacturer of the item
	Form       *CodeableConcept      `json:"form,omitempty"`       // powder | tablets | capsule +
	Amount     *Ratio                `json:"amount,omitempty"`     // Amount of drug in package
	Ingredient []MedicationIngredient `json:"ingredient,omitempty"` // Active or inactive ingredient
	Batch      *MedicationBatch      `json:"batch,omitempty"`      // Details about packaged medications
}

// MedicationIngredient - Active or inactive ingredient
type MedicationIngredient struct {
	BackboneElement
	ItemCodeableConcept *CodeableConcept `json:"itemCodeableConcept,omitempty"` // The actual ingredient or content
	ItemReference       *Reference       `json:"itemReference,omitempty"`       // The actual ingredient or content
	IsActive            *bool            `json:"isActive,omitempty"`            // Active ingredient indicator
	Strength            *Ratio           `json:"strength,omitempty"`            // Quantity of ingredient present
}

// MedicationBatch - Details about packaged medications
type MedicationBatch struct {
	BackboneElement
	LotNumber      string `json:"lotNumber,omitempty"`      // Identifier assigned to batch
	ExpirationDate string `json:"expirationDate,omitempty"` // When batch will expire
}

// MedicationRequest - Ordering of medication for patient
type MedicationRequest struct {
	DomainResource
	Identifier                []Identifier                       `json:"identifier,omitempty"`                // External ids for this request
	Status                    string                             `json:"status"`                              // active | on-hold | cancelled | completed | entered-in-error | stopped | draft | unknown (REQUIRED)
	StatusReason              *CodeableConcept                   `json:"statusReason,omitempty"`              // Reason for current status
	Intent                    string                             `json:"intent"`                              // proposal | plan | order | original-order | reflex-order | filler-order | instance-order | option (REQUIRED)
	Category                  []CodeableConcept                  `json:"category,omitempty"`                  // Type of medication usage
	Priority                  string                             `json:"priority,omitempty"`                  // routine | urgent | asap | stat
	DoNotPerform              *bool                              `json:"doNotPerform,omitempty"`              // True if medication was not administered
	ReportedBoolean           *bool                              `json:"reportedBoolean,omitempty"`           // Reported rather than primary record
	ReportedReference         *Reference                         `json:"reportedReference,omitempty"`         // Reported rather than primary record
	MedicationCodeableConcept *CodeableConcept                   `json:"medicationCodeableConcept,omitempty"` // Medication to be taken
	MedicationReference       *Reference                         `json:"medicationReference,omitempty"`       // Medication to be taken
	Subject                   *Reference                         `json:"subject"`                             // Who or group medication request is for (REQUIRED)
	Encounter                 *Reference                         `json:"encounter,omitempty"`                 // Encounter created as part of encounter/admission/stay
	SupportingInformation     []Reference                        `json:"supportingInformation,omitempty"`     // Information to support ordering of the medication
	AuthoredOn                string                             `json:"authoredOn,omitempty"`                // When request was initially authored
	Requester                 *Reference                         `json:"requester,omitempty"`                 // Who/What requested the Request
	Performer                 *Reference                         `json:"performer,omitempty"`                 // Intended performer of administration
	PerformerType             *CodeableConcept                   `json:"performerType,omitempty"`             // Desired kind of performer of the medication administration
	Recorder                  *Reference                         `json:"recorder,omitempty"`                  // Person who entered the request
	ReasonCode                []CodeableConcept                  `json:"reasonCode,omitempty"`                // Reason or indication for ordering or not ordering the medication
	ReasonReference           []Reference                        `json:"reasonReference,omitempty"`           // Condition or observation that supports why the prescription is being written
	InstantiatesCanonical     []string                           `json:"instantiatesCanonical,omitempty"`     // Instantiates FHIR protocol or definition
	InstantiatesURI           []string                           `json:"instantiatesUri,omitempty"`           // Instantiates external protocol or definition
	BasedOn                   []Reference                        `json:"basedOn,omitempty"`                   // What request fulfills
	GroupIdentifier           *Identifier                        `json:"groupIdentifier,omitempty"`           // Composite request this is part of
	CourseOfTherapyType       *CodeableConcept                   `json:"courseOfTherapyType,omitempty"`       // Overall pattern of medication administration
	Insurance                 []Reference                        `json:"insurance,omitempty"`                 // Associated insurance coverage
	Note                      []Annotation                       `json:"note,omitempty"`                      // Information about the prescription
	DosageInstruction         []Dosage                           `json:"dosageInstruction,omitempty"`         // How the medication should be taken
	DispenseRequest           *MedicationRequestDispenseRequest  `json:"dispenseRequest,omitempty"`           // Medication supply authorization
	Substitution              *MedicationRequestSubstitution     `json:"substitution,omitempty"`              // Any restrictions on medication substitution
	PriorPrescription         *Reference                         `json:"priorPrescription,omitempty"`         // An order/prescription that is being replaced
	DetectedIssue             []Reference                        `json:"detectedIssue,omitempty"`             // Clinical Issue with action
	EventHistory              []Reference                        `json:"eventHistory,omitempty"`              // A list of events of interest in the lifecycle
}

// MedicationRequestDispenseRequest - Medication supply authorization
type MedicationRequestDispenseRequest struct {
	BackboneElement
	InitialFill          *MedicationRequestDispenseRequestInitialFill `json:"initialFill,omitempty"`          // First fill details
	DispenseInterval     *Duration                                    `json:"dispenseInterval,omitempty"`     // Minimum period of time between dispenses
	ValidityPeriod       *Period                                      `json:"validityPeriod,omitempty"`       // Time period supply is authorized for
	NumberOfRepeatsAllowed *int                                       `json:"numberOfRepeatsAllowed,omitempty"` // Number of refills authorized
	Quantity             *Quantity                                    `json:"quantity,omitempty"`             // Amount of medication to supply per dispense
	ExpectedSupplyDuration *Duration                                  `json:"expectedSupplyDuration,omitempty"` // Number of days supply per dispense
	Performer            *Reference                                   `json:"performer,omitempty"`            // Intended dispenser
}

// MedicationRequestDispenseRequestInitialFill - First fill details
type MedicationRequestDispenseRequestInitialFill struct {
	BackboneElement
	Quantity *Quantity `json:"quantity,omitempty"` // First fill quantity
	Duration *Duration `json:"duration,omitempty"` // First fill duration
}

// MedicationRequestSubstitution - Any restrictions on medication substitution
type MedicationRequestSubstitution struct {
	BackboneElement
	AllowedBoolean      *bool            `json:"allowedBoolean,omitempty"`      // Whether substitution is allowed or not
	AllowedCodeableConcept *CodeableConcept `json:"allowedCodeableConcept,omitempty"` // Whether substitution is allowed or not
	Reason              *CodeableConcept `json:"reason,omitempty"`              // Why should (not) substitution be made
}

// Dosage - How medication should be taken
type Dosage struct {
	BackboneElement
	Sequence               *int              `json:"sequence,omitempty"`               // The order of the dosage instructions
	Text                   string            `json:"text,omitempty"`                   // Free text dosage instructions
	AdditionalInstruction  []CodeableConcept `json:"additionalInstruction,omitempty"`  // Supplemental instruction
	PatientInstruction     string            `json:"patientInstruction,omitempty"`     // Patient or consumer oriented instructions
	Timing                 *Timing           `json:"timing,omitempty"`                 // When medication should be administered
	AsNeededBoolean        *bool             `json:"asNeededBoolean,omitempty"`        // Take "as needed"
	AsNeededCodeableConcept *CodeableConcept `json:"asNeededCodeableConcept,omitempty"` // Take "as needed"
	Site                   *CodeableConcept  `json:"site,omitempty"`                   // Body site to administer to
	Route                  *CodeableConcept  `json:"route,omitempty"`                  // How drug should enter body
	Method                 *CodeableConcept  `json:"method,omitempty"`                 // Technique for administering medication
	DoseAndRate            []DosageDoseAndRate `json:"doseAndRate,omitempty"`          // Amount of medication administered
	MaxDosePerPeriod       *Ratio            `json:"maxDosePerPeriod,omitempty"`       // Upper limit on medication per unit of time
	MaxDosePerAdministration *Quantity       `json:"maxDosePerAdministration,omitempty"` // Upper limit on medication per administration
	MaxDosePerLifetime     *Quantity         `json:"maxDosePerLifetime,omitempty"`     // Upper limit on medication per lifetime of the patient
}

// DosageDoseAndRate - Amount of medication administered
type DosageDoseAndRate struct {
	BackboneElement
	Type          *CodeableConcept `json:"type,omitempty"`          // The kind of dose or rate specified
	DoseRange     *Range           `json:"doseRange,omitempty"`     // Amount of medication per dose
	DoseQuantity  *Quantity        `json:"doseQuantity,omitempty"`  // Amount of medication per dose
	RateRatio     *Ratio           `json:"rateRatio,omitempty"`     // Amount of medication per unit of time
	RateRange     *Range           `json:"rateRange,omitempty"`     // Amount of medication per unit of time
	RateQuantity  *Quantity        `json:"rateQuantity,omitempty"`  // Amount of medication per unit of time
}

// MedicationDispense - Dispensing a medication to a named patient
type MedicationDispense struct {
	DomainResource
	Identifier                []Identifier                        `json:"identifier,omitempty"`                // External identifier
	PartOf                    []Reference                         `json:"partOf,omitempty"`                    // Event that dispense is part of
	Status                    string                              `json:"status"`                              // preparation | in-progress | cancelled | on-hold | completed | entered-in-error | stopped | declined | unknown (REQUIRED)
	StatusReasonCodeableConcept *CodeableConcept                  `json:"statusReasonCodeableConcept,omitempty"` // Why a dispense was not performed
	StatusReasonReference     *Reference                          `json:"statusReasonReference,omitempty"`     // Why a dispense was not performed
	Category                  *CodeableConcept                    `json:"category,omitempty"`                  // Type of medication dispense
	MedicationCodeableConcept *CodeableConcept                    `json:"medicationCodeableConcept,omitempty"` // What medication was supplied
	MedicationReference       *Reference                          `json:"medicationReference,omitempty"`       // What medication was supplied
	Subject                   *Reference                          `json:"subject,omitempty"`                   // Who the dispense is for
	Context                   *Reference                          `json:"context,omitempty"`                   // Encounter / Episode associated with event
	SupportingInformation     []Reference                         `json:"supportingInformation,omitempty"`     // Information that supports the dispensing of the medication
	Performer                 []MedicationDispensePerformer       `json:"performer,omitempty"`                 // Who performed event
	Location                  *Reference                          `json:"location,omitempty"`                  // Where the dispense occurred
	AuthorizingPrescription   []Reference                         `json:"authorizingPrescription,omitempty"`   // Medication order that authorizes the dispense
	Type                      *CodeableConcept                    `json:"type,omitempty"`                      // Trial fill, partial fill, emergency fill, etc.
	Quantity                  *Quantity                           `json:"quantity,omitempty"`                  // Amount dispensed
	DaysSupply                *Quantity                           `json:"daysSupply,omitempty"`                // Amount of medication expressed as a timing amount
	WhenPrepared              string                              `json:"whenPrepared,omitempty"`              // When product was packaged and reviewed
	WhenHandedOver            string                              `json:"whenHandedOver,omitempty"`            // When product was given out
	Destination               *Reference                          `json:"destination,omitempty"`               // Where the medication was sent
	Receiver                  []Reference                         `json:"receiver,omitempty"`                  // Who collected the medication
	Note                      []Annotation                        `json:"note,omitempty"`                      // Information about the dispense
	DosageInstruction         []Dosage                            `json:"dosageInstruction,omitempty"`         // How the medication is to be used by the patient
	Substitution              *MedicationDispenseSubstitution     `json:"substitution,omitempty"`              // Whether a substitution was performed on the dispense
	DetectedIssue             []Reference                         `json:"detectedIssue,omitempty"`             // Clinical issue with action
	EventHistory              []Reference                         `json:"eventHistory,omitempty"`              // A list of relevant lifecycle events
}

// MedicationDispensePerformer - Who performed event
type MedicationDispensePerformer struct {
	BackboneElement
	Function *CodeableConcept `json:"function,omitempty"` // Who performed the dispense and what they did
	Actor    *Reference       `json:"actor"`              // Individual who was performing (REQUIRED)
}

// MedicationDispenseSubstitution - Whether a substitution was performed on the dispense
type MedicationDispenseSubstitution struct {
	BackboneElement
	WasSubstituted bool              `json:"wasSubstituted"`       // Whether a substitution was or was not performed on the dispense (REQUIRED)
	Type           *CodeableConcept  `json:"type,omitempty"`       // Code signifying whether a different drug was dispensed from what was prescribed
	Reason         []CodeableConcept `json:"reason,omitempty"`     // Why was substitution made
	ResponsibleParty []Reference     `json:"responsibleParty,omitempty"` // Who is responsible for the substitution
}

// MedicationAdministration - Administration of medication to a patient
type MedicationAdministration struct {
	DomainResource
	Identifier                []Identifier                              `json:"identifier,omitempty"`                // External identifier
	InstantiatesURI           []string                                  `json:"instantiatesUri,omitempty"`           // Instantiates external protocol or definition
	PartOf                    []Reference                               `json:"partOf,omitempty"`                    // Part of referenced event
	Status                    string                                    `json:"status"`                              // in-progress | not-done | on-hold | completed | entered-in-error | stopped | unknown (REQUIRED)
	StatusReason              []CodeableConcept                         `json:"statusReason,omitempty"`              // Reason administration not performed
	Category                  *CodeableConcept                          `json:"category,omitempty"`                  // Type of medication usage
	MedicationCodeableConcept *CodeableConcept                          `json:"medicationCodeableConcept,omitempty"` // What was administered
	MedicationReference       *Reference                                `json:"medicationReference,omitempty"`       // What was administered
	Subject                   *Reference                                `json:"subject"`                             // Who received medication (REQUIRED)
	Context                   *Reference                                `json:"context,omitempty"`                   // Encounter or Episode of Care administered as part of
	SupportingInformation     []Reference                               `json:"supportingInformation,omitempty"`     // Additional information to support administration
	EffectiveDateTime         string                                    `json:"effectiveDateTime,omitempty"`         // Start and end time of administration
	EffectivePeriod           *Period                                   `json:"effectivePeriod,omitempty"`           // Start and end time of administration
	Performer                 []MedicationAdministrationPerformer       `json:"performer,omitempty"`                 // Who performed the medication administration
	ReasonCode                []CodeableConcept                         `json:"reasonCode,omitempty"`                // Reason administration performed
	ReasonReference           []Reference                               `json:"reasonReference,omitempty"`           // Condition or observation that supports why the medication was administered
	Request                   *Reference                                `json:"request,omitempty"`                   // Request administration performed against
	Device                    []Reference                               `json:"device,omitempty"`                    // Device used to administer
	Note                      []Annotation                              `json:"note,omitempty"`                      // Information about the administration
	Dosage                    *MedicationAdministrationDosage           `json:"dosage,omitempty"`                    // Details of how medication was taken
	EventHistory              []Reference                               `json:"eventHistory,omitempty"`              // A list of events of interest in the lifecycle
}

// MedicationAdministrationPerformer - Who performed the medication administration
type MedicationAdministrationPerformer struct {
	BackboneElement
	Function *CodeableConcept `json:"function,omitempty"` // Type of performance
	Actor    *Reference       `json:"actor"`              // Who performed the medication administration (REQUIRED)
}

// MedicationAdministrationDosage - Details of how medication was taken
type MedicationAdministrationDosage struct {
	BackboneElement
	Text      string           `json:"text,omitempty"`      // Free text dosage instructions
	Site      *CodeableConcept `json:"site,omitempty"`      // Body site administered to
	Route     *CodeableConcept `json:"route,omitempty"`     // Path of substance into body
	Method    *CodeableConcept `json:"method,omitempty"`    // How drug was administered
	Dose      *Quantity        `json:"dose,omitempty"`      // Amount of medication per dose
	RateRatio *Ratio           `json:"rateRatio,omitempty"` // Dose quantity per unit of time
	RateQuantity *Quantity     `json:"rateQuantity,omitempty"` // Dose quantity per unit of time
}

// MedicationStatement - Record of medication being taken by a patient
type MedicationStatement struct {
	DomainResource
	Identifier                []Identifier      `json:"identifier,omitempty"`                // External identifier
	BasedOn                   []Reference       `json:"basedOn,omitempty"`                   // Fulfils plan, proposal or order
	PartOf                    []Reference       `json:"partOf,omitempty"`                    // Part of referenced event
	Status                    string            `json:"status"`                              // active | completed | entered-in-error | intended | stopped | on-hold | unknown | not-taken (REQUIRED)
	StatusReason              []CodeableConcept `json:"statusReason,omitempty"`              // Reason for current status
	Category                  *CodeableConcept  `json:"category,omitempty"`                  // Type of medication usage
	MedicationCodeableConcept *CodeableConcept  `json:"medicationCodeableConcept,omitempty"` // What medication was taken
	MedicationReference       *Reference        `json:"medicationReference,omitempty"`       // What medication was taken
	Subject                   *Reference        `json:"subject"`                             // Who is/was taking the medication (REQUIRED)
	Context                   *Reference        `json:"context,omitempty"`                   // Encounter / Episode associated with MedicationStatement
	EffectiveDateTime         string            `json:"effectiveDateTime,omitempty"`         // The date/time or interval when the medication is/was/will be taken
	EffectivePeriod           *Period           `json:"effectivePeriod,omitempty"`           // The date/time or interval when the medication is/was/will be taken
	DateAsserted              string            `json:"dateAsserted,omitempty"`              // When the statement was asserted
	InformationSource         *Reference        `json:"informationSource,omitempty"`         // Person or organization that provided the information
	DerivedFrom               []Reference       `json:"derivedFrom,omitempty"`               // Additional supporting information
	ReasonCode                []CodeableConcept `json:"reasonCode,omitempty"`                // Reason for why the medication is being/was taken
	ReasonReference           []Reference       `json:"reasonReference,omitempty"`           // Condition or observation that supports why the medication is being/was taken
	Note                      []Annotation      `json:"note,omitempty"`                      // Further information about the statement
	Dosage                    []Dosage          `json:"dosage,omitempty"`                    // Details of how medication is/was taken or should be taken
}
