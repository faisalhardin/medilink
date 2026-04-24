package model

import (
	"time"

	"github.com/volatiletech/null/v8"
)

const (
	REF_ICD10_TABLE  = "ref_icd10"
	MST_DOCTOR_TABLE = "mdl_mst_doctor"
	MST_NURSE_TABLE  = "mdl_mst_nurse"
)

// RefICD10 is the ICD-10 diagnosis code reference table.
type RefICD10 struct {
	Code      string      `xorm:"'code' pk" json:"code"`
	Display   string      `xorm:"'display'" json:"display"`
	Category  null.String `xorm:"'category' null" json:"category"`
	CreatedAt time.Time   `xorm:"'created_at' created" json:"-"`
}

// MstDoctor is the clinical-identity record for a physician.
// StaffUUID is an optional link to an mdl_mst_staff login account (NULL for external practitioners).
type MstDoctor struct {
	ID             string      `xorm:"'id' pk" json:"id"`
	StaffUUID      null.String `xorm:"'staff_uuid' null" json:"staff_uuid"`
	Name           string      `xorm:"'name'" json:"name"`
	SIPNumber      null.String `xorm:"'sip_number' null" json:"sip_number"`
	Specialization null.String `xorm:"'specialization' null" json:"specialization"`
	InstitutionID  int64       `xorm:"'institution_id'" json:"-"`
	Active         bool        `xorm:"'active'" json:"active"`
	CreatedAt      time.Time   `xorm:"'created_at' created" json:"-"`
	UpdatedAt      time.Time   `xorm:"'updated_at' updated" json:"-"`
}

// MstNurse is the clinical-identity record for a nurse, midwife, or paramedic.
// StaffUUID is an optional link to an mdl_mst_staff login account.
type MstNurse struct {
	ID            string      `xorm:"'id' pk" json:"id"`
	StaffUUID     null.String `xorm:"'staff_uuid' null" json:"staff_uuid"`
	Name          string      `xorm:"'name'" json:"name"`
	SIPNumber     null.String `xorm:"'sip_number' null" json:"sip_number"`
	Role          string      `xorm:"'role'" json:"role"`
	InstitutionID int64       `xorm:"'institution_id'" json:"-"`
	Active        bool        `xorm:"'active'" json:"active"`
	CreatedAt     time.Time   `xorm:"'created_at' created" json:"-"`
	UpdatedAt     time.Time   `xorm:"'updated_at' updated" json:"-"`
}

// DoctorSearchResult is used for the GET /v1/doctor/search response.
// Fields that the DB may return as NULL (sip_number, specialization, staff_uuid)
// are null.String so JSON distinguishes "not set" from "empty string".
type DoctorSearchResult struct {
	ID             string      `xorm:"'id'" json:"id"`
	Name           string      `xorm:"'name'" json:"name"`
	SIPNumber      null.String `xorm:"'sip_number' null" json:"sip_number"`
	Specialization null.String `xorm:"'specialization' null" json:"specialization"`
	StaffUUID      null.String `xorm:"'staff_uuid' null" json:"staff_uuid"`
}

// NurseSearchResult is used for the GET /v1/nurse/search response.
type NurseSearchResult struct {
	ID        string      `xorm:"'id'" json:"id"`
	Name      string      `xorm:"'name'" json:"name"`
	SIPNumber null.String `xorm:"'sip_number' null" json:"sip_number"`
	Role      string      `xorm:"'role'" json:"role"`
	StaffUUID null.String `xorm:"'staff_uuid' null" json:"staff_uuid"`
}
