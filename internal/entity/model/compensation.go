package model

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/volatiletech/null/v8"
)

const (
	MstStaffWageTableName          = "mdl_mst_staff_wage"
	TrxCompensationPeriodTableName = "mdl_trx_compensation_period"
	TrxVisitCommissionTableName    = "mdl_trx_visit_commission"
	MapVisitContributorTableName   = "mdl_map_visit_contributor"
	TrxWorksheetTableName          = "mdl_trx_worksheet"
)

// ─── Worksheet enums ──────────────────────────────────────────────────────────

// WorksheetStatus is the lifecycle of a single-staff commission worksheet.
type WorksheetStatus string

const (
	WorksheetStatusPending   WorksheetStatus = "pending"
	WorksheetStatusOpen      WorksheetStatus = "open"
	WorksheetStatusFinalized WorksheetStatus = "finalized"
)

func (s WorksheetStatus) IsValid() bool {
	switch s {
	case WorksheetStatusPending, WorksheetStatusOpen, WorksheetStatusFinalized:
		return true
	default:
		return false
	}
}

// WorksheetGenerateStatus tracks the async generate worker.
type WorksheetGenerateStatus string

const (
	WorksheetGenerateStatusIdle      WorksheetGenerateStatus = "idle"
	WorksheetGenerateStatusRunning   WorksheetGenerateStatus = "running"
	WorksheetGenerateStatusSucceeded WorksheetGenerateStatus = "succeeded"
	WorksheetGenerateStatusFailed    WorksheetGenerateStatus = "failed"
)

// ─── Enums ────────────────────────────────────────────────────────────────────

// WageCadence is the pay cadence for a staff wage contract.
type WageCadence string

const (
	WageCadenceMonthly WageCadence = "monthly"
	WageCadenceWeekly  WageCadence = "weekly"
)

func (c WageCadence) IsValid() bool {
	return c == WageCadenceMonthly || c == WageCadenceWeekly
}

// CompensationPeriodStatus is the lifecycle status of a payday period.
type CompensationPeriodStatus string

const (
	CompensationPeriodStatusOpen      CompensationPeriodStatus = "open"
	CompensationPeriodStatusDraft     CompensationPeriodStatus = "draft"
	CompensationPeriodStatusFinalized CompensationPeriodStatus = "finalized"
)

func (s CompensationPeriodStatus) IsValid() bool {
	switch s {
	case CompensationPeriodStatusOpen, CompensationPeriodStatusDraft, CompensationPeriodStatusFinalized:
		return true
	default:
		return false
	}
}

// CanTransition reports whether a status write from s to to is legal.
// Finalize of an already-finalized period is an idempotent no-op handled in the usecase, not here.
func (s CompensationPeriodStatus) CanTransition(to CompensationPeriodStatus) bool {
	switch s {
	case CompensationPeriodStatusOpen:
		return to == CompensationPeriodStatusDraft
	case CompensationPeriodStatusDraft:
		return to == CompensationPeriodStatusDraft || to == CompensationPeriodStatusFinalized
	case CompensationPeriodStatusFinalized:
		return to == CompensationPeriodStatusDraft
	default:
		return false
	}
}

// CompensationAssignmentStatus is staff commission completeness on a payday period.
type CompensationAssignmentStatus string

const (
	CompensationAssignmentStatusUnassigned CompensationAssignmentStatus = "unassigned"
	CompensationAssignmentStatusPartial    CompensationAssignmentStatus = "partial"
	CompensationAssignmentStatusComplete   CompensationAssignmentStatus = "complete"
)

// CommissionType is how a visit commission amount is calculated.
type CommissionType string

const (
	CommissionTypePercent CommissionType = "percent"
	CommissionTypeFlat    CommissionType = "flat"
)

func (t CommissionType) IsValid() bool {
	return t == CommissionTypePercent || t == CommissionTypeFlat
}

// ContributionSourceType is an attribution reason a staff appears on a visit (not pay math).
type ContributionSourceType string

