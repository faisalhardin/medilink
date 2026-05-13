package model

import (
	"database/sql"
	"time"

	"github.com/volatiletech/null/v8"
)

const (
	TRX_ANAMNESA_TABLE = "mdl_trx_anamnesa"
)

// Allowed enum values — enforced by DB CHECK/ENUM constraints; constants here
// allow the usecase to validate before hitting the DB.
const (
	HeightMeasurementBerdiri   = "berdiri"
	HeightMeasurementTelentang = "telentang"

	ConsciousnessComposMentis = "COMPOS MENTIS"
	ConsciousnessSomnolen     = "SOMNOLEN"
	ConsciousnessSopor        = "SOPOR"
	ConsciousnessComa         = "COMA"

	HeartRhythmRegular  = "REGULAR"
	HeartRhythmIregular = "IREGULAR"

	TriageGawatDarurat      = "GAWAT DARURAT"
	TriageDarurat           = "DARURAT"
	TriageTidakGawatDarurat = "TIDAK GAWAT DARURAT"
	TriageMeninggal         = "MENINGGAL"

	PainQualityTekanan     = "tekanan"
	PainQualityTerbakar    = "terbakar"
	PainQualityMelilit     = "melilit"
	PainQualityTertusuk    = "tertusuk"
	PainQualityDiiris      = "diiris"
	PainQualityMencengkram = "mencengkram"

	PainPatternIntermittent = "intermittent"
	PainPatternContinuous   = "continuous"
)

// TrxAnamnesa is the DB write/read model for mdl_trx_anamnesa.
// VSMAP, VSBMI, VSBMIResult are computed by the usecase before write;
// GCSTotal is a stored generated column (read-only via xorm:"<-").
// All json tags are "-" — callers must use ToResponse() for API output.
type TrxAnamnesa struct {
	ID            string  `xorm:"'id' pk"          json:"-"`
	VisitID       int64   `xorm:"'visit_id'"       json:"-"`
	InstitutionID int64   `xorm:"'institution_id'" json:"-"`
	DoctorID      *string `xorm:"'doctor_id' null" json:"-"`
	NurseID       *string `xorm:"'nurse_id' null"  json:"-"`

	// Chief complaint
	ChiefComplaint     null.String `xorm:"'chief_complaint' null"     json:"-"`
	SecondaryComplaint null.String `xorm:"'secondary_complaint' null" json:"-"`
	HistoryOfIllness   null.String `xorm:"'history_of_illness' null"  json:"-"`

	// Illness duration
	IllnessYears  int16 `xorm:"'illness_years'"  json:"-"`
	IllnessMonths int16 `xorm:"'illness_months'" json:"-"`
	IllnessDays   int16 `xorm:"'illness_days'"   json:"-"`

	// Vital signs
	VSSystolic               *int16      `xorm:"'vs_systolic' null"               json:"-"`
	VSDiastolic              *int16      `xorm:"'vs_diastolic' null"              json:"-"`
	VSPulse                  *int16      `xorm:"'vs_pulse' null"                  json:"-"`
	VSTemperature            *float32    `xorm:"'vs_temperature' null"            json:"-"`
	VSRespiratoryRate        *int16      `xorm:"'vs_respiratory_rate' null"       json:"-"`
	VSOxygenSaturation       *int16      `xorm:"'vs_oxygen_saturation' null"      json:"-"`
	VSMAP                    *int16      `xorm:"'vs_map' null"                    json:"-"`
	VSWeight                 *float32    `xorm:"'vs_weight' null"                 json:"-"`
	VSHeight                 *float32    `xorm:"'vs_height' null"                 json:"-"`
	VSBMI                    *float32    `xorm:"'vs_bmi' null"                    json:"-"`
	VSBMIResult              null.String `xorm:"'vs_bmi_result' null"             json:"-"`
	VSHeightMeasurement      null.String `xorm:"'vs_height_measurement' null"     json:"-"`
	VSAbdominalCircumference *float32    `xorm:"'vs_abdominal_circumference' null" json:"-"`
	VSConsciousness          null.String `xorm:"'vs_consciousness' null"          json:"-"`
	VSHeartRhythm            null.String `xorm:"'vs_heart_rhythm' null"           json:"-"`
	VSPregnancyStatus        *bool       `xorm:"'vs_pregnancy_status' null"       json:"-"`
	VSTriage                 null.String `xorm:"'vs_triage' null"                 json:"-"`

	// GCS
	GCSEye    *int16 `xorm:"'gcs_eye' null"    json:"-"`
	GCSVerbal *int16 `xorm:"'gcs_verbal' null" json:"-"`
	GCSMotor  *int16 `xorm:"'gcs_motor' null"  json:"-"`
	// GCSTotal is read-only — generated column in Postgres
	GCSTotal *int16 `xorm:"'gcs_total' <-" json:"-"`

	// Pain assessment
	PainHasPain  *bool          `xorm:"'pain_has_pain' null"  json:"-"`
	PainTrigger  sql.NullString `xorm:"'pain_trigger' null"   json:"-"`
	PainQuality  sql.NullString `xorm:"'pain_quality' null"   json:"-"`
	PainLocation sql.NullString `xorm:"'pain_location' null"  json:"-"`
	PainScale    *int16         `xorm:"'pain_scale' null"     json:"-"`
	PainPattern  sql.NullString `xorm:"'pain_pattern' null"   json:"-"`

	CreatedAt time.Time `xorm:"'created_at' created" json:"-"`
	UpdatedAt time.Time `xorm:"'updated_at' updated" json:"-"`
}

