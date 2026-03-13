package satusehat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
	ss "github.com/faisalhardin/medilink/internal/entity/model/satusehat"
	"github.com/faisalhardin/medilink/internal/entity/repo/cache"
	"github.com/faisalhardin/medilink/internal/repo/satusehat"
)

// SatuSehatUC handles Satu Sehat FHIR API integration use cases
type SatuSehatUC struct {
	client *satusehat.Client
	cfg    *config.Config
}

// NewSatuSehatUC creates a new Satu Sehat use case handler
func NewSatuSehatUC(cfg *config.Config, cache cache.Caching) *SatuSehatUC {
	return &SatuSehatUC{
		client: satusehat.NewClient(cfg, cache),
		cfg:    cfg,
	}
}

// EncounterData contains all data needed to create a complete patient encounter in Satu Sehat
type EncounterData struct {
	// Patient information
	PatientNIK  string
	PatientName string
	PatientID   string // If already known, otherwise will be searched

	// Practitioner information
	PractitionerNIK  string
	PractitionerName string
	PractitionerID   string // If already known

	// Location information
	LocationName string
	LocationID   string // If already known

	// Encounter details
	EncounterClass   string    // e.g., "AMB" (ambulatory), "IMP" (inpatient), "EMER" (emergency)
	EncounterStatus  string    // e.g., "arrived", "in-progress", "finished"
	EncounterPeriod  time.Time // Start time of encounter
	ServiceType      string    // Type of service provided
	EncounterEndTime *time.Time // Optional: End time if encounter is finished

	// Diagnosis information
	DiagnosisCode    string // ICD-10 code
	DiagnosisDisplay string // Human-readable diagnosis
	IsPrimaryDiag    bool   // Whether this is the primary diagnosis
}

// EncounterResult contains the IDs of created resources
type EncounterResult struct {
	PatientID      string
	PractitionerID string
	LocationID     string
	EncounterID    string
	ConditionID    string
	Error          error
}

// CreatePatientEncounter creates a complete patient encounter in Satu Sehat
// This is the main integration point that orchestrates the creation of all required resources
func (uc *SatuSehatUC) CreatePatientEncounter(ctx context.Context, data *EncounterData) (*EncounterResult, error) {
	result := &EncounterResult{}

	// Check if integration is enabled
	if !uc.cfg.SatuSehatConfig.Enabled {
		return nil, fmt.Errorf("satu sehat integration is not enabled")
	}

	// 1. Get or create Patient
	patientID, err := uc.getOrCreatePatient(ctx, data)
	if err != nil {
		result.Error = fmt.Errorf("patient lookup/creation failed: %w", err)
		return result, result.Error
	}
	result.PatientID = patientID

	// 2. Get or create Practitioner
	practitionerID, err := uc.getOrCreatePractitioner(ctx, data)
	if err != nil {
		result.Error = fmt.Errorf("practitioner lookup/creation failed: %w", err)
		return result, result.Error
	}
	result.PractitionerID = practitionerID

	// 3. Get or create Location
	locationID, err := uc.getOrCreateLocation(ctx, data)
	if err != nil {
		result.Error = fmt.Errorf("location lookup/creation failed: %w", err)
		return result, result.Error
	}
	result.LocationID = locationID

	// 4. Create Encounter
	encounterID, err := uc.createEncounter(ctx, data, patientID, practitionerID, locationID)
	if err != nil {
		result.Error = fmt.Errorf("encounter creation failed: %w", err)
		return result, result.Error
	}
	result.EncounterID = encounterID

	// 5. Create Condition (Diagnosis)
	if data.DiagnosisCode != "" {
		conditionID, err := uc.createCondition(ctx, data, patientID, encounterID)
		if err != nil {
			result.Error = fmt.Errorf("condition creation failed: %w", err)
			return result, result.Error
		}
		result.ConditionID = conditionID
	}

	return result, nil
}

// getOrCreatePatient retrieves existing patient by NIK or creates a new one
func (uc *SatuSehatUC) getOrCreatePatient(ctx context.Context, data *EncounterData) (string, error) {
	// If patient ID is already provided, return it
	if data.PatientID != "" {
		return data.PatientID, nil
	}

	// Search for patient by NIK
	params := url.Values{}
	params.Set("identifier", fmt.Sprintf("https://fhir.kemkes.go.id/id/nik|%s", data.PatientNIK))

	var bundle ss.Bundle
	err := uc.client.Search(ctx, "Patient", params, &bundle)
	if err != nil {
		return "", fmt.Errorf("patient search failed: %w", err)
	}

	// If patient found, return ID
	if bundle.Total != nil && *bundle.Total > 0 {
		var patient ss.Patient
		if err := json.Unmarshal(bundle.Entry[0].Resource, &patient); err != nil {
			return "", fmt.Errorf("failed to parse patient: %w", err)
		}
		return patient.ID, nil
	}

	// Patient not found - in Satu Sehat, patients must already exist (registered by government)
	// Return error instead of creating
	return "", fmt.Errorf("patient with NIK %s not found in Satu Sehat", data.PatientNIK)
}

