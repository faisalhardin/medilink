package satusehat

// Additional interoperability resources
// Based on https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/

// ServiceRequest - A request for a service to be performed
type ServiceRequest struct {
	DomainResource
	Identifier            []Identifier      `json:"identifier,omitempty"`            // Identifiers assigned to this order
	InstantiatesCanonical []string          `json:"instantiatesCanonical,omitempty"` // Instantiates FHIR protocol or definition
	InstantiatesURI       []string          `json:"instantiatesUri,omitempty"`       // Instantiates external protocol or definition
	BasedOn               []Reference       `json:"basedOn,omitempty"`               // What request fulfills
	Replaces              []Reference       `json:"replaces,omitempty"`              // What request replaces
	Requisition           *Identifier       `json:"requisition,omitempty"`           // Composite Request ID
	Status                string            `json:"status"`                          // draft | active | on-hold | revoked | completed | entered-in-error | unknown (REQUIRED)
	Intent                string            `json:"intent"`                          // proposal | plan | directive | order | original-order | reflex-order | filler-order | instance-order | option (REQUIRED)
	Category              []CodeableConcept `json:"category,omitempty"`              // Classification of service
	Priority              string            `json:"priority,omitempty"`              // routine | urgent | asap | stat
	DoNotPerform          *bool             `json:"doNotPerform,omitempty"`          // True if service/procedure should not be performed
	Code                  *CodeableConcept  `json:"code,omitempty"`                  // What is being requested/ordered
	OrderDetail           []CodeableConcept `json:"orderDetail,omitempty"`           // Additional order information
	QuantityQuantity      *Quantity         `json:"quantityQuantity,omitempty"`      // Service amount
	QuantityRatio         *Ratio            `json:"quantityRatio,omitempty"`         // Service amount
	QuantityRange         *Range            `json:"quantityRange,omitempty"`         // Service amount
	Subject               *Reference        `json:"subject"`                         // Individual or Entity the service is ordered for (REQUIRED)
	Encounter             *Reference        `json:"encounter,omitempty"`             // Encounter in which the request was created
	OccurrenceDateTime    string            `json:"occurrenceDateTime,omitempty"`    // When service should occur
	OccurrencePeriod      *Period           `json:"occurrencePeriod,omitempty"`      // When service should occur
	OccurrenceTiming      *Timing           `json:"occurrenceTiming,omitempty"`      // When service should occur
	AsNeededBoolean       *bool             `json:"asNeededBoolean,omitempty"`       // Preconditions for service
	AsNeededCodeableConcept *CodeableConcept `json:"asNeededCodeableConcept,omitempty"` // Preconditions for service
	AuthoredOn            string            `json:"authoredOn,omitempty"`            // Date request signed
	Requester             *Reference        `json:"requester,omitempty"`             // Who/what is requesting service
	PerformerType         *CodeableConcept  `json:"performerType,omitempty"`         // Performer role
	Performer             []Reference       `json:"performer,omitempty"`             // Requested performer
	LocationCode          []CodeableConcept `json:"locationCode,omitempty"`          // Requested location
	LocationReference     []Reference       `json:"locationReference,omitempty"`     // Requested location
	ReasonCode            []CodeableConcept `json:"reasonCode,omitempty"`            // Explanation/Justification for procedure or service
	ReasonReference       []Reference       `json:"reasonReference,omitempty"`       // Explanation/Justification for service or service
	Insurance             []Reference       `json:"insurance,omitempty"`             // Associated insurance coverage
	SupportingInfo        []Reference       `json:"supportingInfo,omitempty"`        // Additional clinical information
	Specimen              []Reference       `json:"specimen,omitempty"`              // Procedure Samples
	BodySite              []CodeableConcept `json:"bodySite,omitempty"`              // Location on Body
	Note                  []Annotation      `json:"note,omitempty"`                  // Comments
	PatientInstruction    string            `json:"patientInstruction,omitempty"`    // Patient or consumer-oriented instructions
	RelevantHistory       []Reference       `json:"relevantHistory,omitempty"`       // Request provenance
}

// Specimen - Sample for analysis
type Specimen struct {
	DomainResource
	Identifier       []Identifier          `json:"identifier,omitempty"`       // External Identifier
	AccessionIdentifier *Identifier        `json:"accessionIdentifier,omitempty"` // Identifier assigned by the lab
	Status           string                `json:"status,omitempty"`           // available | unavailable | unsatisfactory | entered-in-error
	Type             *CodeableConcept      `json:"type,omitempty"`             // Kind of material that forms the specimen
	Subject          *Reference            `json:"subject,omitempty"`          // Where the specimen came from
	ReceivedTime     string                `json:"receivedTime,omitempty"`     // The time when specimen was received
	Parent           []Reference           `json:"parent,omitempty"`           // Specimen from which this specimen originated
	Request          []Reference           `json:"request,omitempty"`          // Why the specimen was collected
	Collection       *SpecimenCollection   `json:"collection,omitempty"`       // Collection details
	Processing       []SpecimenProcessing  `json:"processing,omitempty"`       // Processing and processing step details
	Container        []SpecimenContainer   `json:"container,omitempty"`        // Direct container of specimen
	Condition        []CodeableConcept     `json:"condition,omitempty"`        // State of the specimen
	Note             []Annotation          `json:"note,omitempty"`             // Comments
}

