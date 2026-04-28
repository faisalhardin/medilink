package model

import (
	"time"

	"github.com/volatiletech/null/v8"
)

const (
	TRX_ANAMNESA_TABLE = "mdl_trx_anamnesa"
)

// TrxAnamnesa holds the subjective note, vital signs, and GCS for a visit.
// VSMAP, VSBMI, VSBMIResult are computed by the usecase and stored here;
// GCSTotal is a DB generated column (read-only via xorm:"<-").
type TrxAnamnesa struct {
	ID                 string      `xorm:"'id' pk" json:"id"`
	VisitID            int64       `xorm:"'visit_id'" json:"visit_id"`
	InstitutionID      int64       `xorm:"'institution_id'" json:"-"`
	NurseID            *string     `xorm:"'nurse_id' null" json:"nurse_id,omitempty"`
	ChiefComplaint     null.String `xorm:"'chief_complaint' null" json:"chief_complaint"`
	HistoryOfIllness   null.String `xorm:"'history_of_illness' null" json:"history_of_illness"`
	VSSystolic         *int16      `xorm:"'vs_systolic' null" json:"vs_systolic,omitempty"`
	VSDiastolic        *int16      `xorm:"'vs_diastolic' null" json:"vs_diastolic,omitempty"`
	VSPulse            *int16      `xorm:"'vs_pulse' null" json:"vs_pulse,omitempty"`
	VSTemperature      *float32    `xorm:"'vs_temperature' null" json:"vs_temperature,omitempty"`
	VSRespiratoryRate  *int16      `xorm:"'vs_respiratory_rate' null" json:"vs_respiratory_rate,omitempty"`
	VSOxygenSaturation *int16      `xorm:"'vs_oxygen_saturation' null" json:"vs_oxygen_saturation,omitempty"`
	VSMAP              *int16      `xorm:"'vs_map' null" json:"vs_map,omitempty"`
	VSWeight           *float32    `xorm:"'vs_weight' null" json:"vs_weight,omitempty"`
	VSHeight           *float32    `xorm:"'vs_height' null" json:"vs_height,omitempty"`
	VSBMI              *float32    `xorm:"'vs_bmi' null" json:"vs_bmi,omitempty"`
	VSBMIResult        null.String `xorm:"'vs_bmi_result' null" json:"vs_bmi_result"`
	GCSEye             *int16      `xorm:"'gcs_eye' null" json:"gcs_eye,omitempty"`
	GCSVerbal          *int16      `xorm:"'gcs_verbal' null" json:"gcs_verbal,omitempty"`
	GCSMotor           *int16      `xorm:"'gcs_motor' null" json:"gcs_motor,omitempty"`
	// GCSTotal is read-only — generated column in Postgres
	GCSTotal  *int16    `xorm:"'gcs_total' <-" json:"gcs_total,omitempty"`
	CreatedAt time.Time `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt time.Time `xorm:"'updated_at' updated" json:"updated_at"`
}

// UpsertAnamnesaRequest is the minimal API contract aligned with
// mdl_trx_anamnesa columns currently present in migrations.
type UpsertAnamnesaRequest struct {
	NurseID          string          `json:"nurse_id" validate:"required,uuid4"`
	ChiefComplaint   string          `json:"chief_complaint" validate:"required,min=1,max=2000"`
	HistoryOfIllness string          `json:"history_of_illness" validate:"omitempty,max=2000"`
	VitalSigns       VitalSignsInput `json:"vital_signs"`
	GCS              GCSInput        `json:"gcs"`
}

type VitalSignsInput struct {
	Systolic         *int16   `json:"systolic" validate:"omitempty,min=0,max=300"`
	Diastolic        *int16   `json:"diastolic" validate:"omitempty,min=0,max=200"`
	Pulse            *int16   `json:"pulse" validate:"omitempty,min=0,max=300"`
	Temperature      *float32 `json:"temperature" validate:"omitempty,gte=30,lte=45"`
	RespiratoryRate  *int16   `json:"respiratory_rate" validate:"omitempty,min=0,max=100"`
	OxygenSaturation *int16   `json:"oxygen_saturation" validate:"omitempty,min=0,max=100"`
	Weight           *float32 `json:"weight" validate:"omitempty,gte=0.5,lte=500"`
	Height           *float32 `json:"height" validate:"omitempty,gte=20,lte=300"`
}

type GCSInput struct {
	Eye    *int16 `json:"eye" validate:"omitempty,min=1,max=4"`
	Verbal *int16 `json:"verbal" validate:"omitempty,min=1,max=5"`
	Motor  *int16 `json:"motor" validate:"omitempty,min=1,max=6"`
}

type UpsertAnamnesaResponse struct {
	ID string `json:"id"`
}
