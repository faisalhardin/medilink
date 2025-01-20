package model

import (
	"time"
)

type TrxPatientVisit struct {
	ID                int64      `xorm:"'id' pk autoincr" json:"id"`
	IDMstPatient      int64      `xorm:"'id_mst_patient'" json:"-"`
	IDMstInstitution  int64      `xorm:"'id_mst_institution'" json:"-"`
	IDMstJourneyBoard int64      `xorm:"'id_mst_journey_board'" json:"board_id"`
	IDMstJourneyPoint int64      `xorm:"'id_mst_journey_point' null" json:"journey_point_id"`
	IDMstServicePoint int64      `xorm:"'id_mst_service_point' null" json:"service_point_id"`
	Action            string     `xorm:"'action'" json:"action"`
	Status            string     `xorm:"'status'" json:"status"`
	Notes             string     `xorm:"'notes'" json:"notes"`
	CreateTime        time.Time  `json:"create_time" xorm:"'create_time' created"`
	UpdateTime        time.Time  `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime        *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type DtlPatientVisit struct {
	ID                int64      `xorm:"'id' pk autoincr" json:"id"`
	IDTrxPatientVisit int64      `xorm:"'id_trx_patient_visit'" json:"id_trx_patient_visit"`
	JourneyPointName  string     `xorm:"'name_mst_journey_point'" json:"name_mst_journey_point"`
	IDMstJourneyPoint string     `xorm:"id_mst_journey_point" json:"journey_point_id"`
	ActionBy          int64      `xorm:"action_by_id_mst_staff" json:"-"`
	Notes             string     `xorm:"'notes'" json:"notes"`
	CreateTime        time.Time  `json:"create_time" xorm:"'create_time' created"`
	UpdateTime        time.Time  `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime        *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type InsertNewVisitRequest struct {
	PatientUUID string `json:"patient_uuid"`
	TrxPatientVisit
}

type UpdatePatientVisitRequest struct {
	InsertNewVisitRequest
}

type GetPatientVisitParams struct {
	PatientID         int64  `xorm:"id"`
	PatientUUID       string `xorm:"uuid"`
	IDPatientVisit    int64  `xorm:"id_trx_patient_visit" schema:"visit_id"`
	IDMstInstitution  int64  `xorm:"id_mst_institution"`
	IDMstJourneyBoard int64  `schema:"journey_board_id"`
}

type DtlPatientVisitRequest struct {
	ID                int64  `json:"id" schema:"id"`
	IDTrxPatientVisit int64  `json:"id_trx_patient_visit" schema:"id_trx_patient_visit"`
	JourneyPointName  string `json:"name_mst_journey_point"`
	Notes             string `json:"notes"`
}

type InsertPatientVisitRequest struct {
	ID                int64  `json:"id" validation:"required"`
	IDTrxPatientVisit int64  `json:"id_trx_patient_visit" validation:"required"`
	JourneyPointName  string `json:"name_mst_journey_point"`
	Notes             string `json:"notes"`
}

type GetPatientVisitDetailResponse struct {
	TrxPatientVisit
	DtlPatientVisit []DtlPatientVisit `json:"patient_checkpoints"`
}