// ToResponse converts the DB row to the API response DTO.
func (a *TrxAnamnesa) ToResponse() AnamnesaResponse {
	return AnamnesaResponse{
		ID:                 a.ID,
		VisitID:            a.VisitID,
		DoctorID:           a.DoctorID,
		NurseID:            a.NurseID,
		ChiefComplaint:     a.ChiefComplaint,
		SecondaryComplaint: a.SecondaryComplaint,
		HistoryOfIllness:   a.HistoryOfIllness,
		IllnessDuration: IllnessDurationResponse{
			Years:  a.IllnessYears,
			Months: a.IllnessMonths,
			Days:   a.IllnessDays,
		},
		VitalSigns: VitalSignsResponse{
			Systolic:               a.VSSystolic,
			Diastolic:              a.VSDiastolic,
			Pulse:                  a.VSPulse,
			Temperature:            a.VSTemperature,
			RespiratoryRate:        a.VSRespiratoryRate,
			OxygenSaturation:       a.VSOxygenSaturation,
			MAP:                    a.VSMAP,
			Weight:                 a.VSWeight,
			Height:                 a.VSHeight,
			BMI:                    a.VSBMI,
			BMIResult:              a.VSBMIResult,
			HeightMeasurement:      a.VSHeightMeasurement,
			AbdominalCircumference: a.VSAbdominalCircumference,
			Consciousness:          a.VSConsciousness,
			HeartRhythm:            a.VSHeartRhythm,
			PregnancyStatus:        a.VSPregnancyStatus,
			Triage:                 a.VSTriage,
		},
		GCS: GCSResponse{
			Eye:    a.GCSEye,
			Verbal: a.GCSVerbal,
			Motor:  a.GCSMotor,
			Total:  a.GCSTotal,
		},
		PainAssessment: PainAssessmentResponse{
			HasPain:  a.PainHasPain,
			Trigger:  a.PainTrigger.String,
			Quality:  a.PainQuality.String,
			Location: a.PainLocation.String,
			Scale:    a.PainScale,
			Pattern:  a.PainPattern.String,
		},
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

type AnamnesaResponse struct {
	ID                 string                  `json:"id"`
	VisitID            int64                   `json:"visit_id"`
	DoctorID           *string                 `json:"doctor_id,omitempty"`
	NurseID            *string                 `json:"nurse_id,omitempty"`
	ChiefComplaint     null.String             `json:"chief_complaint"`
	SecondaryComplaint null.String             `json:"secondary_complaint"`
	HistoryOfIllness   null.String             `json:"history_of_illness"`
	IllnessDuration    IllnessDurationResponse `json:"illness_duration"`
	VitalSigns         VitalSignsResponse      `json:"vital_signs"`
	GCS                GCSResponse             `json:"gcs"`
	PainAssessment     PainAssessmentResponse  `json:"pain_assessment"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

// AnamnesaDetailedResponse is the viewing payload for GET anamnesa with practitioner
// display names (same shape as AnamnesaResponse plus doctor_name / nurse_name).
type AnamnesaDetailedResponse struct {
	AnamnesaResponse
	DoctorName null.String `json:"doctor_name"`
	NurseName  null.String `json:"nurse_name"`
}

// TrxAnamnesaDetailRow is the read-model from GetDetailedByVisitID (joins mst_doctor / mst_nurse).
// Embedded TrxAnamnesa maps columns from SELECT a.*; doctor_name / nurse_name come from the joins.
type TrxAnamnesaDetailRow struct {
	TrxAnamnesa `xorm:"extends"`
	DoctorName  string `xorm:"'doctor_name'"`
	NurseName   string `xorm:"'nurse_name'"`
}

// ToDetailedResponse maps the joined row to the API DTO.
func (r *TrxAnamnesaDetailRow) ToDetailedResponse() *AnamnesaDetailedResponse {
	base := r.TrxAnamnesa.ToResponse()
	return &AnamnesaDetailedResponse{
		AnamnesaResponse: base,
		DoctorName:       null.StringFrom(r.DoctorName),
		NurseName:        null.StringFrom(r.NurseName),
	}
}

type IllnessDurationResponse struct {
	Years  int16 `json:"years"`
	Months int16 `json:"months"`
	Days   int16 `json:"days"`
}

type VitalSignsResponse struct {
	Systolic               *int16      `json:"systolic,omitempty"`
	Diastolic              *int16      `json:"diastolic,omitempty"`
	Pulse                  *int16      `json:"pulse,omitempty"`
	Temperature            *float32    `json:"temperature,omitempty"`
	RespiratoryRate        *int16      `json:"respiratory_rate,omitempty"`
	OxygenSaturation       *int16      `json:"oxygen_saturation,omitempty"`
	MAP                    *int16      `json:"map,omitempty"`
	Weight                 *float32    `json:"weight,omitempty"`
	Height                 *float32    `json:"height,omitempty"`
	BMI                    *float32    `json:"bmi,omitempty"`
	BMIResult              null.String `json:"bmi_result"`
	HeightMeasurement      null.String `json:"height_measurement"`
	AbdominalCircumference *float32    `json:"abdominal_circumference,omitempty"`
	Consciousness          null.String `json:"consciousness"`
	HeartRhythm            null.String `json:"heart_rhythm"`
	PregnancyStatus        *bool       `json:"pregnancy_status,omitempty"`
	Triage                 null.String `json:"triage"`
}

type GCSResponse struct {
	Eye    *int16 `json:"eye,omitempty"`
	Verbal *int16 `json:"verbal,omitempty"`
	Motor  *int16 `json:"motor,omitempty"`
	Total  *int16 `json:"total,omitempty"`
}

type PainAssessmentResponse struct {
	HasPain  *bool  `json:"has_pain,omitempty"`
	Trigger  string `json:"trigger,omitempty"`
	Quality  string `json:"quality,omitempty"`
	Location string `json:"location,omitempty"`
	Scale    *int16 `json:"scale,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// UpsertAnamnesaRequest is the API payload for upsert.
// doctor_id and nurse_id are mutually required: if either is provided, both must be.
// fall_risk and lifestyle are accepted for completeness but are NOT persisted to the DB.
type UpsertAnamnesaRequest struct {
	DoctorID           string               `json:"doctor_id" validate:"required_without=NurseID,omitempty,uuid4"`
	NurseID            string               `json:"nurse_id" validate:"required_without=DoctorID,omitempty,uuid4"`
	ChiefComplaint     string               `json:"chief_complaint" validate:"required,min=1,max=2000"`
	SecondaryComplaint null.String          `json:"secondary_complaint" validate:"omitempty,max=2000"`
	HistoryOfIllness   null.String          `json:"history_of_illness" validate:"omitempty,max=2000"`
	IllnessDuration    IllnessDurationInput `json:"illness_duration"`
	VitalSigns         VitalSignsInput      `json:"vital_signs"`
	GCS                GCSInput             `json:"gcs"`
	PainAssessment     PainAssessmentInput  `json:"pain_assessment"`
	FallRisk           FallRiskInput        `json:"fall_risk"`
	Lifestyle          LifestyleInput       `json:"lifestyle"`
}

type IllnessDurationInput struct {
	Years  int16 `json:"years" validate:"min=0,max=200"`
	Months int16 `json:"months" validate:"min=0,max=11"`
	Days   int16 `json:"days" validate:"min=0,max=31"`
}

type VitalSignsInput struct {
	Systolic         null.Int16   `json:"systolic" validate:"omitempty,min=0,max=300"`
	Diastolic        null.Int16   `json:"diastolic" validate:"omitempty,min=0,max=200"`
	Pulse            null.Int16   `json:"pulse" validate:"omitempty,min=0,max=300"`
	Temperature      null.Float32 `json:"temperature" validate:"omitempty,gte=30,lte=45"`
	RespiratoryRate  null.Int16   `json:"respiratory_rate" validate:"omitempty,min=0,max=100"`
	OxygenSaturation null.Int16   `json:"oxygen_saturation" validate:"omitempty,min=0,max=100"`
	Weight           null.Float32 `json:"weight" validate:"omitempty,gte=0.5,lte=500"`
	Height           null.Float32 `json:"height" validate:"omitempty,gte=20,lte=300"`
	// Enum values with spaces (e.g. "COMPOS MENTIS") cannot use oneof tag; validated by DB constraint.
	HeightMeasurement      null.String  `json:"height_measurement" validate:"omitempty,oneof=berdiri telentang"`
	AbdominalCircumference null.Float32 `json:"abdominal_circumference" validate:"omitempty,gte=0,lte=300"`
	Consciousness          null.String  `json:"consciousness" validate:"omitempty"`
	HeartRhythm            null.String  `json:"heart_rhythm" validate:"omitempty,oneof=regular irregular"`
	PregnancyStatus        null.Bool    `json:"pregnancy_status"`
	Triage                 null.String  `json:"triage" validate:"omitempty"`
}

type GCSInput struct {
	Eye    null.Int16 `json:"eye" validate:"omitempty,min=1,max=4"`
	Verbal null.Int16 `json:"verbal" validate:"omitempty,min=1,max=5"`
	Motor  null.Int16 `json:"motor" validate:"omitempty,min=1,max=6"`
}

type PainAssessmentInput struct {
	HasPain  null.Bool   `json:"has_pain"`
	Trigger  null.String `json:"trigger" validate:"omitempty,max=500"`
	Quality  null.String `json:"quality" validate:"omitempty,oneof=tekanan terbakar melilit tertusuk diiris mencengkram"`
	Location null.String `json:"location" validate:"omitempty,max=500"`
	Scale    null.Int16  `json:"scale" validate:"omitempty,min=0,max=10"`
	Pattern  null.String `json:"pattern" validate:"omitempty,oneof=intermittent continuous"`
}

// FallRisk and Lifestyle are captured in the request but intentionally not stored to DB.

type FallRiskInput struct {
	Gait    bool `json:"gait"`
	Support bool `json:"support"`
}

type LifestyleInput struct {
	Smoking           bool `json:"smoking"`
	Alcohol           bool `json:"alcohol"`
	LowVegetableFruit bool `json:"low_vegetable_fruit"`
}

type UpsertAnamnesaResponse struct {
	ID string `json:"id"`
}