const (
	ContributionSourceTypeProcedure ContributionSourceType = "procedure"
	ContributionSourceTypeDiagnosis ContributionSourceType = "diagnosis"
	ContributionSourceTypeAnamnesa  ContributionSourceType = "anamnesa"
	ContributionSourceTypeJourney   ContributionSourceType = "journey"
	ContributionSourceTypeManual    ContributionSourceType = "manual"
)

func (t ContributionSourceType) IsValid() bool {
	switch t {
	case ContributionSourceTypeProcedure,
		ContributionSourceTypeDiagnosis,
		ContributionSourceTypeAnamnesa,
		ContributionSourceTypeJourney,
		ContributionSourceTypeManual:
		return true
	default:
		return false
	}
}

// ─── Xorm entities ────────────────────────────────────────────────────────────

// TrxWorksheet is a single-staff wrap for visit commissions.
type TrxWorksheet struct {
	ID                   int64                   `xorm:"'id' pk autoincr" json:"-"`
	UUID                 string                  `xorm:"'uuid'" json:"-"`
	InstitutionID        int64                   `xorm:"'institution_id'" json:"-"`
	StaffID              string                  `xorm:"'staff_id'" json:"-"`
	Label                string                  `xorm:"'label'" json:"-"`
	PeriodStart          time.Time               `xorm:"'period_start'" json:"-"`
	PeriodEnd            time.Time               `xorm:"'period_end'" json:"-"`
	Status               WorksheetStatus         `xorm:"'status'" json:"-"`
	GenerateStatus       WorksheetGenerateStatus `xorm:"'generate_status'" json:"-"`
	CompensationPeriodID sql.NullInt64           `xorm:"'compensation_period_id' null" json:"-"`
	TotalCommission      int64                   `xorm:"'total_commission'" json:"-"`
	VisitCount           int64                   `xorm:"'visit_count'" json:"-"`
	GenerateStartedAt    sql.NullTime            `xorm:"'generate_started_at' null" json:"-"`
	GenerateFinishedAt   sql.NullTime            `xorm:"'generate_finished_at' null" json:"-"`
	GenerateError        sql.NullString          `xorm:"'generate_error' null" json:"-"`
	FinalizedAt          sql.NullTime            `xorm:"'finalized_at' null" json:"-"`
	FinalizedBy          sql.NullString          `xorm:"'finalized_by' null" json:"-"`
	CreatedBy            sql.NullString          `xorm:"'created_by' null" json:"-"`
	CreateTime           time.Time               `xorm:"'create_time' created" json:"-"`
	UpdateTime           time.Time               `xorm:"'update_time' updated" json:"-"`
	DeleteTime           *time.Time              `xorm:"'delete_time' deleted" json:"-"`
	// CompensationPeriodUUID is filled by GetByUUID/List via LEFT JOIN; not a DB column.
	CompensationPeriodUUID sql.NullString `xorm:"<- 'compensation_period_uuid'" json:"-"`
}

func (TrxWorksheet) TableName() string {
	return TrxWorksheetTableName
}

const worksheetDateLayout = "2006-01-02"

// ToResponse converts a TrxWorksheet row to the JSON DTO.
// CompensationPeriodUUID comes from the worksheet query join (or is set after attach).
func (w TrxWorksheet) ToResponse() WorksheetResponse {
	return WorksheetResponse{
		UUID:                   w.UUID,
		StaffID:                w.StaffID,
		Label:                  w.Label,
		PeriodStart:            w.PeriodStart.UTC().Format(worksheetDateLayout),
		PeriodEnd:              w.PeriodEnd.UTC().Format(worksheetDateLayout),
		Status:                 w.Status,
		GenerateStatus:         w.GenerateStatus,
		CompensationPeriodUUID: nullStringToNullable(w.CompensationPeriodUUID),
		TotalCommission:        w.TotalCommission,
		VisitCount:             w.VisitCount,
		GenerateStartedAt:      nullTimeFromSQL(w.GenerateStartedAt),
		GenerateFinishedAt:     nullTimeFromSQL(w.GenerateFinishedAt),
		GenerateError:          nullStringToNullable(w.GenerateError),
		FinalizedAt:            nullTimeFromSQL(w.FinalizedAt),
	}
}

