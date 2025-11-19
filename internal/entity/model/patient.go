package model

import (
	"time"
)

const (
	MstPatientInstitutionTableName = "mdl_mst_patient_institution"
	TrxPatientVisitTableName       = "mdl_trx_patient_visit"
	DtlPatientVisitTableName       = "mdl_dtl_patient_visit"
	TrxVisitProductTableName       = "mdl_trx_visit_product"
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
	PhoneNumber   string     `json:"phone_number" xorm:"phone_number"`
	InstitutionID int64      `json:"institution_id" xorm:"'id_mst_institution'"`
	CreateTime    time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime    time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime    *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

// type MstPatientVisit struct {
// 	ID         int64      `json:"-" xorm:"'id' pk autoincr"`
// 	UUID       string     `json:"uuid" xorm:"'uuid' uuid_generate_v4()"`
// 	PatientID  int64      `json:"patient_id" xorm:"'id_mst_patient'"`
// 	Action     string     `json:"action" xorm:"'action'"`
// 	Status     string     `json:"status" xorm:"'status'"`
// 	Notes      string     `json:"notes" xorm:"'notes'"`
// 	CreateTime time.Time  `json:"-" xorm:"'create_time' created"`
// 	UpdateTime time.Time  `json:"-" xorm:"'update_time' updated"`
// 	DeleteTime *time.Time `json:"-" xorm:"'delete_time' deleted"`
// }

type RegisterNewPatientRequest struct {
	NIK          string `json:"nik"`
	Name         string `json:"name" validate:"required"`
	Sex          string `json:"sex" validate:"required,oneof=male female"`
	PlaceOfBirth string `json:"place_of_birth"`
	DateOfBirth  Time   `json:"date_of_birth" validate:"required"`
	Address      string `json:"address"`
	Religion     string `json:"religion"`
}

type GetPatientParams struct {
	PatientUUIDs  []string `schema:"patient_ids"`
	DateOfBirth   Time     `schema:"date_of_birth"`
	Name          string   `schema:"name"`
	InstitutionID int64    `schema:"institution_id"`
	NIK           string   `schema:"nik"`
	PhoneNumber   string   `schema:"phone_number"`
	CommonRequestPayload
}

type GetPatientResponse struct {
	UUID         string    `json:"uuid" xorm:"'uuid' <-"`
	NIK          string    `json:"nik" xorm:"'nik'"`
	Name         string    `json:"name" xorm:"'name'"`
	PlaceOfBirth string    `json:"place_of_birth" xorm:"'place_of_birth'"`
	DateOfBirth  time.Time `json:"date_of_birth" xorm:"'date_of_birth'"`
	Address      string    `json:"address" xorm:"'address'"`
	Religion     string    `json:"religion" xorm:"'religion'"`
	PhoneNumber  string    `json:"phone_number"`
	Sex          string    `json:"sex" xorm:"'sex'"`
}

type UpdatePatientRequest struct {
	UUID         string `json:"uuid" xorm:"'uuid' <-"`
	NIK          string `json:"nik" xorm:"'nik'"`
	Name         string `json:"name" xorm:"'name'"`
	Sex          string `json:"sex" xorm:"'sex'"`
	PlaceOfBirth string `json:"place_of_birth" xorm:"'place_of_birth'"`
	DateOfBirth  Time   `json:"date_of_birth" xorm:"'date_of_birth'"`
	Address      string `json:"address" xorm:"'address'"`
	Religion     string `json:"religion" xorm:"'religion'"`
}