// getOrCreatePractitioner retrieves existing practitioner or creates a new one
func (uc *SatuSehatUC) getOrCreatePractitioner(ctx context.Context, data *EncounterData) (string, error) {
	// If practitioner ID is already provided, return it
	if data.PractitionerID != "" {
		return data.PractitionerID, nil
	}

	// Search for practitioner by NIK
	params := url.Values{}
	params.Set("identifier", fmt.Sprintf("https://fhir.kemkes.go.id/id/nik|%s", data.PractitionerNIK))

	var bundle ss.Bundle
	err := uc.client.Search(ctx, "Practitioner", params, &bundle)
	if err != nil {
		return "", fmt.Errorf("practitioner search failed: %w", err)
	}

	// If practitioner found, return ID
	if bundle.Total != nil && *bundle.Total > 0 {
		var practitioner ss.Practitioner
		if err := json.Unmarshal(bundle.Entry[0].Resource, &practitioner); err != nil {
			return "", fmt.Errorf("failed to parse practitioner: %w", err)
		}
		return practitioner.ID, nil
	}

	// Practitioner not found - create new one
	practitioner := &ss.Practitioner{
		DomainResource: ss.DomainResource{
			Resource: ss.Resource{
				ResourceType: "Practitioner",
			},
		},
		Active: boolPtr(true),
		Identifier: []ss.Identifier{
			{
				Use:    "official",
				System: "https://fhir.kemkes.go.id/id/nik",
				Value:  data.PractitionerNIK,
			},
		},
		Name: []ss.HumanName{
			{
				Use:  "official",
				Text: data.PractitionerName,
			},
		},
	}

	var result ss.Practitioner
	err = uc.client.Post(ctx, "Practitioner", practitioner, &result)
	if err != nil {
		return "", fmt.Errorf("practitioner creation failed: %w", err)
	}

	return result.ID, nil
}

// getOrCreateLocation retrieves existing location or creates a new one
func (uc *SatuSehatUC) getOrCreateLocation(ctx context.Context, data *EncounterData) (string, error) {
	// If location ID is already provided, return it
	if data.LocationID != "" {
		return data.LocationID, nil
	}

	// Create new location
	orgID := uc.client.GetOrganizationID()

	location := &ss.Location{
		DomainResource: ss.DomainResource{
			Resource: ss.Resource{
				ResourceType: "Location",
			},
		},
		Status: "active",
		Name:   data.LocationName,
		Identifier: []ss.Identifier{
			{
				System: fmt.Sprintf("http://sys-ids.kemkes.go.id/location/%s", orgID),
				Value:  fmt.Sprintf("%s-LOC-%d", orgID, time.Now().Unix()),
			},
		},
		ManagingOrganization: &ss.Reference{
			Reference: fmt.Sprintf("Organization/%s", orgID),
		},
		PhysicalType: &ss.CodeableConcept{
			Coding: []ss.Coding{
				{
					System:  "http://terminology.hl7.org/CodeSystem/location-physical-type",
					Code:    "ro",
					Display: "Room",
				},
			},
		},
	}

	var result ss.Location
	err := uc.client.Post(ctx, "Location", location, &result)
	if err != nil {
		return "", fmt.Errorf("location creation failed: %w", err)
	}

	return result.ID, nil
}

