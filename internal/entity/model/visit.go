package model

import "time"

type TrxPatientVisit struct {
	ID               int64      `xorm:"'id' pk autoincr" json:"id"`
	IDMstPatient     int64      `xorm:"'id_mst_patient'" json:"-"`
	IDMstInstitution int64      `xorm:"'id_mst_institution'" json:"-"`
	Action           string     `xorm:"'action'" json:"action"`
	Status           string     `xorm:"'status'" json:"status"`
	Notes            string     `xorm:"'notes'" json:"notes"`
	CreateTime       time.Time  `json:"create_time" xorm:"'create_time' created"`
	UpdateTime       time.Time  `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime       *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type DtlPatientVisit struct {
	ID                 int64      `xorm:"'id' pk autoincr" json:"id"`
	IDTrxPatientVisit  int64      `xorm:"'id_trx_patient_visit'" json:"id_dtl_patient_visit"`
	TouchpointName     string     `xorm:"'touchpoint_name'" json:"touchpoint_name"`
	TouchpointCategory string     `xorm:"'touchpoint_category'" json:"touchpoint_category"`
	Notes              string     `xorm:"'notes'" json:"notes"`
	CreateTime         time.Time  `json:"create_time" xorm:"'create_time' created"`
	UpdateTime         time.Time  `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime         *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type InsertNewVisitRequest struct {
	PatientUUID string `json:"patient_uuid"`
	TrxPatientVisit
}

type UpdatePatientVisitRequest struct {
	InsertNewVisitRequest
}

type GetPatientVisitParams struct {
	PatientID        int64  `xorm:"id"`
	PatientUUID      string `xorm:"uuid"`
	IDPatientVisit   int64  `xorm:"id_trx_patient_visit"`
	IDMstInstitution int64  `xorm:"id_mst_institution"`
}

type DtlPatientVisitRequest struct {
	ID                 int64  `json:"id" schema:"id"`
	IDTrxPatientVisit  int64  `json:"id_dtl_patient_visit" schema:"id_dtl_patient_visit"`
	TouchpointName     string `json:"touchpoint_name"`
	TouchpointCategory string `json:"touchpoint_category"`
	Notes              string `json:"notes"`
}