// SpecimenCollection - Collection details
type SpecimenCollection struct {
	BackboneElement
	Collector         *Reference       `json:"collector,omitempty"`         // Who collected the specimen
	CollectedDateTime string           `json:"collectedDateTime,omitempty"` // Collection time
	CollectedPeriod   *Period          `json:"collectedPeriod,omitempty"`   // Collection time
	Duration          *Duration        `json:"duration,omitempty"`          // How long it took to collect specimen
	Quantity          *Quantity        `json:"quantity,omitempty"`          // The quantity of specimen collected
	Method            *CodeableConcept `json:"method,omitempty"`            // Technique used to perform collection
	BodySite          *CodeableConcept `json:"bodySite,omitempty"`          // Anatomical collection site
	FastingStatusCodeableConcept *CodeableConcept `json:"fastingStatusCodeableConcept,omitempty"` // Whether or how long patient abstained from food and/or drink
	FastingStatusDuration *Duration    `json:"fastingStatusDuration,omitempty"` // Whether or how long patient abstained from food and/or drink
}

// SpecimenProcessing - Processing and processing step details
type SpecimenProcessing struct {
	BackboneElement
	Description   string           `json:"description,omitempty"`   // Textual description of procedure
	Procedure     *CodeableConcept `json:"procedure,omitempty"`     // Indicates the treatment step applied to the specimen
	Additive      []Reference      `json:"additive,omitempty"`      // Material used in the processing step
	TimeDateTime  string           `json:"timeDateTime,omitempty"`  // Date and time of specimen processing
	TimePeriod    *Period          `json:"timePeriod,omitempty"`    // Date and time of specimen processing
}

// SpecimenContainer - Direct container of specimen
type SpecimenContainer struct {
	BackboneElement
	Identifier              []Identifier     `json:"identifier,omitempty"`              // Id for the container
	Description             string           `json:"description,omitempty"`             // Textual description of the container
	Type                    *CodeableConcept `json:"type,omitempty"`                    // Kind of container directly associated with specimen
	Capacity                *Quantity        `json:"capacity,omitempty"`                // Container volume or size
	SpecimenQuantity        *Quantity        `json:"specimenQuantity,omitempty"`        // Quantity of specimen within container
	AdditiveCodeableConcept *CodeableConcept `json:"additiveCodeableConcept,omitempty"` // Additive associated with container
	AdditiveReference       *Reference       `json:"additiveReference,omitempty"`       // Additive associated with container
}

// DiagnosticReport - A diagnostic report
type DiagnosticReport struct {
	DomainResource
	Identifier            []Identifier      `json:"identifier,omitempty"`            // Business identifier for report
	BasedOn               []Reference       `json:"basedOn,omitempty"`               // What was requested
	Status                string            `json:"status"`                          // registered | partial | preliminary | final (REQUIRED)
	Category              []CodeableConcept `json:"category,omitempty"`              // Service category
	Code                  *CodeableConcept  `json:"code"`                            // Name/Code for this diagnostic report (REQUIRED)
	Subject               *Reference        `json:"subject,omitempty"`               // The subject of the report
	Encounter             *Reference        `json:"encounter,omitempty"`             // Health care event when test ordered
	EffectiveDateTime     string            `json:"effectiveDateTime,omitempty"`     // Clinically relevant time/time-period for report
	EffectivePeriod       *Period           `json:"effectivePeriod,omitempty"`       // Clinically relevant time/time-period for report
	Issued                string            `json:"issued,omitempty"`                // DateTime this version was made
	Performer             []Reference       `json:"performer,omitempty"`             // Responsible Diagnostic Service
	ResultsInterpreter    []Reference       `json:"resultsInterpreter,omitempty"`    // Primary result interpreter
	Specimen              []Reference       `json:"specimen,omitempty"`              // Specimens this report is based on
	Result                []Reference       `json:"result,omitempty"`                // Observations
	ImagingStudy          []Reference       `json:"imagingStudy,omitempty"`          // Reference to full details of imaging associated with the diagnostic report
	Media                 []DiagnosticReportMedia `json:"media,omitempty"`       // Key images associated with this report
	Conclusion            string            `json:"conclusion,omitempty"`            // Clinical conclusion (interpretation) of test results
	ConclusionCode        []CodeableConcept `json:"conclusionCode,omitempty"`        // Codes for the clinical conclusion of test results
	PresentedForm         []Attachment      `json:"presentedForm,omitempty"`         // Entire report as issued
}