func nullStringToNullable(v sql.NullString) null.String {
	if !v.Valid {
		return null.String{}
	}
	return null.StringFrom(v.String)
}

// MstStaffWage is a staff wage contract for an institution.
type MstStaffWage struct {
	ID            int64          `xorm:"'id' pk autoincr" json:"-"`
	StaffID       string         `xorm:"'staff_id'" json:"-"`
	InstitutionID int64          `xorm:"'institution_id'" json:"-"`
	WageAmount    int64          `xorm:"'wage_amount'" json:"-"`
	WageCadence   WageCadence    `xorm:"'wage_cadence'" json:"-"`
	IsActive      bool           `xorm:"'is_active'" json:"-"`
	EffectiveFrom time.Time      `xorm:"'effective_from'" json:"-"`
	EffectiveTo   sql.NullTime   `xorm:"'effective_to' null" json:"-"`
	CreatedBy     sql.NullString `xorm:"'created_by' null" json:"-"`
	UpdatedBy     sql.NullString `xorm:"'updated_by' null" json:"-"`
	CreateTime    time.Time      `xorm:"'create_time' created" json:"-"`
	UpdateTime    time.Time      `xorm:"'update_time' updated" json:"-"`
	DeleteTime    *time.Time     `xorm:"'delete_time' deleted" json:"-"`
}

func (MstStaffWage) TableName() string {
	return MstStaffWageTableName
}

// TrxCompensationPeriod is a payday period for an institution.
type TrxCompensationPeriod struct {
	ID              int64                    `xorm:"'id' pk autoincr" json:"-"`
	UUID            string                   `xorm:"'uuid'" json:"-"`
	InstitutionID   int64                    `xorm:"'institution_id'" json:"-"`
	Label           string                   `xorm:"'label'" json:"-"`
	PeriodStart     time.Time                `xorm:"'period_start'" json:"-"`
	PeriodEnd       time.Time                `xorm:"'period_end'" json:"-"`
	Status          CompensationPeriodStatus `xorm:"'status'" json:"-"`
	WageSnapshot    json.RawMessage          `xorm:"'wage_snapshot' jsonb" json:"-"`
	TotalWage       sql.NullInt64            `xorm:"'total_wage' null" json:"-"`
	TotalCommission sql.NullInt64            `xorm:"'total_commission' null" json:"-"`
	TotalPayout     sql.NullInt64            `xorm:"'total_payout' null" json:"-"`
	StaffCount      sql.NullInt64            `xorm:"'staff_count' null" json:"-"`
	VisitCount      sql.NullInt64            `xorm:"'visit_count' null" json:"-"`
	DraftedAt       sql.NullTime             `xorm:"'drafted_at' null" json:"-"`
	DraftedBy       sql.NullString           `xorm:"'drafted_by' null" json:"-"`
	FinalizedAt     sql.NullTime             `xorm:"'finalized_at' null" json:"-"`
	FinalizedBy     sql.NullString           `xorm:"'finalized_by' null" json:"-"`
	CreateTime      time.Time                `xorm:"'create_time' created" json:"-"`
	UpdateTime      time.Time                `xorm:"'update_time' updated" json:"-"`
	DeleteTime      *time.Time               `xorm:"'delete_time' deleted" json:"-"`
}

func (TrxCompensationPeriod) TableName() string {
	return TrxCompensationPeriodTableName
}

const compensationPeriodDateLayout = "2006-01-02"

