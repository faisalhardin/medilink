package model

import (
	"database/sql"
	"time"

	customtime "github.com/faisalhardin/medilink/pkg/type/time"
	"github.com/volatiletech/null/v8"
)

const (
	TRX_VISIT_PROCEDURE_TABLE = "mdl_trx_visit_procedure"
	REF_ICD9CM_TABLE          = "mdl_ref_icd9cm"
)

// TrxVisitProcedure is the persisted procedure record for a visit.
// Snapshot columns are captured at write time to avoid JOINs on reads
// and to preserve history when master records change.
type TrxVisitProcedure struct {
	ID            int64          `xorm:"'id' pk autoincr" json:"-"`
	VisitID       int64          `xorm:"'visit_id'" json:"-"`
	InstitutionID int64          `xorm:"'institution_id'" json:"-"`
	ProductID     sql.NullInt64  `xorm:"'product_id' null" json:"-"`
	ProductName   sql.NullString `xorm:"'product_name' null" json:"-"`
	DoctorID      string         `xorm:"'doctor_id'" json:"-"`
	DoctorName    string         `xorm:"'doctor_name'" json:"-"`
	NurseID       sql.NullString `xorm:"'nurse_id' null" json:"-"`
	NurseName     sql.NullString `xorm:"'nurse_name' null" json:"-"`
	PlannedAt     *time.Time     `xorm:"'planned_at' null" json:"-"`
	Category      sql.NullString `xorm:"'category' null" json:"-"`
	Duration      sql.NullString `xorm:"'duration' null" json:"-"`
	ICD9CMCode    sql.NullString `xorm:"'icd9cm_code' null" json:"-"`
	ICD9CMDisplay sql.NullString `xorm:"'icd9cm_display' null" json:"-"`
	Description   sql.NullString `xorm:"'description' null" json:"-"`
	Notes         sql.NullString `xorm:"'notes' null" json:"-"`
	Rank          int16          `xorm:"'rank'" json:"-"`
	CreatedAt     time.Time      `xorm:"'created_at' created" json:"-"`
	UpdatedAt     time.Time      `xorm:"'updated_at' updated" json:"-"`
	DeletedAt     *time.Time     `xorm:"'deleted_at' null" json:"-"`
}

// RefICD9CM represents a row in the mdl_ref_icd9cm table.
type RefICD9CM struct {
	Code       string  `xorm:"'code' pk"`
	Display    string  `xorm:"'display'"`
	ParentCode *string `xorm:"'parent_code' null"`
	Depth      int16   `xorm:"'depth'"`
	IsLeaf     bool    `xorm:"'is_leaf'"`
	Version    string  `xorm:"'version'"`
}

// ─── Response DTOs ────────────────────────────────────────────────────────────

// ICD9CMOption is the search result shape for GET /v1/icd9cm/search.
type ICD9CMOption struct {
	Code    string `json:"code"`
	Display string `json:"display"`
	Depth   int16  `json:"depth"`
	IsLeaf  bool   `json:"is_leaf"`
}

// ProcedureEntry is the response shape for GET /v1/visit/:visitId/procedure.
type ProcedureEntry struct {
	ID            int64           `json:"id"`
	VisitID       int64           `json:"visit_id"`
	ProductID     null.Int64      `json:"product_id"`
	ProductName   null.String     `json:"product_name"`
	DoctorID      string          `json:"doctor_id"`
	DoctorName    string          `json:"doctor_name"`
	NurseID       null.String      `json:"nurse_id"`
	NurseName     null.String      `json:"nurse_name"`
	PlannedAt     *customtime.Time `json:"planned_at,omitempty"`
	Category      null.String      `json:"category"`
	Duration      null.String      `json:"duration"`
	ICD9CMCode    null.String      `json:"icd9cm_code"`
	ICD9CMDisplay null.String      `json:"icd9cm_display"`
	Description   null.String      `json:"description"`
	Notes         null.String      `json:"notes"`
	Rank          int16            `json:"rank"`
	CreatedAt     customtime.Time  `json:"created_at"`
	UpdatedAt     customtime.Time  `json:"updated_at"`
}