// DiagnosticReportMedia - Key images associated with this report
type DiagnosticReportMedia struct {
	BackboneElement
	Comment string     `json:"comment,omitempty"` // Comment about the image
	Link    *Reference `json:"link"`              // Reference to the image source (REQUIRED)
}

// AllergyIntolerance - Allergy or Intolerance
type AllergyIntolerance struct {
	DomainResource
	Identifier           []Identifier                  `json:"identifier,omitempty"`           // External ids for this item
	ClinicalStatus       *CodeableConcept              `json:"clinicalStatus,omitempty"`       // active | inactive | resolved
	VerificationStatus   *CodeableConcept              `json:"verificationStatus,omitempty"`   // unconfirmed | confirmed | refuted | entered-in-error
	Type                 string                        `json:"type,omitempty"`                 // allergy | intolerance
	Category             []string                      `json:"category,omitempty"`             // food | medication | environment | biologic
	Criticality          string                        `json:"criticality,omitempty"`          // low | high | unable-to-assess
	Code                 *CodeableConcept              `json:"code,omitempty"`                 // Code that identifies the allergy or intolerance
	Patient              *Reference                    `json:"patient"`                        // Who the sensitivity is for (REQUIRED)
	Encounter            *Reference                    `json:"encounter,omitempty"`            // Encounter when the allergy or intolerance was asserted
	OnsetDateTime        string                        `json:"onsetDateTime,omitempty"`        // When allergy or intolerance was identified
	OnsetAge             *Age                          `json:"onsetAge,omitempty"`             // When allergy or intolerance was identified
	OnsetPeriod          *Period                       `json:"onsetPeriod,omitempty"`          // When allergy or intolerance was identified
	OnsetRange           *Range                        `json:"onsetRange,omitempty"`           // When allergy or intolerance was identified
	OnsetString          string                        `json:"onsetString,omitempty"`          // When allergy or intolerance was identified
	RecordedDate         string                        `json:"recordedDate,omitempty"`         // Date first version of the resource instance was recorded
	Recorder             *Reference                    `json:"recorder,omitempty"`             // Who recorded the sensitivity
	Asserter             *Reference                    `json:"asserter,omitempty"`             // Source of the information about the allergy
	LastOccurrence       string                        `json:"lastOccurrence,omitempty"`       // Date(/time) of last known occurrence of a reaction
	Note                 []Annotation                  `json:"note,omitempty"`                 // Additional text not captured in other fields
	Reaction             []AllergyIntoleranceReaction  `json:"reaction,omitempty"`             // Adverse Reaction Events linked to exposure to substance
}

// AllergyIntoleranceReaction - Adverse Reaction Events
type AllergyIntoleranceReaction struct {
	BackboneElement
	Substance     *CodeableConcept  `json:"substance,omitempty"`     // Specific substance or pharmaceutical product considered to be responsible
	Manifestation []CodeableConcept `json:"manifestation"`           // Clinical symptoms/signs associated with the Event (REQUIRED)
	Description   string            `json:"description,omitempty"`   // Description of the event as a whole
	Onset         string            `json:"onset,omitempty"`         // Date(/time) when manifestations showed
	Severity      string            `json:"severity,omitempty"`      // mild | moderate | severe
	ExposureRoute *CodeableConcept  `json:"exposureRoute,omitempty"` // How the subject was exposed to the substance
	Note          []Annotation      `json:"note,omitempty"`          // Text about event not captured in other fields
}

