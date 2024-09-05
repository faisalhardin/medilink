package model

import (
	"time"
)

const (
	MST_PATIENT_INSTITUTION = "mdl_mst_patient_institution"
	MST_PATIENT_VISIT       = "mdl_mst_patient_visit"
)

type MstPatientInstitution struct {
	ID            int64      `json:"-" xorm:"'id' pk autoincr"`
	UUID          string     `json:"uuid" xorm:"'uuid' <-"`
	NIK           string     `json:"nik" xorm:"'nik'"`
	Name          string     `json:"name" xorm:"'name'"`
	Sex           string     `json:"sex" xorm:"'sex'"`
	PlaceOfBirth  string     `json:"place_of_birth" xorm:"'place_of_birth'"`
	DateOfBirth   time.Time  `json:"date_of_birth" xorm:"'date_of_birth'"`
	Address       string     `json:"address" xorm:"'address'"`
	Religion      string     `json:"religion" xorm:"'religion'"`
	InstitutionID int64      `json:"institution_id" xorm:"'id_mst_institution'"`
	CreateTime    time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime    time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime    *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type MstPatientVisit struct {
	ID         int64      `json:"-" xorm:"'id' pk autoincr"`
	UUID       string     `json:"uuid" xorm:"'uuid' uuid_generate_v4()"`
	PatientID  int64      `json:"patient_id" xorm:"'id_mst_patient'"`
	Action     string     `json:"action" xorm:"'action'"`
	Status     string     `json:"status" xorm:"'status'"`
	Notes      string     `json:"notes" xorm:"'notes'"`
	CreateTime time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type RegisterNewPatientRequest struct {
	NIK          string    `json:"nik"`
	Name         string    `json:"name" validate:"required"`
	Sex          string    `json:"sex" validate:"required,oneof=male female"`
	PlaceOfBirth string    `json:"place_of_birth"`
	DateOfBirth  time.Time `json:"date_of_birth" validate:"required"`
	Address      string    `json:"address"`
	Religion     string    `json:"religion"`
}

type GetPatientParams struct {
	PatientUUIDs  []string  `scheme:"patient_ids"`
	DateOfBirth   time.Time `scheme:"date_of_birth"`
	InstitutionID int64     `scheme:"institution_id"`
}

type GetPatientResponse struct {
	UUID         string    `json:"uuid" xorm:"'uuid' <-"`
	NIK          string    `json:"nik" xorm:"'nik'"`
	Name         string    `json:"name" xorm:"'name'"`
	PlaceOfBirth string    `json:"place_of_birth" xorm:"'place_of_birth'"`
	DateOfBirth  time.Time `json:"date_of_birth" xorm:"'date_of_birth'"`
	Address      string    `json:"address" xorm:"'address'"`
	Religion     string    `json:"religion" xorm:"'religion'"`
}