// ToResponse maps a payday-period row to the JSON DTO. SQL NULL totals become 0.
func (p TrxCompensationPeriod) ToResponse(noContributorCount int64) CompensationPeriodResponse {
	return CompensationPeriodResponse{
		UUID:               p.UUID,
		Label:              p.Label,
		PeriodStart:        p.PeriodStart.UTC().Format(compensationPeriodDateLayout),
		PeriodEnd:          p.PeriodEnd.UTC().Format(compensationPeriodDateLayout),
		Status:             p.Status,
		TotalWage:          nullInt64OrZero(p.TotalWage),
		TotalCommission:    nullInt64OrZero(p.TotalCommission),
		TotalPayout:        nullInt64OrZero(p.TotalPayout),
		StaffCount:         nullInt64OrZero(p.StaffCount),
		VisitCount:         nullInt64OrZero(p.VisitCount),
		NoContributorCount: noContributorCount,
		DraftedAt:          nullTimeFromSQL(p.DraftedAt),
		FinalizedAt:        nullTimeFromSQL(p.FinalizedAt),
	}
}

func nullInt64OrZero(v sql.NullInt64) int64 {
	if !v.Valid {
		return 0
	}
	return v.Int64
}

func nullTimeFromSQL(v sql.NullTime) null.Time {
	if !v.Valid {
		return null.Time{}
	}
	return null.TimeFrom(v.Time)
}

// ListCompensationPeriodParams filters paginated payday-period reads.
// Empty Status means all statuses. Limit is applied only when greater than 0.
type ListCompensationPeriodParams struct {
	InstitutionID int64
	Status        CompensationPeriodStatus
	Limit         int
	Offset        int
}

// TrxVisitCommission is a per-visit, per-staff commission row within a worksheet.
type TrxVisitCommission struct {
	ID                   int64           `xorm:"'id' pk autoincr" json:"-"`
	WorksheetID          int64           `xorm:"'worksheet_id'" json:"-"`
	VisitID              int64           `xorm:"'visit_id'" json:"-"`
	StaffID              string          `xorm:"'staff_id'" json:"-"`
	RevenueBase          int64           `xorm:"'revenue_base'" json:"-"`
	CommissionType       CommissionType  `xorm:"'commission_type'" json:"-"`
	CommissionPercent    sql.NullFloat64 `xorm:"'commission_percent' null" json:"-"`
	CommissionFlatAmount sql.NullInt64   `xorm:"'commission_flat_amount' null" json:"-"`
	CommissionAmount     int64           `xorm:"'commission_amount'" json:"-"`
	Sources              json.RawMessage `xorm:"'sources' jsonb" json:"-"`
	Note                 sql.NullString  `xorm:"'note' null" json:"-"`
	IncludedManually     bool            `xorm:"'included_manually'" json:"-"`
	ApprovedAt           sql.NullTime    `xorm:"'approved_at' null" json:"-"`
	CreateTime           time.Time       `xorm:"'create_time' created" json:"-"`
	UpdateTime           time.Time       `xorm:"'update_time' updated" json:"-"`
	DeleteTime           *time.Time      `xorm:"'delete_time' deleted" json:"-"`
}

func (TrxVisitCommission) TableName() string {
	return TrxVisitCommissionTableName
}

// MapVisitContributor is a manual staff attachment to a visit for commission eligibility.
type MapVisitContributor struct {
	ID            int64      `xorm:"'id' pk autoincr" json:"-"`
	VisitID       int64      `xorm:"'visit_id'" json:"-"`
	StaffID       string     `xorm:"'staff_id'" json:"-"`
	InstitutionID int64      `xorm:"'institution_id'" json:"-"`
	AddedBy       string     `xorm:"'added_by'" json:"-"`
	CreateTime    time.Time  `xorm:"'create_time' created" json:"-"`
	DeleteTime    *time.Time `xorm:"'delete_time' deleted" json:"-"`
}

func (MapVisitContributor) TableName() string {
	return MapVisitContributorTableName
}

// ─── JSON shapes ──────────────────────────────────────────────────────────────

// ContributionSource is the attribution snapshot stored in commission sources JSONB.
// Optional fields use null.v8 for JSON encoding (not an Xorm row).
type ContributionSource struct {
	Type        ContributionSourceType `json:"type"`
	ProcedureID null.Int64             `json:"procedure_id"`
	DiagnosisID null.Int64             `json:"diagnosis_id"`
	ProductID   null.Int64             `json:"product_id"`
	Label       null.String            `json:"label"`
	LabelSource null.String            `json:"label_source"`
}