// ClinicalImpression - A clinical assessment
type ClinicalImpression struct {
	DomainResource
	Identifier          []Identifier                  `json:"identifier,omitempty"`          // Business identifier
	Status              string                        `json:"status"`                        // in-progress | completed | entered-in-error (REQUIRED)
	StatusReason        *CodeableConcept              `json:"statusReason,omitempty"`        // Reason for current status
	Code                *CodeableConcept              `json:"code,omitempty"`                // Kind of assessment performed
	Description         string                        `json:"description,omitempty"`         // Why/how the assessment was performed
	Subject             *Reference                    `json:"subject"`                       // Patient or group assessed (REQUIRED)
	Encounter           *Reference                    `json:"encounter,omitempty"`           // Encounter created as part of
	EffectiveDateTime   string                        `json:"effectiveDateTime,omitempty"`   // Time of assessment
	EffectivePeriod     *Period                       `json:"effectivePeriod,omitempty"`     // Time of assessment
	Date                string                        `json:"date,omitempty"`                // When the assessment was documented
	Assessor            *Reference                    `json:"assessor,omitempty"`            // The clinician performing the assessment
	Previous            *Reference                    `json:"previous,omitempty"`            // Reference to last assessment
	Problem             []Reference                   `json:"problem,omitempty"`             // Relevant impressions of patient state
	Investigation       []ClinicalImpressionInvestigation `json:"investigation,omitempty"`   // One or more sets of investigations (signs, symptoms, etc.)
	Protocol            []string                      `json:"protocol,omitempty"`            // Clinical Protocol followed
	Summary             string                        `json:"summary,omitempty"`             // Summary of the assessment
	Finding             []ClinicalImpressionFinding   `json:"finding,omitempty"`             // Possible or likely findings and diagnoses
	PrognosisCodeableConcept []CodeableConcept        `json:"prognosisCodeableConcept,omitempty"` // Estimate of likely outcome
	PrognosisReference  []Reference                   `json:"prognosisReference,omitempty"`  // RiskAssessment expressing likely outcome
	SupportingInfo      []Reference                   `json:"supportingInfo,omitempty"`      // Information supporting the clinical impression
	Note                []Annotation                  `json:"note,omitempty"`                // Comments made about the ClinicalImpression
}

// ClinicalImpressionInvestigation - One or more sets of investigations
type ClinicalImpressionInvestigation struct {
	BackboneElement
	Code *CodeableConcept `json:"code"`          // A name/code for the set (REQUIRED)
	Item []Reference      `json:"item,omitempty"` // Record of a specific investigation
}

// ClinicalImpressionFinding - Possible or likely findings and diagnoses
type ClinicalImpressionFinding struct {
	BackboneElement
	ItemCodeableConcept *CodeableConcept `json:"itemCodeableConcept,omitempty"` // What was found
	ItemReference       *Reference       `json:"itemReference,omitempty"`       // What was found
	Basis               string           `json:"basis,omitempty"`               // Which investigations support finding
}

// Immunization - Immunization event information
type Immunization struct {
	DomainResource
	Identifier              []Identifier            `json:"identifier,omitempty"`              // Business identifier
	Status                  string                  `json:"status"`                            // completed | entered-in-error | not-done (REQUIRED)
	StatusReason            *CodeableConcept        `json:"statusReason,omitempty"`            // Reason not done
	VaccineCode             *CodeableConcept        `json:"vaccineCode"`                       // Vaccine product administered (REQUIRED)
	Patient                 *Reference              `json:"patient"`                           // Who was immunized (REQUIRED)
	Encounter               *Reference              `json:"encounter,omitempty"`               // Encounter immunization was part of
	OccurrenceDateTime      string                  `json:"occurrenceDateTime,omitempty"`      // Vaccine administration date
	OccurrenceString        string                  `json:"occurrenceString,omitempty"`        // Vaccine administration date
	Recorded                string                  `json:"recorded,omitempty"`                // When the immunization was first captured
	PrimarySource           *bool                   `json:"primarySource,omitempty"`           // Indicates context the data was recorded in
	ReportOrigin            *CodeableConcept        `json:"reportOrigin,omitempty"`            // Indicates the source of a reported record
	Location                *Reference              `json:"location,omitempty"`                // Where immunization occurred
	Manufacturer            *Reference              `json:"manufacturer,omitempty"`            // Vaccine manufacturer
	LotNumber               string                  `json:"lotNumber,omitempty"`               // Vaccine lot number
	ExpirationDate          string                  `json:"expirationDate,omitempty"`          // Vaccine expiration date
	Site                    *CodeableConcept        `json:"site,omitempty"`                    // Body site vaccine was administered
	Route                   *CodeableConcept        `json:"route,omitempty"`                   // How vaccine entered body
	DoseQuantity            *Quantity               `json:"doseQuantity,omitempty"`            // Amount of vaccine administered
	Performer               []ImmunizationPerformer `json:"performer,omitempty"`               // Who performed event
	Note                    []Annotation            `json:"note,omitempty"`                    // Additional immunization notes
	ReasonCode              []CodeableConcept       `json:"reasonCode,omitempty"`              // Why immunization occurred
	ReasonReference         []Reference             `json:"reasonReference,omitempty"`         // Why immunization occurred
	IsSubpotent             *bool                   `json:"isSubpotent,omitempty"`             // Dose potency
	SubpotentReason         []CodeableConcept       `json:"subpotentReason,omitempty"`         // Reason for being subpotent
	Education               []ImmunizationEducation `json:"education,omitempty"`               // Educational material presented to patient
	ProgramEligibility      []CodeableConcept       `json:"programEligibility,omitempty"`      // Patient eligibility for a vaccination program
	FundingSource           *CodeableConcept        `json:"fundingSource,omitempty"`           // Funding source for the vaccine
	Reaction                []ImmunizationReaction  `json:"reaction,omitempty"`                // Details of a reaction that follows immunization
	ProtocolApplied         []ImmunizationProtocolApplied `json:"protocolApplied,omitempty"` // Protocol followed by the provider
}

