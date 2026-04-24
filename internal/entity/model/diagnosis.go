package model

import "time"

const (
	TRX_DIAGNOSIS_TABLE = "mdl_trx_diagnosis"
)

// TrxDiagnosis is the persisted diagnosis record for a visit.
type TrxDiagnosis struct {
	ID                   string     `xorm:"'id' pk" json:"id"`
	VisitID              int64      `xorm:"'visit_id'" json:"visit_id"`
	InstitutionID        int64      `xorm:"'institution_id'" json:"institution_id"`
	DoctorID             string     `xorm:"'doctor_id'" json:"doctor_id"`
	ICD10Code            string     `xorm:"'icd10_code'" json:"icd10_code"`
	Type                 string     `xorm:"'type'" json:"type"`
	Case                 string     `xorm:"'case'" json:"case"`
	ClinicalStatus       string     `xorm:"'clinical_status'" json:"clinical_status"`
	VerificationStatus   string     `xorm:"'verification_status'" json:"verification_status"`
	Prognosis            string     `xorm:"'prognosis'" json:"prognosis"`
	Note                 string     `xorm:"'note'" json:"note,omitempty"`
	OnsetDate            *time.Time `xorm:"'onset_date' null" json:"onset_date,omitempty"`
	SatuSehatConditionID string     `xorm:"'satusehat_condition_id'" json:"satusehat_condition_id,omitempty"`
	DeletedAt            *time.Time `xorm:"'deleted_at' null" json:"deleted_at,omitempty"`
	CreatedAt            time.Time  `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt            time.Time  `xorm:"'updated_at' updated" json:"updated_at"`
}

// TrxDiagnosisWithDoctor is the join result used in GET /v1/visit/:visit_id/diagnosis.
type TrxDiagnosisWithDoctor struct {
	TrxDiagnosis  `xorm:"extends"`
	DoctorName    string `xorm:"'doctor_name'" json:"doctor_name"`
	ICD10Display  string `xorm:"'icd10_display'" json:"icd10_display"`
	ICD10Category string `xorm:"'icd10_category'" json:"icd10_category,omitempty"`
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

	PrognosisExcellent = "excellent"
	PrognosisGood      = "good"
	PrognosisFair      = "fair"
	PrognosisPoor      = "poor"
	PrognosisUnknown   = "unknown"
)
