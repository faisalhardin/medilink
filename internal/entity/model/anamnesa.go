package model

import "time"

const (
	TRX_ANAMNESA_TABLE = "trx_anamnesa"
)

// TrxAnamnesa holds the subjective note, vital signs, and GCS for a visit.
// VSMAP, VSBMI, VSBMIResult are computed by the usecase and stored here;
// GCSTotal is a DB generated column (read-only via xorm:"<-").
type TrxAnamnesa struct {
	ID                 string    `xorm:"'id' pk" json:"id"`
	VisitID            int64     `xorm:"'visit_id'" json:"visit_id"`
	NurseID            *string   `xorm:"'nurse_id' null" json:"nurse_id,omitempty"`
	// Subjective
	ChiefComplaint    string    `xorm:"'chief_complaint'" json:"chief_complaint,omitempty"`
	HistoryOfIllness  string    `xorm:"'history_of_illness'" json:"history_of_illness,omitempty"`
	// Vital signs
	VSSystolic        *int16    `xorm:"'vs_systolic' null" json:"vs_systolic,omitempty"`
	VSDiastolic       *int16    `xorm:"'vs_diastolic' null" json:"vs_diastolic,omitempty"`
	VSPulse           *int16    `xorm:"'vs_pulse' null" json:"vs_pulse,omitempty"`
	VSTemperature     *float32  `xorm:"'vs_temperature' null" json:"vs_temperature,omitempty"`
	VSRespiratoryRate *int16    `xorm:"'vs_respiratory_rate' null" json:"vs_respiratory_rate,omitempty"`
	VSOxygenSaturation *int16   `xorm:"'vs_oxygen_saturation' null" json:"vs_oxygen_saturation,omitempty"`
	// Derived (computed in usecase)
	VSMAP             *int16    `xorm:"'vs_map' null" json:"vs_map,omitempty"`
	VSWeight          *float32  `xorm:"'vs_weight' null" json:"vs_weight,omitempty"`
	VSHeight          *float32  `xorm:"'vs_height' null" json:"vs_height,omitempty"`
	VSBMI             *float32  `xorm:"'vs_bmi' null" json:"vs_bmi,omitempty"`
	VSBMIResult       string    `xorm:"'vs_bmi_result'" json:"vs_bmi_result,omitempty"`
	// GCS
	GCSEye            *int16    `xorm:"'gcs_eye' null" json:"gcs_eye,omitempty"`
	GCSVerbal         *int16    `xorm:"'gcs_verbal' null" json:"gcs_verbal,omitempty"`
	GCSMotor          *int16    `xorm:"'gcs_motor' null" json:"gcs_motor,omitempty"`
	// GCSTotal is read-only — generated column in Postgres
	GCSTotal          *int16    `xorm:"'gcs_total' <-" json:"gcs_total,omitempty"`
	// Timestamps
	CreatedAt         time.Time `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt         time.Time `xorm:"'updated_at' updated" json:"updated_at"`
}
