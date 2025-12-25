package model

import (
	"encoding/json"
)

// HstOdontogram represents an odontogram event in the history table
type HstOdontogram struct {
	EventID             string          `xorm:"'event_id' pk" json:"event_id"`
	InstitutionID       int64           `xorm:"'institution_id'" json:"-"`
	PatientID           int64           `xorm:"'patient_id' notnull" json:"patient_id"`
	VisitID             int64           `xorm:"'visit_id' notnull" json:"visit_id"`
	JourneyPointShortID string          `xorm:"'journey_point_short_id'" json:"journey_point_id"`
	EventType           string          `xorm:"'event_type' notnull" json:"event_type"`
	ToothID             string          `xorm:"'tooth_id' notnull" json:"tooth_id"`
	SequenceNumber      int64           `xorm:"'sequence_number' notnull" json:"sequence_number"`
	EventData           json.RawMessage `xorm:"'event_data' jsonb notnull" json:"event_data"`
	LogicalTimestamp    int64           `xorm:"'logical_timestamp' notnull" json:"logical_timestamp"`
	CreatedByStaffID    int64           `xorm:"'created_by_staff_id' notnull" json:"created_by_staff_id"`
	UnixTimestamp       int64           `xorm:"'unix_timestamp' notnull" json:"unix_timestamp"`
	CreatedBy           string          `xorm:"'created_by' notnull" json:"created_by"`
	CreateTime          int64           `xorm:"'create_time' notnull" json:"create_time"`
}

// TableName returns the table name for XORM
func (HstOdontogram) TableName() string {
	return "mdl_hst_odontogram"
}

// MstPatientOdontogram represents the snapshot of a patient's odontogram
type MstPatientOdontogram struct {
	ID                  string          `xorm:"'id' pk" json:"id"`
	InstitutionID       int64           `xorm:"'institution_id' notnull" json:"-"`
	PatientID           int64           `xorm:"'patient_id' notnull unique" json:"-"`
	Snapshot            json.RawMessage `xorm:"'snapshot' jsonb notnull" json:"snapshot"`
	LastEventSequence   int64           `xorm:"'last_event_sequence' notnull" json:"last_event_sequence"`
	MaxLogicalTimestamp int64           `xorm:"'max_logical_timestamp' notnull" json:"max_logical_timestamp"`
	LastUpdated         int64           `xorm:"'last_updated' notnull" json:"last_updated"`
}

// TableName returns the table name for XORM
func (MstPatientOdontogram) TableName() string {
	return "mdl_mst_patient_odontogram"
}

// OdontogramEventData represents the event_data field structure
type OdontogramEventData struct {
	WholeToothCode []string `json:"whole_tooth_code,omitempty"`
	GeneralNotes   string   `json:"general_notes,omitempty"`
	Surface        string   `json:"surface,omitempty"`
	SurfaceCode    string   `json:"surface_code,omitempty"`
	SurfaceNotes   string   `json:"surface_notes,omitempty"`
}

// CreateOdontogramEventRequest represents a request to create odontogram events
type CreateOdontogramEventRequest struct {
	EventID        string              `json:"event_id,omitempty"`
	VisitID        int64               `json:"visit_id" schema:"visit_id"`
	JourneyPointID string              `json:"journey_point_id" schema:"journey_point_id"`
	EventType      string              `json:"event_type" schema:"event_type"`
	ToothID        string              `json:"tooth_id" schema:"tooth_id"`
	PatientUUID    string              `json:"patient_uuid" schema:"patient_uuid"`
	EventData      OdontogramEventData `json:"event_data"`
	UnixTimestamp  int64               `json:"unix_timestamp,omitempty"`
}

// BulkCreateOdontogramEventRequest represents batch event creation
type BulkCreateOdontogramEventRequest struct {
	Events []CreateOdontogramEventRequest `json:"events"`
}

// CreateOdontogramEventResponse represents the response for a single event creation
type CreateOdontogramEventResponse struct {
	EventID          string `json:"event_id"`
	SequenceNumber   int64  `json:"sequence_number"`
	LogicalTimestamp int64  `json:"logical_timestamp"`
	Status           string `json:"status"`
}