// createEncounter creates a new encounter record
func (uc *SatuSehatUC) createEncounter(ctx context.Context, data *EncounterData, patientID, practitionerID, locationID string) (string, error) {
	orgID := uc.client.GetOrganizationID()

	// Default to ambulatory if not specified
	encounterClass := data.EncounterClass
	if encounterClass == "" {
		encounterClass = "AMB"
	}

	// Default to arrived if not specified
	encounterStatus := data.EncounterStatus
	if encounterStatus == "" {
		encounterStatus = "arrived"
	}

	encounter := &ss.Encounter{
		DomainResource: ss.DomainResource{
			Resource: ss.Resource{
				ResourceType: "Encounter",
			},
		},
		Status: encounterStatus,
		Class: &ss.Coding{
			System:  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			Code:    encounterClass,
			Display: getEncounterClassDisplay(encounterClass),
		},
		Subject: &ss.Reference{
			Reference: fmt.Sprintf("Patient/%s", patientID),
			Display:   "Patient",
		},
		Participant: []ss.EncounterParticipant{
			{
				Type: []ss.CodeableConcept{
					{
						Coding: []ss.Coding{
							{
								System:  "http://terminology.hl7.org/CodeSystem/v3-ParticipationType",
								Code:    "ATND",
								Display: "attender",
							},
						},
					},
				},
				Individual: &ss.Reference{
					Reference: fmt.Sprintf("Practitioner/%s", practitionerID),
					Display:   "Practitioner",
				},
			},
		},
		Period: &ss.Period{
			Start: data.EncounterPeriod.Format(time.RFC3339),
		},
		Location: []ss.EncounterLocation{
			{
				Location: &ss.Reference{
					Reference: fmt.Sprintf("Location/%s", locationID),
					Display:   "Location",
				},
			},
		},
		ServiceProvider: &ss.Reference{
			Reference: fmt.Sprintf("Organization/%s", orgID),
			Display:   "Organization",
		},
	}

	// Add end time if encounter is finished
	if data.EncounterEndTime != nil {
		encounter.Period.End = data.EncounterEndTime.Format(time.RFC3339)
	}

	var result ss.Encounter
	err := uc.client.Post(ctx, "Encounter", encounter, &result)
	if err != nil {
		return "", fmt.Errorf("encounter creation failed: %w", err)
	}

	return result.ID, nil
}

// createCondition creates a condition (diagnosis) record
func (uc *SatuSehatUC) createCondition(ctx context.Context, data *EncounterData, patientID, encounterID string) (string, error) {
	condition := &ss.Condition{
		DomainResource: ss.DomainResource{
			Resource: ss.Resource{
				ResourceType: "Condition",
			},
		},
		ClinicalStatus: &ss.CodeableConcept{
			Coding: []ss.Coding{
				{
					System: "http://terminology.hl7.org/CodeSystem/condition-clinical",
					Code:   "active",
				},
			},
		},
		Category: []ss.CodeableConcept{
			{
				Coding: []ss.Coding{
					{
						System:  "http://terminology.hl7.org/CodeSystem/condition-category",
						Code:    "encounter-diagnosis",
						Display: "Encounter Diagnosis",
					},
				},
			},
		},
		Code: &ss.CodeableConcept{
			Coding: []ss.Coding{
				{
					System:  "http://hl7.org/fhir/sid/icd-10",
					Code:    data.DiagnosisCode,
					Display: data.DiagnosisDisplay,
				},
			},
		},
		Subject: &ss.Reference{
			Reference: fmt.Sprintf("Patient/%s", patientID),
			Display:   "Patient",
		},
		Encounter: &ss.Reference{
			Reference: fmt.Sprintf("Encounter/%s", encounterID),
			Display:   "Encounter",
		},
	}

	var result ss.Condition
	err := uc.client.Post(ctx, "Condition", condition, &result)
	if err != nil {
		return "", fmt.Errorf("condition creation failed: %w", err)
	}

	return result.ID, nil
}

// SyncPatientVisit syncs a Medilink patient visit to Satu Sehat
// This is an example of how to integrate with existing Medilink data structures
// NOTE: This method needs to be customized based on your actual data model
// You'll need to fetch patient, practitioner, and diagnosis information from your database
func (uc *SatuSehatUC) SyncPatientVisit(ctx context.Context, visit *model.TrxPatientVisit, patientNIK, practitionerNIK, diagnosisCode, diagnosisText string) (*EncounterResult, error) {
	// Map Medilink visit to Satu Sehat encounter data
	data := &EncounterData{
		PatientNIK:       patientNIK,          // Patient's NIK from your database
		PractitionerNIK:  practitionerNIK,     // Practitioner's NIK from your database
		LocationName:     "Consultation Room", // Location name from your database
		EncounterClass:   "AMB",               // Default to ambulatory
		EncounterStatus:  "finished",          // Based on visit status
		EncounterPeriod:  visit.CreateTime,    // Using visit create time
		DiagnosisCode:    diagnosisCode,       // ICD-10 code from your diagnosis data
		DiagnosisDisplay: diagnosisText,       // Diagnosis text from your diagnosis data
		IsPrimaryDiag:    true,
	}

	// Create encounter in Satu Sehat
	return uc.CreatePatientEncounter(ctx, data)
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

func getEncounterClassDisplay(code string) string {
	displays := map[string]string{
		"AMB":  "ambulatory",
		"EMER": "emergency",
		"FLD":  "field",
		"HH":   "home health",
		"IMP":  "inpatient encounter",
		"ACUTE": "inpatient acute",
		"NONAC": "inpatient non-acute",
		"OBSENC": "observation encounter",
		"PRENC": "pre-admission",
		"SS":   "short stay",
		"VR":   "virtual",
	}
	if display, ok := displays[code]; ok {
		return display
	}
	return code
}
