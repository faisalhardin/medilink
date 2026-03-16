package model

import "time"

const (
	TrxRecallTableName = "mdl_trx_recall"
)

// RecallType represents the kind of scheduled follow-up
const (
	RecallTypeControl     = "control"
	RecallTypeAppointment = "appointment"
	RecallTypeOther       = "other"
)

// TrxRecall stores a scheduled control or future appointment for a patient (doctor reminder)
type TrxRecall struct {
	ID                  int64      `xorm:"'id' pk autoincr" json:"id"`
	IDMstPatient        int64      `xorm:"'id_mst_patient' notnull" json:"-"`
	IDMstInstitution    int64      `xorm:"'id_mst_institution' notnull" json:"-"`
	ScheduledAt         time.Time  `xorm:"'scheduled_at' notnull" json:"scheduled_at"`
	RecallType          string     `xorm:"'recall_type' notnull" json:"recall_type"` // control, appointment, other
	Notes               string     `xorm:"'notes'" json:"notes"`
	CreatedByIDMstStaff int64      `xorm:"'created_by_id_mst_staff' notnull" json:"-"`
	IDTrxPatientVisit   int64      `xorm:"'id_trx_patient_visit'" json:"-"` // optional link to visit that created this recall
	CreateTime          time.Time  `xorm:"'create_time' created" json:"create_time"`
	UpdateTime          time.Time  `xorm:"'update_time' updated" json:"update_time"`
	DeleteTime          *time.Time `xorm:"'delete_time' deleted" json:"-"`
}

type TrxRecallJoinPatient struct {
	TrxRecall             `xorm:"extends"`
	MstPatientInstitution `xorm:"extends"`
}

// TableName returns the table name for XORM
func (TrxRecall) TableName() string {
	return TrxRecallTableName
}

// CreateRecallRequest is the request to schedule a recall
type CreateRecallRequest struct {
	PatientUUID       string `json:"patient_uuid" validate:"required"`
	ScheduledAt       Time   `json:"scheduled_at" validate:"required"`
	RecallType        string `json:"recall_type" validate:"required,oneof=control appointment other"`
	Notes             string `json:"notes"`
	IDTrxPatientVisit int64  `json:"id_trx_patient_visit"` // optional
}

// UpdateRecallRequest is the request to patch a recall (only scheduled_at, recall_type, notes)
type UpdateRecallRequest struct {
	IDMstRecall int64   `json:"id" validate:"required"`
	ScheduledAt *Time   `json:"scheduled_at,omitempty"`
	RecallType  *string `json:"recall_type,omitempty" validate:"omitempty,oneof=control appointment other"`
	Notes       *string `json:"notes,omitempty"`
}

// GetRecallParams for listing/filtering recalls
type GetRecallParams struct {
	PatientUUID  string `schema:"patient_uuid"`
	IDMstPatient int64  `schema:"-"`
	FromTime     Time   `schema:"from_time"`
	ToTime       Time   `schema:"to_time"`
	RecallType   string `schema:"recall_type"`
	CommonRequestPayload
}

// RecallResponse is a single recall with optional patient info for listing
type RecallResponse struct {
	ID                int64     `json:"id"`
	PatientUUID       string    `json:"patient_uuid"`
	PatientName       string    `json:"patient_name,omitempty"`
	ScheduledAt       time.Time `json:"scheduled_at"`
	RecallType        string    `json:"recall_type"`
	Notes             string    `json:"notes"`
	IDTrxPatientVisit int64     `json:"id_trx_patient_visit,omitempty"`
	CreateTime        time.Time `json:"create_time"`
}

// NextRecallResponse is the next upcoming recall for a patient (for doctor reminder)
type NextRecallResponse struct {
	RecallResponse
	HasNext bool `json:"has_next"`
}