// CreateOdontogramEventsResponse represents batch event creation response
type CreateOdontogramEventsResponse struct {
	Results             []CreateOdontogramEventResponse `json:"results"`
	MaxLogicalTimestamp int64                           `json:"max_logical_timestamp"`
	MaxSequenceNumber   int64                           `json:"max_sequence_number"`
}

// GetOdontogramEventsParams represents query parameters for getting events
type GetOdontogramEventsParams struct {
	InstitutionID int64  `schema:"institution_id"`
	PatientUUID   string `schema:"patient_uuid" validate:"required"`
	PatientID     int64
	EventID       string `schema:"event_id"`
	ToothID       string `schema:"tooth_id"`
	EventType     string `schema:"event_type"`
	VisitID       int64  `schema:"visit_id"`
	FromSequence  int64  `schema:"from_sequence"`
	ToSequence    int64  `schema:"to_sequence"`
	CommonRequestPayload
}

// GetOdontogramSnapshotParams represents query parameters for getting snapshot
type GetOdontogramSnapshotParams struct {
	PatientUUID    string `schema:"patient_uuid" validate:"required"`
	SequenceNumber int64  `schema:"sequence_number"`
	VisitID        int64  `schema:"visit_id"`
}

// OdontogramSnapshot represents the EditorJS format snapshot
type OdontogramSnapshot struct {
	Teeth map[string]ToothData `json:"teeth"`
}

// ToothData represents data for a single tooth
type ToothData struct {
	ID             string        `json:"id"`
	Surfaces       []SurfaceData `json:"surfaces"`
	WholeToothCode []string      `json:"wholeToothCode"`
	GeneralNotes   string        `json:"generalNotes"`
}

// SurfaceData represents data for a tooth surface
type SurfaceData struct {
	Surface   string `json:"surface"`
	Code      string `json:"code"`
	Condition string `json:"condition,omitempty"`
	Color     string `json:"color,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	Notes     string `json:"notes,omitempty"`
	// CRDT metadata for conflict resolution
	LogicalTimestamp int64 `json:"-"` // Not serialized to client
	CreatedByStaffID int64 `json:"-"` // Not serialized to client
}

// GetOdontogramSnapshotResponse represents the response for snapshot retrieval
type GetOdontogramSnapshotResponse struct {
	Snapshot            OdontogramSnapshot `json:"snapshot"`
	MaxLogicalTimestamp int64              `json:"max_logical_timestamp"`
	MaxSequenceNumber   int64              `json:"max_sequence_number"`
	LastUpdated         int64              `json:"last_updated"`
}

// GetOdontogramEventsResponse represents the response for events retrieval
type GetOdontogramEventsResponse struct {
	Events              []HstOdontogramResponse `json:"events"`
	MaxLogicalTimestamp int64                   `json:"max_logical_timestamp"`
	MaxSequenceNumber   int64                   `json:"max_sequence_number"`
	Total               int                     `json:"total"`
}

// HstOdontogramResponse represents a single event in the response
type HstOdontogramResponse struct {
	EventID          string              `json:"event_id"`
	VisitID          int64               `json:"visit_id"`
	JourneyPointID   string              `json:"journey_point_id"`
	EventType        string              `json:"event_type"`
	ToothID          string              `json:"tooth_id"`
	PatientUUID      string              `json:"patient_uuid"`
	EventData        OdontogramEventData `json:"event_data"`
	SequenceNumber   int64               `json:"sequence_number"`
	LogicalTimestamp int64               `json:"logical_timestamp"`
	UnixTimestamp    int64               `json:"unix_timestamp"`
	CreatedBy        string              `json:"created_by"`
}

type GetEventsByPatientParams struct {
	PatientID     int64 `schema:"patient_id"`
	InstitutionID int64 `schema:"institution_id"`
	FromSequence  int64 `schema:"from_sequence"`
	ToSequence    int64 `schema:"to_sequence"`
	VisitID       int64 `schema:"visit_id"`
	Limit         int   `schema:"limit"`
	Offset        int   `schema:"offset"`
}
