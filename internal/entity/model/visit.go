package model

import "time"

type TrxPatientVisit struct {
	ID           int64      `xorm:"'id' pk autoincr" json:"id"`
	IDMstPatient int64      `xorm:"'id_mst_patient'" json:"-"`
	Action       string     `xorm:"'action'" json:"action"`
	Status       string     `xorm:"'status'" json:"status"`
	Notes        string     `xorm:"'notes'" json:"notes"`
	CreateTime   time.Time  `json:"create_time" xorm:"'create_time' created"`
	UpdateTime   time.Time  `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime   *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type InsertNewVisitRequest struct {
	PatientUUID string `json:"patient_uuid"`
	TrxPatientVisit
}

type UpdatePatientVisitRequest struct {
	InsertNewVisitRequest
}

type GetPatientVisitParams struct {
	PatientID   int64  `xorm:"id"`
	PatientUUID string `xorm:"uuid"`
}
