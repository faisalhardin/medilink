# Satu Sehat FHIR Data Structures

This package contains Go data structures for Satu Sehat FHIR (Fast Healthcare Interoperability Resources) integration, based on the official [Satu Sehat FHIR documentation](https://satusehat.kemkes.go.id/platform/docs/id/fhir/).

## Package Structure

```
satusehat/
├── types.go              # Shared FHIR data types (Identifier, Coding, CodeableConcept, etc.)
├── framework.go          # FHIR framework types (Resource, DomainResource, Bundle)
├── prerequisites.go      # Prerequisites resources (Organization, Location, Practitioner, Patient)
├── encounter.go          # Encounter resource and nested types
├── condition.go          # Condition resource for diagnoses
├── observation.go        # Observation resource for measurements and assessments
├── composition.go        # Composition resource for clinical documents
├── procedure.go          # Procedure resource
├── medication.go         # Medication-related resources (Medication, MedicationRequest, MedicationDispense, etc.)
├── other_resources.go    # Additional interoperability resources
├── responses.go          # OperationOutcome and response wrappers
└── README.md            # This file
```

## Implemented Resources

### Prerequisites (Onboarding)
- **Organization** - Healthcare facility organizational structure
- **Location** - Physical locations where services are provided
- **Practitioner** - Healthcare practitioners (doctors, nurses, etc.)
- **Patient** - Patient demographic information

### Core Clinical Resources
- **Encounter** - Patient visits and healthcare interactions
- **Condition** - Diagnoses and health conditions
- **Observation** - Measurements, lab results, vital signs
- **Composition** - Clinical documents and summaries

### Procedures and Medications
- **Procedure** - Surgical and medical procedures
- **Medication** - Medication definitions
- **MedicationRequest** - Prescription orders
- **MedicationDispense** - Medication dispensing records
- **MedicationAdministration** - Medication administration records
- **MedicationStatement** - Patient medication history

### Diagnostic and Laboratory
- **ServiceRequest** - Orders for diagnostic services
- **Specimen** - Laboratory specimens
- **DiagnosticReport** - Diagnostic test reports
- **ImagingStudy** - Medical imaging studies

### Care Management
- **EpisodeOfCare** - Patient care episodes
- **CarePlan** - Care plans and treatment protocols
- **ClinicalImpression** - Clinical assessments
- **AllergyIntolerance** - Allergy and intolerance records
- **Immunization** - Vaccination records

### Supporting Resources
- **QuestionnaireResponse** - Structured questionnaire answers
- **RelatedPerson** - People related to the patient

### Framework and Utilities
- **Bundle** - Collections of resources (search results, transactions, batches)
- **OperationOutcome** - Error and validation responses
- **TokenResponse** - OAuth2 authentication tokens

## Key Features

### JSON-Only Tags
All structs use `json` tags only (no `xorm` or database tags) for HTTP request/response serialization:

```go
type Patient struct {
    DomainResource
    Identifier []Identifier `json:"identifier,omitempty"`
    Name       []HumanName  `json:"name,omitempty"`
    Gender     string       `json:"gender,omitempty"`
    BirthDate  string       `json:"birthDate,omitempty"`
    // ...
}
```

### Date/Time Format
All date and time fields use `string` type to preserve exact ISO8601 format as required by Satu Sehat:
- **Dates**: `YYYY-MM-DD` (e.g., `"2023-08-23"`)
- **DateTime**: `YYYY-MM-DDThh:mm:ss+zz:zz` (e.g., `"2023-08-23T10:35:00+00:00"`)
- **Note**: Satu Sehat requires UTC+00 timezone

### Optional Fields
Optional/repeating fields use pointers or slices:
```go
Active     *bool             `json:"active,omitempty"`      // Optional boolean
Identifier []Identifier      `json:"identifier,omitempty"`  // Repeating element
```

### FHIR References
Resource references use the `Reference` struct with format `"ResourceType/id"`:
```go
Subject: &Reference{
    Reference: "Patient/100000030009",
    Display:   "Budi Santoso",
}
```

## Usage Examples

### Creating a Patient Resource

```go
import "github.com/faisalhardin/medilink/internal/entity/model/satusehat"

patient := &satusehat.Patient{
    DomainResource: satusehat.DomainResource{
        Resource: satusehat.Resource{
            ResourceType: "Patient",
        },
    },
    Identifier: []satusehat.Identifier{
        {
            Use:    "official",
            System: "https://fhir.kemkes.go.id/id/nik",
            Value:  "3201234567890123",
        },
    },
    Name: []satusehat.HumanName{
        {
            Use:  "official",
            Text: "John Doe",
        },
    },
    Gender:    "male",
    BirthDate: "1990-01-15",
}
```

### Creating an Encounter

```go
encounter := &satusehat.Encounter{
    DomainResource: satusehat.DomainResource{
        Resource: satusehat.Resource{
            ResourceType: "Encounter",
        },
    },
    Identifier: []satusehat.Identifier{
        {
            System: "http://sys-ids.kemkes.go.id/encounter/10000004",
            Use:    "official",
            Value:  "P20240001",
        },
    },
    Status: "finished",
    Class: &satusehat.Coding{
        System:  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
        Code:    "AMB",
        Display: "ambulatory",
    },
    Subject: &satusehat.Reference{
        Reference: "Patient/100000030009",
        Display:   "John Doe",
    },
    Period: &satusehat.Period{
        Start: "2024-01-15T08:00:00+00:00",
        End:   "2024-01-15T09:00:00+00:00",
    },
}
```

### Creating a Bundle for Search Results

```go
bundle := &satusehat.Bundle{
    Resource: satusehat.Resource{
        ResourceType: "Bundle",
    },
    Type:      "searchset",
    Total:     intPtr(2),
    Timestamp: "2024-01-15T10:00:00+00:00",
    Entry: []satusehat.BundleEntry{
        {
            FullURL: "https://api.satusehat.kemkes.go.id/fhir-r4/v1/Patient/100000030009",
            Resource: json.RawMessage(`{"resourceType":"Patient",...}`),
        },
    },
}
```

## Integration with Existing Models

These FHIR DTOs are **separate from database models**. Your existing domain models in `internal/entity/model/` (e.g., `TrxPatientVisit`, `MstPatientInstitution`) remain unchanged.

### Mapping Strategy

Create mapper functions in your usecase or service layer:

```go
// Example: Map domain model to FHIR Patient
func MapToFHIRPatient(mstPatient *model.MstPatientInstitution) *satusehat.Patient {
    return &satusehat.Patient{
        DomainResource: satusehat.DomainResource{
            Resource: satusehat.Resource{
                ResourceType: "Patient",
            },
        },
        Name: []satusehat.HumanName{
            {
                Use:  "official",
                Text: mstPatient.Name,
            },
        },
        Gender:    mstPatient.Sex,
        BirthDate: mstPatient.DateOfBirth.Format("2006-01-02"),
        // ... map other fields
    }
}

// Example: Map FHIR Patient to domain model
func MapFromFHIRPatient(fhirPatient *satusehat.Patient) *model.MstPatientInstitution {
    birthDate, _ := time.Parse("2006-01-02", fhirPatient.BirthDate)
    
    return &model.MstPatientInstitution{
        Name:        fhirPatient.Name[0].Text,
        Sex:         fhirPatient.Gender,
        DateOfBirth: birthDate,
        // ... map other fields
    }
}
```

## Validation

According to Satu Sehat requirements:

1. **Date/Time Format**: All dates must be >= `2014-06-03`
2. **Timezone**: Always use UTC+00 (`+00:00`)
3. **Required Fields**: Fields marked with `(REQUIRED)` in comments must be populated
4. **NIK**: Use dummy data for sandbox environment (see [Patient documentation](https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/patient/))

## Testing

Satu Sehat provides dummy data for sandbox testing. Example patient IDs:
- `P02478375538` - Ardianto Putra (male, 1992-01-09)
- `P03647103112` - Claudia Sintia (female, 1989-11-03)

See the full list in the [Patient resource documentation](https://satusehat.kemkes.go.id/platform/docs/id/fhir/resources/patient/).

## References

- [Satu Sehat FHIR Documentation](https://satusehat.kemkes.go.id/platform/docs/id/fhir/)
- [HL7 FHIR R4 Specification](https://www.hl7.org/fhir/)
- [Satu Sehat Postman Collection](https://satusehat.kemkes.go.id/platform/docs/id/postman-workshop/)
- [Terminologi SATUSEHAT](https://satusehat.kemkes.go.id/platform/docs/id/terminology/standar-terminologi/)

## Next Steps

1. **Authentication**: Implement OAuth2 client for Satu Sehat API
2. **API Client**: Create HTTP client wrapper for FHIR endpoints
3. **Mappers**: Build mapping layer between domain models and FHIR DTOs
4. **Validation**: Add validation logic for required fields and formats
5. **Error Handling**: Parse and handle `OperationOutcome` responses

## Contributing

When adding new FHIR resources:
1. Follow existing naming conventions
2. Use `json` tags matching FHIR property names exactly
3. Document required fields with `(REQUIRED)` comments
4. Use string types for dates/times
5. Include nested types as separate structs with `BackboneElement`
6. Add comments with links to official documentation