// ImmunizationPerformer - Who performed event
type ImmunizationPerformer struct {
	BackboneElement
	Function *CodeableConcept `json:"function,omitempty"` // What type of performance was done
	Actor    *Reference       `json:"actor"`              // Individual or organization who was performing (REQUIRED)
}

// ImmunizationEducation - Educational material presented to patient
type ImmunizationEducation struct {
	BackboneElement
	DocumentType         string `json:"documentType,omitempty"`         // Educational material document identifier
	Reference            string `json:"reference,omitempty"`            // Educational material reference pointer
	PublicationDate      string `json:"publicationDate,omitempty"`      // Educational material publication date
	PresentationDate     string `json:"presentationDate,omitempty"`     // Educational material presentation date
}

// ImmunizationReaction - Details of a reaction that follows immunization
type ImmunizationReaction struct {
	BackboneElement
	Date   string     `json:"date,omitempty"`   // When reaction started
	Detail *Reference `json:"detail,omitempty"` // Additional information on reaction
	Reported *bool    `json:"reported,omitempty"` // Indicates self-reported reaction
}

// ImmunizationProtocolApplied - Protocol followed by the provider
type ImmunizationProtocolApplied struct {
	BackboneElement
	Series           string            `json:"series,omitempty"`           // Name of vaccine series
	Authority        *Reference        `json:"authority,omitempty"`        // Who is responsible for publishing the recommendations
	TargetDisease    []CodeableConcept `json:"targetDisease,omitempty"`    // Vaccine preventable disease being targeted
	DoseNumberPositiveInt *int         `json:"doseNumberPositiveInt,omitempty"` // Dose number within series
	DoseNumberString string            `json:"doseNumberString,omitempty"` // Dose number within series
	SeriesDosesPositiveInt *int        `json:"seriesDosesPositiveInt,omitempty"` // Recommended number of doses for immunity
	SeriesDosesString string           `json:"seriesDosesString,omitempty"` // Recommended number of doses for immunity
}

// ImagingStudy - A set of images produced in single study
type ImagingStudy struct {
	DomainResource
	Identifier     []Identifier          `json:"identifier,omitempty"`     // Identifiers for the whole study
	Status         string                `json:"status"`                   // registered | available | cancelled | entered-in-error | unknown (REQUIRED)
	Modality       []Coding              `json:"modality,omitempty"`       // All series modality if actual acquisition modalities
	Subject        *Reference            `json:"subject"`                  // Who or what is the subject of the study (REQUIRED)
	Encounter      *Reference            `json:"encounter,omitempty"`      // Encounter with which this imaging study is associated
	Started        string                `json:"started,omitempty"`        // When the study was started
	BasedOn        []Reference           `json:"basedOn,omitempty"`        // Request fulfilled
	Referrer       *Reference            `json:"referrer,omitempty"`       // Referring physician
	Interpreter    []Reference           `json:"interpreter,omitempty"`    // Who interpreted images
	Endpoint       []Reference           `json:"endpoint,omitempty"`       // Study access endpoint
	NumberOfSeries *int                  `json:"numberOfSeries,omitempty"` // Number of Study Related Series
	NumberOfInstances *int               `json:"numberOfInstances,omitempty"` // Number of Study Related Instances
	ProcedureReference *Reference        `json:"procedureReference,omitempty"` // The performed procedure
	ProcedureCode  []CodeableConcept     `json:"procedureCode,omitempty"`  // The performed procedure code
	Location       *Reference            `json:"location,omitempty"`       // Where ImagingStudy occurred
	ReasonCode     []CodeableConcept     `json:"reasonCode,omitempty"`     // Why the study was requested
	ReasonReference []Reference          `json:"reasonReference,omitempty"` // Why was study performed
	Note           []Annotation          `json:"note,omitempty"`           // User-defined comments
	Description    string                `json:"description,omitempty"`    // Institution-generated description
	Series         []ImagingStudySeries  `json:"series,omitempty"`         // Each study has one or more series of instances
}