// VisitContributorResponse is one staff row on GET /v1/visit/{id}/contributors.
type VisitContributorResponse struct {
	StaffID       string             `json:"staff_id"`
	Name          string             `json:"name"`
	Source        ContributionSource `json:"source"`
	AddedManually bool               `json:"added_manually"`
}

// ListVisitContributorsResponse is the body for GET /v1/visit/{id}/contributors.
// CompensationLockedAt is set when a finalized worksheet locked the visit.
type ListVisitContributorsResponse struct {
	Contributors         []VisitContributorResponse `json:"contributors"`
	CompensationLockedAt null.Time                  `json:"compensation_locked_at"`
}

// AddVisitContributorRequest is the body for POST /v1/visit/{id}/contributors.
type AddVisitContributorRequest struct {
	StaffID string `json:"staff_id"`
}

// AddVisitContributorResponse is the body for POST /v1/visit/{id}/contributors.
type AddVisitContributorResponse struct {
	Contributor VisitContributorResponse `json:"contributor"`
}

// DeleteVisitContributorResponse is the body for DELETE /v1/visit/{id}/contributors/{staffId}.
type DeleteVisitContributorResponse struct {
	Success bool `json:"success"`
}

// ─── JSON request / response DTOs ─────────────────────────────────────────────

// CreateCompensationPeriodRequest is the body for POST /v1/compensation-period.
type CreateCompensationPeriodRequest struct {
	Label       string `json:"label"`
	PeriodStart Time   `json:"period_start"`
	PeriodEnd   Time   `json:"period_end"`
}

// ListCompensationPeriodsRequest is the query for GET /v1/compensation-period.
type ListCompensationPeriodsRequest struct {
	Status CompensationPeriodStatus `json:"status" schema:"status"`
	CommonRequestPayload
}

// CompensationPeriodResponse is the public payday-period shape.
type CompensationPeriodResponse struct {
	UUID               string                   `json:"uuid"`
	Label              string                   `json:"label"`
	PeriodStart        string                   `json:"period_start"`
	PeriodEnd          string                   `json:"period_end"`
	Status             CompensationPeriodStatus `json:"status"`
	TotalWage          int64                    `json:"total_wage"`
	TotalCommission    int64                    `json:"total_commission"`
	TotalPayout        int64                    `json:"total_payout"`
	StaffCount         int64                    `json:"staff_count"`
	VisitCount         int64                    `json:"visit_count"`
	NoContributorCount int64                    `json:"no_contributor_count"`
	DraftedAt          null.Time                `json:"drafted_at"`
	FinalizedAt        null.Time                `json:"finalized_at"`
}

// ListCompensationPeriodsResponse is the paginated period list.
type ListCompensationPeriodsResponse struct {
	Periods []CompensationPeriodResponse `json:"periods"`
	Total   int                          `json:"total"`
}

// FinalizeCompensationPeriodResponse is the body for POST .../finalize.
type FinalizeCompensationPeriodResponse struct {
	Period           CompensationPeriodResponse `json:"period"`
	LockedVisitCount int64                      `json:"locked_visit_count"`
}

// DeleteCompensationPeriodResponse is the body for DELETE .../periods/{periodId}.
type DeleteCompensationPeriodResponse struct {
	Success bool `json:"success"`
}

// ListCompensationPeriodStaffRequest is the query for GET /v1/compensation-period/{periodId}/staffs.
type ListCompensationPeriodStaffRequest struct {
	CommonRequestPayload
}

// CompensationPeriodStaffRow is one staff on GET /v1/compensation-period/{periodId}/staffs.
type CompensationPeriodStaffRow struct {
	StaffID            string                       `json:"staff_id"`
	Name               string                       `json:"name"`
	Roles              []string                     `json:"roles"`
	VisitCount         int64                        `json:"visit_count"`
	Wage               int64                        `json:"wage"`
	CommissionSubtotal int64                        `json:"commission_subtotal"`
	PayTotal           int64                        `json:"pay_total"`
	AssignmentStatus   CompensationAssignmentStatus `json:"assignment_status"`
}

