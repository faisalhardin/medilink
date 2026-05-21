package model

import (
	"time"

	customtime "github.com/faisalhardin/medilink/pkg/type/time"
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
	Name                    string     `xorm:"name" json:"name"`
	Quantity                int        `xorm:"'quantity'" json:"quantity"`
	UnitType                string     `xorm:"'unit_type'" json:"unit_type"`
	Price                   float64    `xorm:"'price'" json:"price"`
	DiscountRate            float64    `xorm:"'discount_rate'" json:"discount_rate"`
	DiscountPrice           float64    `xorm:"'discount_price'" json:"discount_price"`
	TotalPrice              float64    `xorm:"'total_price'" json:"total_price"`
	AdjustedPrice           float64    `xorm:"adjusted_price" json:"adjusted_price"`
	IDMstStaffCreatedBy     int64      `xorm:"'id_mst_staff_created_by'" json:"staff_created_by"`
	IDMstStaffUpdatedBy     int64      `xorm:"'id_mst_staff_updated_by'" json:"staff_updated_by"`
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

type ProductStockResupplyItem struct {
	IDTrxInstitutionProduct int64 `json:"product_id" validate:"required"`
	Quantity                int64 `json:"quantity" validate:"required,ne=0"` // not equal 0
}

type ProductStockResupplyRequest struct {
	Products []ProductStockResupplyItem `json:"products" validate:"dive,required"`
}

type TrxInstitutionProductJoinStock struct {
	TrxInstitutionProduct      TrxInstitutionProduct      `xorm:"extends"`
	DtlInstitutionProductStock DtlInstitutionProductStock `xorm:"extends"`
}

type InsertInstitutionProductRequest struct {
	Name         string     `json:"name" validate:"required"`
	IDMstProduct null.Int64 `json:"id_mst_product"`
	Price        float64    `json:"price"`
	IsItem       bool       `json:"is_item"`
	IsTreatment  bool       `json:"is_treatment"`
	Quantity     int64      `json:"quantity" validate:"gte=0"`
	UnitType     string     `json:"unit_type" validate:"required"`
}

type UpdateInstitutionProductRequest struct {
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	IDMstProduct null.Int64  `json:"id_mst_product"`
	Price        float64     `json:"price"`
	IsItem       null.Bool   `json:"is_item"`
	IsTreatment  null.Bool   `json:"is_treatment"`
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
	IDDtlPatientVisit int64              `json:"patient_visit_detail"`
	IDTrxPatientVisit int64              `json:"visit_id"`
}

type UpsertTrxVisitProductRequest struct {
	Products          []PurchasedProduct `json:"products"`
	IDDtlPatientVisit int64              `json:"patient_visit_detail"`
	IDTrxPatientVisit int64              `json:"visit_id"`
}

type GetVisitProductRequest struct {
	VisitProductID int64 `json:"visit_product_id" schema:"visit_product_id"`
	VisitID        int64 `json:"visit_id" schema:"visit_id" validate:"required"`
	InstitutionID  int64
	VisitIDs       []int64
}

type PurchasedProduct struct {
	IDTrxInstitutionProduct int64   `json:"id"`
	Quantity                int     `json:"quantity"`
	Name                    string  `json:"name,omitempty"`
	Price                   int32   `json:"price"`
	TotalPrice              int32   `json:"total_price,omitempty"`
	UnitType                string  `json:"unit_type,omitempty"`
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

const (
	ProductStatisticsGranularityHour = "hour"
	ProductStatisticsGranularityDay  = "day"
	ProductStatisticsGranularityWeek = "week"
)

// ProductStatisticsParams is bound from GET query parameters.
type ProductStatisticsParams struct {
	StartTime   customtime.Time `schema:"start_time" validate:"required"`
	EndTime     customtime.Time `schema:"end_time" validate:"required"`
	Granularity string          `schema:"granularity"`
	ProductID   int64           `schema:"product_id"`
}

// ProductStatisticsQuery is the validated, institution-scoped query passed to the repo.
type ProductStatisticsQuery struct {
	IDMstInstitution        int64
	StartTime               time.Time
	EndTime                 time.Time
	Granularity             string
	IDTrxInstitutionProduct int64
}

// ProductStatisticsRow is one grouped row returned from the database.
type ProductStatisticsRow struct {
	PeriodStart             time.Time `xorm:"period_start"`
	IDTrxInstitutionProduct int64     `xorm:"id_trx_institution_product"`
	Name                    string    `xorm:"name"`
	TotalQuantity           int64     `xorm:"total_quantity"`
	TotalRevenue            float64   `xorm:"total_revenue"`
	AvgUnitPrice            float64   `xorm:"avg_unit_price"`
}

type ProductStatisticsPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ProductStatisticsProductItem struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	UnitPrice     float64 `json:"unit_price"`
	TotalQuantity int64   `json:"total_quantity"`
	TotalRevenue  float64 `json:"total_revenue"`
}

type ProductStatisticsBucket struct {
	PeriodStart   string                         `json:"period_start"`
	PeriodEnd     string                         `json:"period_end"`
	TotalRevenue  float64                        `json:"total_revenue"`
	TotalQuantity int64                          `json:"total_quantity"`
	Products      []ProductStatisticsProductItem `json:"products"`
}

type ProductStatisticsSummaryItem struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	UnitPrice     float64 `json:"unit_price"`
	TotalQuantity int64   `json:"total_quantity"`
	TotalRevenue  float64 `json:"total_revenue"`
}

type ProductStatisticsSummary struct {
	TotalRevenue          float64                        `json:"total_revenue"`
	TotalQuantity         int64                          `json:"total_quantity"`
	TopProductsByRevenue  []ProductStatisticsSummaryItem `json:"top_products_by_revenue"`
	TopProductsByQuantity []ProductStatisticsSummaryItem `json:"top_products_by_quantity"`
}

type ProductStatisticsResponse struct {
	Period      ProductStatisticsPeriod   `json:"period"`
	Granularity string                    `json:"granularity"`
	UTCOffset   string                    `json:"utc_offset"`
	Buckets     []ProductStatisticsBucket `json:"buckets"`
	Summary     ProductStatisticsSummary  `json:"summary"`
}