// ImagingStudySeries - Each study has one or more series of instances
type ImagingStudySeries struct {
	BackboneElement
	UID           string                      `json:"uid"`                    // DICOM Series Instance UID (REQUIRED)
	Number        *int                        `json:"number,omitempty"`       // Numeric identifier of this series
	Modality      *Coding                     `json:"modality"`               // The modality of the instances in the series (REQUIRED)
	Description   string                      `json:"description,omitempty"`  // A short human readable summary of the series
	NumberOfInstances *int                    `json:"numberOfInstances,omitempty"` // Number of Series Related Instances
	Endpoint      []Reference                 `json:"endpoint,omitempty"`     // Series access endpoint
	BodySite      *Coding                     `json:"bodySite,omitempty"`     // Body part examined
	Laterality    *Coding                     `json:"laterality,omitempty"`   // Body part laterality
	Specimen      []Reference                 `json:"specimen,omitempty"`     // Specimen imaged
	Started       string                      `json:"started,omitempty"`      // When the series started
	Performer     []ImagingStudySeriesPerformer `json:"performer,omitempty"`  // Who performed the series
	Instance      []ImagingStudySeriesInstance `json:"instance,omitempty"`    // A single SOP instance from the series
}

// ImagingStudySeriesPerformer - Who performed the series
type ImagingStudySeriesPerformer struct {
	BackboneElement
	Function *CodeableConcept `json:"function,omitempty"` // Type of performance
	Actor    *Reference       `json:"actor"`              // Who performed the series (REQUIRED)
}

// ImagingStudySeriesInstance - A single SOP instance from the series
type ImagingStudySeriesInstance struct {
	BackboneElement
	UID      string  `json:"uid"`                // DICOM SOP Instance UID (REQUIRED)
	SopClass *Coding `json:"sopClass"`           // DICOM class type (REQUIRED)
	Number   *int    `json:"number,omitempty"`   // The number of this instance in the series
	Title    string  `json:"title,omitempty"`    // Description of instance
}

// EpisodeOfCare - An association of a Patient with an Organization and Healthcare Provider(s)
type EpisodeOfCare struct {
	DomainResource
	Identifier        []Identifier               `json:"identifier,omitempty"`        // Business Identifier(s) relevant for this EpisodeOfCare
	Status            string                     `json:"status"`                      // planned | waitlist | active | onhold | finished | cancelled | entered-in-error (REQUIRED)
	StatusHistory     []EpisodeOfCareStatusHistory `json:"statusHistory,omitempty"`   // Past list of status codes
	Type              []CodeableConcept          `json:"type,omitempty"`              // Type/class
	Diagnosis         []EpisodeOfCareDiagnosis   `json:"diagnosis,omitempty"`         // The list of diagnosis relevant to this episode of care
	Patient           *Reference                 `json:"patient"`                     // The patient who is the focus of this episode of care (REQUIRED)
	ManagingOrganization *Reference              `json:"managingOrganization,omitempty"` // Organization that assumes care
	Period            *Period                    `json:"period,omitempty"`            // Interval during responsibility is assumed
	ReferralRequest   []Reference                `json:"referralRequest,omitempty"`   // Originating Referral Request(s)
	CareManager       *Reference                 `json:"careManager,omitempty"`       // Care manager/care coordinator for the patient
	Team              []Reference                `json:"team,omitempty"`              // Other practitioners facilitating this episode of care
	Account           []Reference                `json:"account,omitempty"`           // The set of accounts that may be used for billing
}

// EpisodeOfCareStatusHistory - Past list of status codes
type EpisodeOfCareStatusHistory struct {
	BackboneElement
	Status string  `json:"status"` // planned | waitlist | active | onhold | finished | cancelled | entered-in-error (REQUIRED)
	Period *Period `json:"period"` // Duration the EpisodeOfCare was in the specified status (REQUIRED)
}

// EpisodeOfCareDiagnosis - The list of diagnosis relevant to this episode of care
type EpisodeOfCareDiagnosis struct {
	BackboneElement
	Condition *Reference       `json:"condition"` // Conditions/problems/diagnoses this episode of care is for (REQUIRED)
	Role      *CodeableConcept `json:"role,omitempty"` // Role that this diagnosis has within the episode of care
	Rank      *int             `json:"rank,omitempty"` // Ranking of the diagnosis (for each role type)
}

