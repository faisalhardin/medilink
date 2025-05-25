package model

import (
	"time"

	"github.com/volatiletech/null/v8"
)

type MstProduct struct {
	ID          int64      `xorm:"'id' pk autoincr" json:"id"`
	Name        string     `xorm:"'name'" json:"name" schema:"name"`
	Description string     `xorm:"'description'" json:"description" schema:"description"`
	AddedBy     string     `xorm:"'added_by'" json:"added_by" schema:"added_by"`
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

type FindTrxInstitutionProductParams struct {
	IDs              []int64 `schema:"id"`
	Name             string  `schema:"name"`
	IDMstProduct     []int64 `schema:"id_mst_product"`
	IDMstProducts    []int64 `schema:"id_mst_product"`
	IDMstInstitution int64   `schema:"id_mst_institution"`
	IsItem           bool    `schema:"is_item"`
	IsTreatment      bool    `schema:"is_treatment"`
	CommonRequestPayload
}

type TrxVisitProduct struct {
	ID                      int64      `xorm:"'id' pk autoincr" json:"id"`
	IDTrxInstitutionProduct int64      `xorm:"'id_trx_institution_product'" json:"id_trx_institution_product"`
	IDMstInstitution        int64      `xorm:"'id_mst_institution'" json:"-"`
	IDTrxPatientVisit       int64      `xorm:"'id_trx_patient_visit'" json:"id_trx_patient_visit"`
	IDDtlPatientVisit       int64      `xorm:"'id_dtl_patient_visit'" json:"id_dtl_patient_visit"`
	Quantity                int        `xorm:"'quantity'" json:"quantity"`
	UnitType                string     `xorm:"'unit_type'" json:"unit_type"`
	Price                   float64    `xorm:"'price'" json:"price"`
	DiscountRate            float64    `xorm:"'discount_rate'" json:"discount_rate"`
	DiscountPrice           float64    `xorm:"'discount_price'" json:"discount_price"`
	TotalPrice              float64    `xorm:"'total_price'" json:"total_price"`
	AdjustedPrice           float64    `xorm:"adjusted_price" json:"adjusted_price"`
	CreateTime              time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime              time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime              *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type DtlInstitutionProductStock struct {
	ID                      int64      `xorm:"'id' pk autoincr" json:"id"`
	Quantity                int64      `xorm:"'quantity'" json:"quantity"`
	UnitType                string     `xorm:"'unit_type'" json:"unit_type"`
	IDTrxInstitutionProduct int64      `xorm:"id_trx_institution_product" json:"id_trx_institution_product"`
	CreateTime              time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime              time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime              *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type TrxInstitutionProductJoinStock struct {
	TrxInstitutionProduct      TrxInstitutionProduct      `xorm:"extends"`
	DtlInstitutionProductStock DtlInstitutionProductStock `xorm:"extends"`
}

type InsertInstitutionProductRequest struct {
	Name         string     `json:"name"`
	IDMstProduct null.Int64 `json:"id_mst_product"`
	Price        float64    `json:"price"`
	IsItem       bool       `json:"is_item"`
	IsTreatment  bool       `json:"is_treatment"`
	Quantity     int64      `json:"quantity"`
	UnitType     string     `json:"unit_type"`
}

type UpdateInstitutionProductRequest struct {
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	IDMstProduct null.Int64  `json:"id_mst_product"`
	Price        float64     `json:"price"`
	IsItem       bool        `json:"is_item"`
	IsTreatment  bool        `json:"is_treatment"`
	Quantity     null.Int64  `json:"quantity"`
	UnitType     null.String `json:"unit_type"`
}

type GetInstitutionProductResponse struct {
	ID           int64      `xorm:"'id'" json:"id"`
	Name         string     `xorm:"'name'" json:"name"`
	IDMstProduct null.Int64 `xorm:"id_mst_product" json:"id_mst_product,omitempty"`
	Price        float64    `xorm:"'price'" json:"price,omitempty"`
	IsItem       bool       `xorm:"'is_item'" json:"is_item,omitempty"`
	IsTreatment  bool       `xorm:"'is_treatment'" json:"is_treatment"`
	Quantity     int64      `xorm:"'quantity'" json:"quantity"`
	UnitType     string     `xorm:"'unit_type'" json:"unit_type,omitempty"`
}

type InsertTrxVisitProductRequest struct {
	Products          []PurchasedProduct `json:"products"`
	IDDtlPatientVisit int64              `json:"id_dtl_patient_visit"`
}

type PurchasedProduct struct {
	IDTrxInstitutionProduct int64   `json:"id_trx_institution_product"`
	Quantity                int     `json:"quantity"`
	DiscountRate            float64 `json:"discount_rate,omitempty"`
	DiscountPrice           float64 `json:"discount_price,omitempty"`
	AdjustedPrice           float64 `json:"adjusted_price,omitempty"`
}

type UpdateTrxVisitProductRequest struct {
	ID                      int64   `json:"id"`
	IDTrxInstitutionProduct int64   `json:"id_trx_institution_product"`
	IDTrxPatientVisit       int64   `json:"id_trx_patient_visit"`
	Quantity                int     `json:"quantity"`
	Price                   float64 `json:"price"`
	DiscountAmount          float64 `json:"discount_amount"`
	DiscountPrice           float64 `json:"discount_price"`
	FinalPrice              float64 `json:"final_price"`
	AdjustedPrice           float64 `json:"adjusted_price"`
}

type DeleteTrxVisitProductRequest struct {
	ID int64 `json:"id"`
}
