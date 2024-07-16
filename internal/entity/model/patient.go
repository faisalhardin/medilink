package model

import "time"

const (
	MST_PATIENT_INSTITUTION = "mdl_mst_patient_institution"
)

type MstPatientInstitution struct {
	ID            int64      `json:"-" xorm:"'id' pk autoincr"`
	UUID          string     `json:"uuid" xorm:"'uuid'"`
	NIK           string     `json:"nik" xorm:"'nik'"`
	Name          string     `json:"name" xorm:"'name'"`
	PlaceOfBirth  string     `json:"place_of_birth" xorm:"'place_of_birth'"`
	DateOfBirth   time.Time  `json:"date_of_birth" xorm:"'date_of_birth'"`
	Address       string     `json:"address" xorm:"'address'"`
	InstitutionID int64      `json:"institution_id" xorm:"'id_mst_institution'"`
	CreateTime    time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime    time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime    *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type MstPatientVisit struct {
	ID         int64      `json:"-" xorm:"'id' pk autoincr"`
	UUID       string     `json:"uuid" xorm:"'uuid'"`
	PatientID  int64      `json:"patient_id" xorm:"'id_mst_patient'"`
	Action     string     `json:"action" xorm:"'action'"`
	Status     string     `json:"status" xorm:"'status'"`
	Notes      string     `json:"notes" xorm:"'notes'"`
	CreateTime time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime *time.Time `json:"-" xorm:"'delete_time' deleted"`
}
