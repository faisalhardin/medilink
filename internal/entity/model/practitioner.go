package model

import "time"

const (
	REF_ICD10_TABLE  = "ref_icd10"
	MST_DOCTOR_TABLE = "mdl_mst_doctor"
	MST_NURSE_TABLE  = "mdl_mst_nurse"
)

// RefICD10 is the ICD-10 diagnosis code reference table.
type RefICD10 struct {
	Code      string    `xorm:"'code' pk" json:"code"`
	Display   string    `xorm:"'display'" json:"display"`
	Category  string    `xorm:"'category'" json:"category,omitempty"`
	CreatedAt time.Time `xorm:"'created_at' created" json:"-"`
}

// MstDoctor is the clinical-identity record for a physician.
// StaffUUID is an optional link to an mdl_mst_staff login account (NULL for external practitioners).
type MstDoctor struct {
	ID             string    `xorm:"'id' pk" json:"id"`
	StaffUUID      *string   `xorm:"'staff_uuid' null" json:"staff_uuid,omitempty"`
	Name           string    `xorm:"'name'" json:"name"`
	SIPNumber      string    `xorm:"'sip_number'" json:"sip_number,omitempty"`
	Specialization string    `xorm:"'specialization'" json:"specialization,omitempty"`
	InstitutionID  int64     `xorm:"'institution_id'" json:"-"`
	Active         bool      `xorm:"'active'" json:"active"`
	CreatedAt      time.Time `xorm:"'created_at' created" json:"-"`
	UpdatedAt      time.Time `xorm:"'updated_at' updated" json:"-"`
}

// MstNurse is the clinical-identity record for a nurse, midwife, or paramedic.
// StaffUUID is an optional link to an mdl_mst_staff login account.
type MstNurse struct {
	ID            string    `xorm:"'id' pk" json:"id"`
	StaffUUID     *string   `xorm:"'staff_uuid' null" json:"staff_uuid,omitempty"`
	Name          string    `xorm:"'name'" json:"name"`
	SIPNumber     string    `xorm:"'sip_number'" json:"sip_number,omitempty"`
	Role          string    `xorm:"'role'" json:"role"`
	InstitutionID int64     `xorm:"'institution_id'" json:"-"`
	Active        bool      `xorm:"'active'" json:"active"`
	CreatedAt     time.Time `xorm:"'created_at' created" json:"-"`
	UpdatedAt     time.Time `xorm:"'updated_at' updated" json:"-"`
}

// DoctorSearchResult is used for the GET /v1/doctor/search response.
type DoctorSearchResult struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	SIPNumber      string  `json:"sip_number,omitempty"`
	Specialization string  `json:"specialization,omitempty"`
	StaffUUID      *string `json:"staff_uuid,omitempty"`
}

// NurseSearchResult is used for the GET /v1/nurse/search response.
type NurseSearchResult struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	SIPNumber string  `json:"sip_number,omitempty"`
	Role      string  `json:"role"`
	StaffUUID *string `json:"staff_uuid,omitempty"`
}
