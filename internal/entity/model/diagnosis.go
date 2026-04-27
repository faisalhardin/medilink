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
	ID                   int64       `xorm:"'id' pk autoincr" json:"id"`
	VisitID              int64       `xorm:"'visit_id'" json:"visit_id"`
	InstitutionID        int64       `xorm:"'institution_id'" json:"institution_id"`
	DoctorID             string      `xorm:"'doctor_id'" json:"doctor_id"`
	ICD10Code            string      `xorm:"'icd10_code'" json:"icd10_code"`
	Rank                 int16       `xorm:"'rank'" json:"rank"`
	Type                 string      `xorm:"'type'" json:"type"`
	Case                 string      `xorm:"'case'" json:"case"`
	ClinicalStatus       string      `xorm:"'clinical_status'" json:"clinical_status"`
	VerificationStatus   string      `xorm:"'verification_status'" json:"verification_status"`
	Prognosis            string      `xorm:"'prognosis'" json:"prognosis"`
	Note                 null.String `xorm:"'note' null" json:"note"`
	OnsetDate            *time.Time  `xorm:"'onset_date' null" json:"onset_date,omitempty"`
	SatuSehatConditionID null.String `xorm:"'satusehat_condition_id' null" json:"satusehat_condition_id"`
	DeletedAt            *time.Time  `xorm:"'deleted_at' null" json:"deleted_at,omitempty"`
	CreatedAt            time.Time   `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt            time.Time   `xorm:"'updated_at' updated" json:"updated_at"`
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

// SaveDiagnosesRequest is the payload for POST /v1/visit/:visit_id/diagnosis.
// prognosis is visit-level in the API contract and will be copied to each row.
type SaveDiagnosesRequest struct {
	Diagnoses []SaveDiagnosisItem `json:"diagnoses" validate:"required,min=1,dive"`
	Prognosis string              `json:"prognosis" validate:"omitempty,oneof=sanam bonam malam dubia_ad_sanam dubia_ad_malam"`
}

type SaveDiagnosisItem struct {
	ID                 *int64  `json:"id"`
	ICD10Code          string  `json:"icd10_code" validate:"required,min=1,max=10"`
	Type               string  `json:"type" validate:"required,oneof=primary secondary comorbidity"`
	Case               string  `json:"case" validate:"required,oneof=new chronic acute_on_chronic"`
	ClinicalStatus     string  `json:"clinical_status" validate:"required,oneof=active recurrence relapse inactive remission resolved"`
	VerificationStatus string  `json:"verification_status" validate:"required,oneof=unconfirmed provisional differential confirmed refuted entered_in_error"`
	OnsetDate          Time    `json:"onset_date" validate:"omitempty"`
	DoctorID           string  `json:"doctor_id" validate:"required,uuid4"`
	Rank               int16   `json:"rank" validate:"required,min=1"`
	Note               *string `json:"note"`
}

type SaveDiagnosesResponse struct {
	Saved   int `json:"saved"`
	Deleted int `json:"deleted"`
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
