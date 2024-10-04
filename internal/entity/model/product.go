package model

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type MstProduct struct {
	ID          int64      `xorm:"'id' pk autoincr" json:"id"`
	Name        string     `xorm:"'name'" json:"name"`
	Description string     `xorm:"'description'" json:"description"`
	AddedBy     string     `xorm:"'added_by'" json:"added_by"`
	CreateTime  time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime  time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime  *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type TrxInstitutionProduct struct {
	ID               int64      `xorm:"'id' pk autoincr" json:"id"`
	Name             string     `xorm:"'name'" json:"name"`
	IDMstProduct     null.Int64 `xorm:"id_mst_product" json:"id_mst_product"`
	IDMstInstitution int64      `xorm:"id_mst_institution" json:"id_mst_institution"`
	Price            float64    `xorm:"'price'" json:"price"`
	IsItem           bool       `xorm:"'is_item'" json:"is_item"`
	IsTreatment      bool       `xorm:"'is_treatment'" json:"is_treatment"`
	CreateTime       time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime       time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime       *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type TrxVisitProduct struct {
	ID                      int64      `xorm:"'id' pk autoincr" json:"id"`
	IDTrxInstitutionProduct int64      `xorm:"'id_trx_institution_product'" json:"id_trx_institution_product"`
	IDTrxPatientVisit       int64      `xorm:"'id_trx_patient_visit'" json:"id_trx_patient_visit"`
	Quantity                int        `xorm:"'quantity'" json:"quantity"`
	UnitType                string     `xorm:"'unit_type'" json:"unit_type"`
	Price                   float64    `xorm:"'price'" json:"price"`
	DiscountAmount          float64    `xorm:"'discount_amount'" json:"discount_amount"`
	DiscountPrice           float64    `xorm:"'discount_price'" json:"discount_price"`
	FinalPrice              float64    `xorm:"'final_price'" json:"final_price"`
	AdjustedPrice           float64    `xorm:"adjusted_price" json:"adjusted_price"`
	CreateTime              time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime              time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime              *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type DtlInstitutionProductStock struct {
	ID                      int64      `xorm:"'id' pk autoincr" json:"id"`
	Quantity                int64      `xorm:"'quantity'" json:"quantity"`
	UnitType                string     `xorm:"'unit_type'" json:"unit_type"`
	IDMstInstitutionProduct int64      `xorm:"id_mst_institution_product" json:"id_mst_institution_product"`
	CreateTime              time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime              time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime              *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type TrxInstitutionProductJoinStock struct {
	TrxInstitutionProduct      TrxInstitutionProduct      `xorm:"extends"`
	DtlInstitutionProductStock DtlInstitutionProductStock `xorm:"extends"`
}