// ListCompensationPeriodStaffResponse is the paginated staff list for a payday period.
type ListCompensationPeriodStaffResponse struct {
	Staff []CompensationPeriodStaffRow `json:"staff"`
	Total int                          `json:"total"`
}

// GetCompensationPeriodStaffRequest is the request for GET /v1/compensation-period/{periodId}/staff/{staffId}.
// PeriodUUID and StaffID are set from URL path params.
type GetCompensationPeriodStaffRequest struct {
	PeriodUUID string `json:"-" schema:"-"`
	StaffID    string `json:"-" schema:"-"`
}

// CompensationPeriodStaffInfo is the staff header on GET .../staff/{staffId}.
type CompensationPeriodStaffInfo struct {
	StaffID string   `json:"staff_id"`
	Name    string   `json:"name"`
	Roles   []string `json:"roles"`
}

// GetCompensationPeriodStaffResponse is the header body for GET .../staff/{staffId}.
type GetCompensationPeriodStaffResponse struct {
	StaffInfo    CompensationPeriodStaffInfo `json:"staff_info"`
	ComputedWage int64                       `json:"computed_wage"`
	WageOverride null.Int64                  `json:"wage_override"`
}

// PatchCommissionItemRequest is the body for PATCH /v1/visit-commissions/{id}.
// ID is set from the path param.
type PatchCommissionItemRequest struct {
	ID                   int64          `json:"-"`
	CommissionType       CommissionType `json:"commission_type"`
	CommissionPercent    null.Float64   `json:"commission_percent"`
	CommissionFlatAmount null.Int64     `json:"commission_flat_amount"`
	Note                 null.String    `json:"note"`
}

// PatchCommissionItemResponse is the body for PATCH /v1/visit-commissions/{id}.
type PatchCommissionItemResponse struct {
	UpdatedCount       int   `json:"updated_count"`
	CommissionSubtotal int64 `json:"commission_subtotal"`
}

// ─── Worksheet DTOs ───────────────────────────────────────────────────────────

// CreateWorksheetRequest is the body for POST /v1/worksheet.
type CreateWorksheetRequest struct {
	StaffID     string `json:"staff_id"`
	Label       string `json:"label"`
	PeriodStart Time   `json:"period_start"`
	PeriodEnd   Time   `json:"period_end"`
}

// ListWorksheetsRequest is the query for GET /v1/worksheet.
// Cursor is the last worksheet id from the previous page (exclusive); empty = first page.
// CompensationPeriodUUID, when set, returns only worksheets linked to that payday.
type ListWorksheetsRequest struct {
	StaffID                string          `schema:"staff_id"`
	Status                 WorksheetStatus `schema:"status"`
	CompensationPeriodUUID string          `schema:"compensation_period_uuid"`
	Cursor                 string          `schema:"cursor"`
	Limit                  int             `schema:"limit"`
	InstitutionID          int64           `schema:"-"`
}

// WorksheetResponse is the public worksheet shape.
// CompensationPeriodUUID is the public payday period UUID when attached (null when unattached).
// Internal BIGINT FK is not exposed.
type WorksheetResponse struct {
	UUID                   string                  `json:"uuid"`
	StaffID                string                  `json:"staff_id"`
	Label                  string                  `json:"label"`
	PeriodStart            string                  `json:"period_start"`
	PeriodEnd              string                  `json:"period_end"`
	Status                 WorksheetStatus         `json:"status"`
	GenerateStatus         WorksheetGenerateStatus `json:"generate_status"`
	CompensationPeriodUUID null.String             `json:"compensation_period_uuid"`
	TotalCommission        int64                   `json:"total_commission"`
	VisitCount             int64                   `json:"visit_count"`
	GenerateStartedAt      null.Time               `json:"generate_started_at"`
	GenerateFinishedAt     null.Time               `json:"generate_finished_at"`
	GenerateError          null.String             `json:"generate_error"`
	FinalizedAt            null.Time               `json:"finalized_at"`
}

