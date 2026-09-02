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

// TrxVisitCommission is a per-visit, per-staff commission row within a payday period.
type TrxVisitCommission struct {
	ID                   int64           `xorm:"'id' pk autoincr" json:"-"`
	PeriodID             int64           `xorm:"'period_id'" json:"-"`
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

// ─── JSON request / response DTOs ─────────────────────────────────────────────

// CreateCompensationPeriodRequest is the body for POST /v1/compensation/periods.
type CreateCompensationPeriodRequest struct {
	Label       string `json:"label"`
	PeriodStart Time   `json:"period_start"`
	PeriodEnd   Time   `json:"period_end"`
}

// ListCompensationPeriodsRequest is the query for GET /v1/compensation/periods.
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