// ToResponse converts the DB struct to the JSON response DTO.
func (r TrxVisitProcedure) ToResponse() ProcedureEntry {
	entry := ProcedureEntry{
		ID:            r.ID,
		VisitID:       r.VisitID,
		ProductID:     null.Int64{Int64: r.ProductID.Int64, Valid: r.ProductID.Valid},
		ProductName:   null.String{String: r.ProductName.String, Valid: r.ProductName.Valid},
		DoctorID:      r.DoctorID,
		DoctorName:    r.DoctorName,
		NurseID:       null.String{String: r.NurseID.String, Valid: r.NurseID.Valid},
		NurseName:     null.String{String: r.NurseName.String, Valid: r.NurseName.Valid},
		Category:      null.String{String: r.Category.String, Valid: r.Category.Valid},
		Duration:      null.String{String: r.Duration.String, Valid: r.Duration.Valid},
		ICD9CMCode:    null.String{String: r.ICD9CMCode.String, Valid: r.ICD9CMCode.Valid},
		ICD9CMDisplay: null.String{String: r.ICD9CMDisplay.String, Valid: r.ICD9CMDisplay.Valid},
		Description:   null.String{String: r.Description.String, Valid: r.Description.Valid},
		Notes:         null.String{String: r.Notes.String, Valid: r.Notes.Valid},
		Rank:          r.Rank,
		CreatedAt:     customtime.Time{Time: r.CreatedAt},
		UpdatedAt:     customtime.Time{Time: r.UpdatedAt},
	}
	if r.PlannedAt != nil {
		ct := customtime.Time{Time: *r.PlannedAt}
		entry.PlannedAt = &ct
	}
	return entry
}

// ProcedureHistoryEntry is the response shape for GET /v1/patient/:uuid/procedure/history.
type ProcedureHistoryEntry struct {
	VisitID       int64       `json:"visit_id" xorm:"'visit_id'"`
	VisitDate     string      `json:"visit_date" xorm:"'visit_date'"`
	ProductName   null.String `json:"product_name" xorm:"'product_name'"`
	ICD9CMCode    null.String `json:"icd9cm_code" xorm:"'icd9cm_code'"`
	ICD9CMDisplay null.String `json:"icd9cm_display" xorm:"'icd9cm_display'"`
	DoctorName    string      `json:"doctor_name" xorm:"'doctor_name'"`
}

// ProcedureHistoryResponse is the paginated envelope for GET /v1/patient/:uuid/procedure/history.
type ProcedureHistoryResponse struct {
	Data   []ProcedureHistoryEntry `json:"data"`
	Total  int                     `json:"total"`
	Limit  int                     `json:"limit"`
	Offset int                     `json:"offset"`
}

// SaveProceduresResult is the response envelope for POST /v1/visit/:visitId/procedure.
type SaveProceduresResult struct {
	Message string                `json:"message"`
	Data    SaveProceduresSummary `json:"data"`
}

type SaveProceduresSummary struct {
	Saved   int `json:"saved"`
	Deleted int `json:"deleted"`
}

// ─── Request DTOs ─────────────────────────────────────────────────────────────

// ICD9CMSearchRequest binds GET /v1/icd9cm/search query params.
type ICD9CMSearchRequest struct {
	Query string `schema:"q" validate:"required,min=2,max=100"`
	Limit int    `schema:"limit" validate:"omitempty,min=1,max=50"`
}

// SaveProceduresRequest is the payload for POST /v1/visit/:visitId/procedure.
type SaveProceduresRequest struct {
	Procedures []SaveProcedureRow `json:"procedures" validate:"required"`
}

// SaveProcedureRow is a single procedure line within SaveProceduresRequest.
type SaveProcedureRow struct {
	ID          null.Int64  `json:"id"`
	ProductID   null.Int64  `json:"product_id"`
	DoctorID    string      `json:"doctor_id" validate:"required"`
	NurseID     null.String `json:"nurse_id"`
	PlannedAt   null.String `json:"planned_at"`
	Category    null.String `json:"category" validate:"omitempty,oneof=103693007 24642003 277132007 387713003 409063005 409073007 410606002 46947000"`
	Duration    null.String `json:"duration"`
	ICD9CMCode  null.String `json:"icd9cm_code"`
	Description null.String `json:"description"`
	Notes       null.String `json:"notes"`
	Rank        int16       `json:"rank" validate:"required,min=1"`
}