// ListWorksheetsResponse is a cursor-paginated worksheet list (no total count).
type ListWorksheetsResponse struct {
	Worksheets []WorksheetResponse `json:"worksheets"`
	NextCursor null.String         `json:"next_cursor"`
}

// PatchWorksheetRequest is the body for PATCH /v1/worksheet/{id}.
// CompensationPeriodUUID: nil = omit (no change); non-nil invalid/empty = detach;
// non-nil valid string = attach by public period UUID.
type PatchWorksheetRequest struct {
	UUID                   string       `json:"-"`
	Label                  null.String  `json:"label"`
	PeriodStart            *Time        `json:"period_start"`
	PeriodEnd              *Time        `json:"period_end"`
	CompensationPeriodUUID *null.String `json:"compensation_period_uuid"`
}

// DeleteWorksheetResponse is the body for DELETE /v1/worksheet/{id}.
type DeleteWorksheetResponse struct {
	Success bool `json:"success"`
}

// FinalizeWorksheetResponse is the body for POST /v1/worksheet/{id}/finalize.
type FinalizeWorksheetResponse struct {
	Worksheet        WorksheetResponse `json:"worksheet"`
	LockedVisitCount int64             `json:"locked_visit_count"`
}

// GenerateVisitCommissionsRequest is the body for POST /v1/visit-commissions/generate.
// IncludeGenerate must be true to start the async worker; false returns current status without generating.
type GenerateVisitCommissionsRequest struct {
	WorksheetID     string `json:"worksheet_id"`
	IncludeGenerate bool   `json:"include_generate"`
}

// GenerateVisitCommissionsResponse is the body for POST /v1/visit-commissions/generate (202 Accepted).
type GenerateVisitCommissionsResponse struct {
	UUID           string                  `json:"uuid"`
	Status         WorksheetStatus         `json:"status"`
	GenerateStatus WorksheetGenerateStatus `json:"generate_status"`
}

// ListVisitCommissionsRequest is the query for GET /v1/visit-commissions
// (?staff_id=&start=&end=) or GET /v1/worksheet/{id}/commissions.
type ListVisitCommissionsRequest struct {
	WorksheetUUID string `json:"-" schema:"-"`
	StaffID       string `schema:"staff_id"`
	Start         Time   `schema:"start"`
	End           Time   `schema:"end"`
	CommonRequestPayload
}

// VisitCommissionResponse is one commission row on GET /v1/visit-commissions
// and GET /v1/worksheet/{id}/commissions.
// WorksheetUUID is the public worksheet id. The internal BIGINT is not exposed.
type VisitCommissionResponse struct {
	ID                   int64                `json:"id"`
	WorksheetUUID        string               `json:"worksheet_uuid"`
	VisitID              int64                `json:"visit_id"`
	PatientName          string               `json:"patient_name"`
	VisitDate            string               `json:"visit_date"`
	Sources              []ContributionSource `json:"sources"`
	RevenueBase          int64                `json:"revenue_base"`
	CommissionType       null.String          `json:"commission_type"`
	CommissionPercent    null.Float64         `json:"commission_percent"`
	CommissionFlatAmount null.Int64           `json:"commission_flat_amount"`
	CommissionAmount     null.Int64           `json:"commission_amount"`
	HasContributors      bool                 `json:"has_contributors"`
	ApprovedAt           null.Time            `json:"approved_at"`
}

// ListVisitCommissionsResponse is the body for GET /v1/visit-commissions.
type ListVisitCommissionsResponse struct {
	Commissions []VisitCommissionResponse `json:"commissions"`
	Total       int                       `json:"total"`
}

// ArchiveVisitCommissionResponse is the body for DELETE /v1/visit-commissions/{id}.
type ArchiveVisitCommissionResponse struct {
	Success bool `json:"success"`
}