// CarePlan - Healthcare plan for patient or group
type CarePlan struct {
	DomainResource
	Identifier            []Identifier         `json:"identifier,omitempty"`            // External Ids for this plan
	InstantiatesCanonical []string             `json:"instantiatesCanonical,omitempty"` // Instantiates FHIR protocol or definition
	InstantiatesURI       []string             `json:"instantiatesUri,omitempty"`       // Instantiates external protocol or definition
	BasedOn               []Reference          `json:"basedOn,omitempty"`               // Fulfills CarePlan
	Replaces              []Reference          `json:"replaces,omitempty"`              // CarePlan replaced by this CarePlan
	PartOf                []Reference          `json:"partOf,omitempty"`                // Part of referenced CarePlan
	Status                string               `json:"status"`                          // draft | active | on-hold | revoked | completed | entered-in-error | unknown (REQUIRED)
	Intent                string               `json:"intent"`                          // proposal | plan | order | option (REQUIRED)
	Category              []CodeableConcept    `json:"category,omitempty"`              // Type of plan
	Title                 string               `json:"title,omitempty"`                 // Human-friendly name for the care plan
	Description           string               `json:"description,omitempty"`           // Summary of nature of plan
	Subject               *Reference           `json:"subject"`                         // Who the care plan is for (REQUIRED)
	Encounter             *Reference           `json:"encounter,omitempty"`             // Encounter created as part of
	Period                *Period              `json:"period,omitempty"`                // Time period plan covers
	Created               string               `json:"created,omitempty"`               // Date record was first recorded
	Author                *Reference           `json:"author,omitempty"`                // Who is the designated responsible party
	Contributor           []Reference          `json:"contributor,omitempty"`           // Who provided the content of the care plan
	CareTeam              []Reference          `json:"careTeam,omitempty"`              // Who's involved in plan?
	Addresses             []Reference          `json:"addresses,omitempty"`             // Health issues this plan addresses
	SupportingInfo        []Reference          `json:"supportingInfo,omitempty"`        // Information considered as part of plan
	Goal                  []Reference          `json:"goal,omitempty"`                  // Desired outcome of plan
	Activity              []CarePlanActivity   `json:"activity,omitempty"`              // Action to occur as part of plan
	Note                  []Annotation         `json:"note,omitempty"`                  // Comments about the plan
}

// CarePlanActivity - Action to occur as part of plan
type CarePlanActivity struct {
	BackboneElement
	OutcomeCodeableConcept []CodeableConcept      `json:"outcomeCodeableConcept,omitempty"` // Results of the activity
	OutcomeReference       []Reference            `json:"outcomeReference,omitempty"`       // Appointment, Encounter, Procedure, etc.
	Progress               []Annotation           `json:"progress,omitempty"`               // Comments about the activity status/progress
	Reference              *Reference             `json:"reference,omitempty"`              // Activity details defined in specific resource
	Detail                 *CarePlanActivityDetail `json:"detail,omitempty"`                // In-line definition of activity
}

// CarePlanActivityDetail - In-line definition of activity
type CarePlanActivityDetail struct {
	BackboneElement
	Kind                      string            `json:"kind,omitempty"`                      // Appointment | CommunicationRequest | DeviceRequest | MedicationRequest | NutritionOrder | Task | ServiceRequest | VisionPrescription
	InstantiatesCanonical     []string          `json:"instantiatesCanonical,omitempty"`     // Instantiates FHIR protocol or definition
	InstantiatesURI           []string          `json:"instantiatesUri,omitempty"`           // Instantiates external protocol or definition
	Code                      *CodeableConcept  `json:"code,omitempty"`                      // Detail type of activity
	ReasonCode                []CodeableConcept `json:"reasonCode,omitempty"`                // Why activity should be done or why activity was prohibited
	ReasonReference           []Reference       `json:"reasonReference,omitempty"`           // Why activity is needed
	Goal                      []Reference       `json:"goal,omitempty"`                      // Goals this activity relates to
	Status                    string            `json:"status"`                              // not-started | scheduled | in-progress | on-hold | completed | cancelled | stopped | unknown | entered-in-error (REQUIRED)
	StatusReason              *CodeableConcept  `json:"statusReason,omitempty"`              // Reason for current status
	DoNotPerform              *bool             `json:"doNotPerform,omitempty"`              // If true, activity is prohibiting action
	ScheduledTiming           *Timing           `json:"scheduledTiming,omitempty"`           // When activity is to occur
	ScheduledPeriod           *Period           `json:"scheduledPeriod,omitempty"`           // When activity is to occur
	ScheduledString           string            `json:"scheduledString,omitempty"`           // When activity is to occur
	Location                  *Reference        `json:"location,omitempty"`                  // Where it should happen
	Performer                 []Reference       `json:"performer,omitempty"`                 // Who will be responsible?
	ProductCodeableConcept    *CodeableConcept  `json:"productCodeableConcept,omitempty"`    // What is to be administered/supplied
	ProductReference          *Reference        `json:"productReference,omitempty"`          // What is to be administered/supplied
	DailyAmount               *Quantity         `json:"dailyAmount,omitempty"`               // How to consume/day?
	Quantity                  *Quantity         `json:"quantity,omitempty"`                  // How much to administer/supply/consume
	Description               string            `json:"description,omitempty"`               // Extra info describing activity to perform
}

