package model

import (
	"time"

	"github.com/volatiletech/null/v8"
)

const (
	TRX_DIAGNOSIS_TABLE = "mdl_trx_diagnosis"
)

// TrxDiagnosis is the persisted diagnosis record for a visit.
type TrxDiagnosis struct {
	ID                   int64      `xorm:"'id' pk autoincr" json:"id"`
	VisitID              int64      `xorm:"'visit_id'" json:"visit_id"`
	InstitutionID        int64      `xorm:"'institution_id'" json:"institution_id"`
	DoctorID             string     `xorm:"'doctor_id'" json:"doctor_id"`
	ICD10Code            string     `xorm:"'icd10_code'" json:"icd10_code"`
	Rank                 int16      `xorm:"'rank'" json:"rank"`
	Type                 string     `xorm:"'type'" json:"type"`
	Case                 string     `xorm:"'case'" json:"case"`
	ClinicalStatus       string     `xorm:"'clinical_status'" json:"clinical_status"`
	VerificationStatus   string     `xorm:"'verification_status'" json:"verification_status"`
	Prognosis            string     `xorm:"'prognosis'" json:"prognosis"`
	Note                 null.String `xorm:"'note' null" json:"note"`
	OnsetDate            *time.Time  `xorm:"'onset_date' null" json:"onset_date,omitempty"`
	SatuSehatConditionID null.String `xorm:"'satusehat_condition_id' null" json:"satusehat_condition_id"`
	DeletedAt            *time.Time `xorm:"'deleted_at' null" json:"deleted_at,omitempty"`
	CreatedAt            time.Time  `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt            time.Time  `xorm:"'updated_at' updated" json:"updated_at"`
}

// TrxDiagnosisWithDoctor is the join result used in GET /v1/visit/:visit_id/diagnosis.
// Doctor/ICD-10 fields come from LEFT JOINs, so they are null.String to survive
// missing reference rows without a scan error.
type TrxDiagnosisWithDoctor struct {
	TrxDiagnosis  `xorm:"extends"`
	DoctorName    null.String `xorm:"'doctor_name' null" json:"doctor_name"`
	ICD10Display  null.String `xorm:"'icd10_display' null" json:"icd10_display"`
	ICD10Category null.String `xorm:"'icd10_category' null" json:"icd10_category"`
}

// TrxDiagnosisHistory is used for paginated history queries across visits.
type TrxDiagnosisHistory struct {
	TrxDiagnosisWithDoctor `xorm:"extends"`
	VisitDate              time.Time `xorm:"'visit_date'" json:"visit_date"`
}

// Allowed enum values — validated at the usecase layer before DB write.
const (
	DiagnosisTypePrimary     = "primary"
	DiagnosisTypeSecondary   = "secondary"
	DiagnosisTypeComorbidity = "comorbidity"

	DiagnosisCaseNew            = "new"
	DiagnosisCaseChronic        = "chronic"
	DiagnosisCaseAcuteOnChronic = "acute_on_chronic"

	ClinicalStatusActive     = "active"
	ClinicalStatusRecurrence = "recurrence"
	ClinicalStatusRelapse    = "relapse"
	ClinicalStatusInactive   = "inactive"
	ClinicalStatusRemission  = "remission"
	ClinicalStatusResolved   = "resolved"

	VerificationStatusUnconfirmed    = "unconfirmed"
	VerificationStatusProvisional    = "provisional"
	VerificationStatusDifferential   = "differential"
	VerificationStatusConfirmed      = "confirmed"
	VerificationStatusRefuted        = "refuted"
	VerificationStatusEnteredInError = "entered_in_error"

	PrognosisSanam        = "sanam"
	PrognosisBonam        = "bonam"
	PrognosisMalam        = "malam"
	PrognosisDubiaAdSanam = "dubia_ad_sanam"
	PrognosisDubiaAdMalam = "dubia_ad_malam"
)
