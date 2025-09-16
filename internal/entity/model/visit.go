package model

import (
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
)

type TrxPatientVisit struct {
	ID                          int64           `xorm:"'id' pk autoincr" json:"id"`
	IDMstPatient                int64           `xorm:"'id_mst_patient'" json:"-"`
	IDMstInstitution            int64           `xorm:"'id_mst_institution'" json:"-"`
	IDMstJourneyBoard           int64           `xorm:"'id_mst_journey_board'" json:"board_id"`
	IDMstJourneyPoint           int64           `xorm:"'id_mst_journey_point' null" json:"-"`
	ShortIDMstJourneyPoint      string          `xorm:"-" json:"journey_point_id"`
	IDMstServicePoint           int64           `xorm:"'id_mst_service_point' null" json:"service_point_id"`
	Action                      string          `xorm:"'action'" json:"action"`
	Status                      string          `xorm:"'status'" json:"status"`
	Notes                       string          `xorm:"'notes'" json:"notes"`
	CreateTime                  time.Time       `json:"create_time" xorm:"'create_time' created"`
	UpdateTime                  time.Time       `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime                  *time.Time      `json:"-" xorm:"'delete_time' deleted"`
	UpdateTimeMstJourneyPointID int64           `json:"column_update_time" xorm:"'mst_journey_point_id_update_unix_time' created"`
	ProductCart                 json.RawMessage `xorm:"'product_cart'" json:"product_cart"`
}

func (tbl *TrxPatientVisit) BeforeUpdate() {
	if tbl.IDMstJourneyPoint > 0 {
		tbl.UpdateTimeMstJourneyPointID = time.Now().Unix()
	}
}

type DtlPatientVisit struct {
	ID                 int64           `xorm:"'id' pk autoincr" json:"id"`
	IDTrxPatientVisit  int64           `xorm:"'id_trx_patient_visit'" json:"id_trx_patient_visit"`
	IDsTrxPatientVisit []int64         `xorm:"-" json:"-"`
	JourneyPointName   string          `xorm:"'name_mst_journey_point'" json:"name_mst_journey_point"`
	IDMstJourneyPoint  int64           `xorm:"id_mst_journey_point" json:"journey_point_id"`
	ActionBy           int64           `xorm:"action_by_id_mst_staff" json:"-"`
	Notes              json.RawMessage `xorm:"'notes'" json:"notes"`
	Contributors       json.RawMessage `xorm:"'contributors'" json:"contributors"`
	IDMstServicePoint  int64           `xorm:"id_mst_service_point" json:"service_point_id"`
	CreateTime         time.Time       `json:"create_time" xorm:"'create_time' created"`
	UpdateTime         time.Time       `json:"update_time" xorm:"'update_time' updated"`
	DeleteTime         *time.Time      `json:"-" xorm:"'delete_time' deleted"`
}

type GetDtlPatientVisitParams struct {
	IDs                []int64 `xorm:"'id' pk autoincr" json:"id"`
	IDsTrxPatientVisit []int64 `xorm:"'id_trx_patient_visit'" json:"id_trx_patient_visit"`
	IDsMstJourneyPoins []int64 `xorm:"id_mst_journey_point" json:"journey_point_id"`
	IDsMstServicePoint []int64 `xorm:"id_mst_service_point" json:"service_point_id"`
	CommonRequestPayload
}

func (dtlPatientVisit *DtlPatientVisit) AddContributor(contributorEmail string) (isNewContributor bool, err error) {
	contributors := []string{}
	contributorsSet := make(map[string]bool)
	if dtlPatientVisit.Contributors != nil {
		err := json.Unmarshal(dtlPatientVisit.Contributors, &contributors)
		if err != nil {
			return false, err
		}

		for _, contributor := range contributors {
			contributorsSet[contributor] = true
		}
	}

	if !contributorsSet[contributorEmail] {
		contributors = append(contributors, contributorEmail)
		isNewContributor = true
	}

	if !isNewContributor {
		return false, nil
	}

	newContributorsJSON, err := json.Marshal(contributors)
	if err != nil {
		return false, err
	}

	dtlPatientVisit.Contributors = newContributorsJSON
	return true, nil
}

type InsertNewVisitRequest struct {
	PatientUUID         string          `json:"patient_uuid"`
	JourneyPointShortID string          `json:"journey_point_id"`
	Notes               json.RawMessage `json:"notes"`
}

type UpdatePatientVisitRequest struct {
	ID                          int64               `xorm:"'id' pk autoincr" json:"id"`
	IDMstInstitution            int64               `xorm:"'id_mst_institution'" json:"-"`
	IDMstJourneyBoard           null.Int64          `xorm:"'id_mst_journey_board'" json:"board_id"`
	ShortIDMstJourneyPoint      null.String         `xorm:"'id_mst_journey_point' null" json:"journey_point_id"`
	IDMstJourneyPoint           null.Int64          `json:"-"`
	IDMstServicePoint           null.Int64          `xorm:"'id_mst_service_point' null" json:"service_point_id"`
	Action                      null.String         `xorm:"'action'" json:"action"`
	Status                      null.String         `xorm:"'status'" json:"status"`
	UpdateTimeMstJourneyPointID null.Int64          `json:"column_update_time" xorm:"'mst_journey_point_id_update_unix_time' created"`
	ProductCart                 *[]PurchasedProduct `xorm:"'product_cart'" json:"product_cart"`
}

type GetPatientVisitParams struct {
	PatientID              int64  `xorm:"id" schema:"-"`
	PatientUUID            string `xorm:"uuid" schema:"patient_uuid"`
	IDPatientVisit         int64  `xorm:"id_trx_patient_visit" schema:"visit_id"`
	IDMstInstitution       int64  `xorm:"id_mst_institution"`
	IDMstJourneyBoard      int64  `schema:"journey_board_id"`
	ShortIDMstJourneyPoint string `schema:"journey_point_id"`
	CommonRequestPayload
}

type GetPatientVisitResponse struct {
	TrxPatientVisit       TrxPatientVisit       `xorm:"extends"`
	MstPatientInstitution MstPatientInstitution `xorm:"extends"`
	MstServicePoint       MstServicePoint       `xorm:"extends"`
	MstJourneyPoint       MstJourneyPoint       `xorm:"extends"`
}

type ListPatientVisitBoards struct {
	ID                          int64     `xorm:"'id' pk autoincr" json:"id"`
	IDMstJourneyBoard           int64     `xorm:"'id_mst_journey_board'" json:"board_id"`
	ShortIDMstJourneyPoint      string    `xorm:"'id_mst_journey_point' null" json:"journey_point_id"`
	IDMstServicePoint           int64     `xorm:"'id_mst_service_point' null" json:"service_point_id"`
	NameMstServicePoint         string    `xorm:"'service_point_name'" json:"service_point_name"`
	Action                      string    `xorm:"'action'" json:"action"`
	Status                      string    `xorm:"'status'" json:"status"`
	Notes                       string    `xorm:"'notes'" json:"notes"`
	CreateTime                  time.Time `json:"create_time" xorm:"'create_time' created"`
	UUID                        string    `json:"uuid" xorm:"'uuid' <-"`
	Name                        string    `json:"name" xorm:"'name'"`
	Sex                         string    `json:"sex" xorm:"'sex'"`
	UpdateTimeMstJourneyPointID int64     `json:"column_update_time" xorm:"'mst_journey_point_id_update_unix_time' created"`
}

type DtlPatientVisitRequest struct {
	ID                int64           `json:"id" schema:"id"`
	IDTrxPatientVisit int64           `json:"id_trx_patient_visit" schema:"id_trx_patient_visit"`
	JourneyPointName  string          `json:"name_mst_journey_point"`
	IDMstJourneyPoint int64           `json:"id_mst_journey_point"`
	Notes             json.RawMessage `json:"notes"`
	IDMstServicePoint null.Int64      `json:"id_mst_service_point"`
}

type InsertPatientVisitRequest struct {
	ID                int64  `json:"id" validation:"required"`
	IDTrxPatientVisit int64  `json:"id_trx_patient_visit" validation:"required"`
	JourneyPointName  string `json:"name_mst_journey_point"`
	Notes             string `json:"notes"`
}

type GetPatientVisitDetailResponse struct {
	TrxPatientVisit
	DtlPatientVisit []DtlPatientVisit     `json:"patient_journeypoints"`
	MstPatient      MstPatientInstitution `json:"patient"`
	JourneyPoint    MstJourneyPoint       `json:"journey_point"`
	ServicePoint    MstServicePoint       `json:"service_point"`
}

type ArchivePatientVisitRequest struct {
	ID int64 `json:"id" validation:"required"`
}