// QuestionnaireResponse - A structured set of questions and their answers
type QuestionnaireResponse struct {
	DomainResource
	Identifier  *Identifier                   `json:"identifier,omitempty"`  // Unique id for this set of answers
	BasedOn     []Reference                   `json:"basedOn,omitempty"`     // Request fulfilled by this QuestionnaireResponse
	PartOf      []Reference                   `json:"partOf,omitempty"`      // Part of this action
	Questionnaire string                      `json:"questionnaire,omitempty"` // Form being answered
	Status      string                        `json:"status"`                // in-progress | completed | amended | entered-in-error | stopped (REQUIRED)
	Subject     *Reference                    `json:"subject,omitempty"`     // The subject of the questions
	Encounter   *Reference                    `json:"encounter,omitempty"`   // Encounter created as part of
	Authored    string                        `json:"authored,omitempty"`    // Date the answers were gathered
	Author      *Reference                    `json:"author,omitempty"`      // Person who received and recorded the answers
	Source      *Reference                    `json:"source,omitempty"`      // The person who answered the questions
	Item        []QuestionnaireResponseItem   `json:"item,omitempty"`        // Groups and questions
}

// QuestionnaireResponseItem - Groups and questions
type QuestionnaireResponseItem struct {
	BackboneElement
	LinkId     string                         `json:"linkId"`               // Pointer to specific item from Questionnaire (REQUIRED)
	Definition string                         `json:"definition,omitempty"` // ElementDefinition - details for the item
	Text       string                         `json:"text,omitempty"`       // Name for group or question text
	Answer     []QuestionnaireResponseItemAnswer `json:"answer,omitempty"` // The response(s) to the question
	Item       []QuestionnaireResponseItem    `json:"item,omitempty"`       // Nested questionnaire response items
}

// QuestionnaireResponseItemAnswer - The response(s) to the question
type QuestionnaireResponseItemAnswer struct {
	BackboneElement
	ValueBoolean    *bool            `json:"valueBoolean,omitempty"`    // Single-valued answer to the question
	ValueDecimal    *float64         `json:"valueDecimal,omitempty"`    // Single-valued answer to the question
	ValueInteger    *int             `json:"valueInteger,omitempty"`    // Single-valued answer to the question
	ValueDate       string           `json:"valueDate,omitempty"`       // Single-valued answer to the question
	ValueDateTime   string           `json:"valueDateTime,omitempty"`   // Single-valued answer to the question
	ValueTime       string           `json:"valueTime,omitempty"`       // Single-valued answer to the question
	ValueString     string           `json:"valueString,omitempty"`     // Single-valued answer to the question
	ValueURI        string           `json:"valueUri,omitempty"`        // Single-valued answer to the question
	ValueAttachment *Attachment      `json:"valueAttachment,omitempty"` // Single-valued answer to the question
	ValueCoding     *Coding          `json:"valueCoding,omitempty"`     // Single-valued answer to the question
	ValueQuantity   *Quantity        `json:"valueQuantity,omitempty"`   // Single-valued answer to the question
	ValueReference  *Reference       `json:"valueReference,omitempty"`  // Single-valued answer to the question
	Item            []QuestionnaireResponseItem `json:"item,omitempty"` // Nested groups and questions
}

// RelatedPerson - A person related to the patient
type RelatedPerson struct {
	DomainResource
	Identifier      []Identifier      `json:"identifier,omitempty"`      // A human identifier for this person
	Active          *bool             `json:"active,omitempty"`          // Whether this related person's record is in active use
	Patient         *Reference        `json:"patient"`                   // The patient this person is related to (REQUIRED)
	Relationship    []CodeableConcept `json:"relationship,omitempty"`    // The nature of the relationship
	Name            []HumanName       `json:"name,omitempty"`            // A name associated with the person
	Telecom         []ContactPoint    `json:"telecom,omitempty"`         // A contact detail for the person
	Gender          string            `json:"gender,omitempty"`          // male | female | other | unknown
	BirthDate       string            `json:"birthDate,omitempty"`       // The date on which the related person was born
	Address         []Address         `json:"address,omitempty"`         // Address where the related person can be contacted or visited
	Photo           []Attachment      `json:"photo,omitempty"`           // Image of the person
	Period          *Period           `json:"period,omitempty"`          // Period of time that this relationship is considered valid
	Communication   []RelatedPersonCommunication `json:"communication,omitempty"` // A language which may be used to communicate with about the patient's health
}

// RelatedPersonCommunication - A language which may be used to communicate with about the patient's health
type RelatedPersonCommunication struct {
	BackboneElement
	Language  *CodeableConcept `json:"language"`            // The language which can be used to communicate with the patient about his or her health (REQUIRED)
	Preferred *bool            `json:"preferred,omitempty"` // Language preference indicator
}
