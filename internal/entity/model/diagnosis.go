package model

import (
	"time"

	customtime "github.com/faisalhardin/medilink/pkg/type/time"

	"github.com/volatiletech/null/v8"
)

const (
	TRX_DIAGNOSIS_TABLE = "mdl_trx_diagnosis"
)

// TrxDiagnosis is the persisted diagnosis record for a visit.
// icd10_display is a snapshot of ref_icd10.display captured at write time;
// the application treats the reference text as immutable, which lets the
// read path skip the ref_icd10 join entirely.
type TrxDiagnosis struct {
	ID                   int64       `xorm:"'id' pk autoincr" json:"-"`
	VisitID              int64       `xorm:"'visit_id'" json:"-"`
	InstitutionID        int64       `xorm:"'institution_id'" json:"-"`
	DoctorID             string      `xorm:"'doctor_id'" json:"-"`
	ICD10Code            string      `xorm:"'icd10_code'" json:"-"`
	ICD10Display         string      `xorm:"'icd10_display'" json:"-"`
	Rank                 int16       `xorm:"'rank'" json:"-"`
	Type                 string      `xorm:"'type'" json:"-"`
	Case                 string      `xorm:"'case'" json:"-"`
	ClinicalStatus       string      `xorm:"'clinical_status'" json:"-"`
	VerificationStatus   string      `xorm:"'verification_status'" json:"-"`
	Prognosis            string      `xorm:"'prognosis'" json:"-"`
	Note                 null.String `xorm:"'note' null" json:"-"`
	OnsetDate            *time.Time  `xorm:"'onset_date' null" json:"-"`
	SatuSehatConditionID null.String `xorm:"'satusehat_condition_id' null" json:"-"`
	DeletedAt            *time.Time  `xorm:"'deleted_at' null" json:"-"`
	CreatedAt            time.Time   `xorm:"'created_at' created" json:"-"`
	UpdatedAt            time.Time   `xorm:"'updated_at' updated" json:"-"`
}

// TrxDiagnosisWithDoctor is the read-model used in GET /v1/visit/:visit_id/diagnosis.
// Only doctor_name still needs a join; icd10_display is read directly from the
// snapshot column on the embedded TrxDiagnosis.
type TrxDiagnosisWithDoctor struct {
	ID                   int64       `xorm:"'id' pk autoincr"`
	VisitID              int64       `xorm:"'visit_id'"`
	InstitutionID        int64       `xorm:"'institution_id'"`
	DoctorID             string      `xorm:"'doctor_id'"`
	ICD10Code            string      `xorm:"'icd10_code'"`
	ICD10Display         string      `xorm:"'icd10_display'"`
	Rank                 int16       `xorm:"'rank'"`
	Type                 string      `xorm:"'type'"`
	Case                 string      `xorm:"'case'"`
	ClinicalStatus       string      `xorm:"'clinical_status'"`
	VerificationStatus   string      `xorm:"'verification_status'"`
	Prognosis            string      `xorm:"'prognosis'"`
	Note                 null.String `xorm:"'note' null"`
	OnsetDate            *time.Time  `xorm:"'onset_date' null"`
	SatuSehatConditionID null.String `xorm:"'satusehat_condition_id' null"`
	DeletedAt            *time.Time  `xorm:"'deleted_at' null"`
	CreatedAt            time.Time   `xorm:"'created_at'"`
	UpdatedAt            time.Time   `xorm:"'updated_at'"`
	DoctorName           null.String `xorm:"'doctor_name' null"`
}

// AsTrxDiagnosis extracts the mutable write-side model from the read model.
// Used by the diff logic in the diagnosis usecase.
func (r TrxDiagnosisWithDoctor) AsTrxDiagnosis() TrxDiagnosis {
	return TrxDiagnosis{
		ID:                   r.ID,
		VisitID:              r.VisitID,
		InstitutionID:        r.InstitutionID,
		DoctorID:             r.DoctorID,
		ICD10Code:            r.ICD10Code,
		ICD10Display:         r.ICD10Display,
		Rank:                 r.Rank,
		Type:                 r.Type,
		Case:                 r.Case,
		ClinicalStatus:       r.ClinicalStatus,
		VerificationStatus:   r.VerificationStatus,
		Prognosis:            r.Prognosis,
		Note:                 r.Note,
		OnsetDate:            r.OnsetDate,
		SatuSehatConditionID: r.SatuSehatConditionID,
		DeletedAt:            r.DeletedAt,
		CreatedAt:            r.CreatedAt,
		UpdatedAt:            r.UpdatedAt,
	}
}

// ToResponse converts the DB read-model to the JSON response DTO.
func (r TrxDiagnosisWithDoctor) ToResponse() DiagnosisResponse {
	res := DiagnosisResponse{
		ID:                   r.ID,
		VisitID:              r.VisitID,
		InstitutionID:        r.InstitutionID,
		DoctorID:             r.DoctorID,
		ICD10Code:            r.ICD10Code,
		ICD10Display:         r.ICD10Display,
		Rank:                 r.Rank,
		Type:                 r.Type,
		Case:                 r.Case,
		ClinicalStatus:       r.ClinicalStatus,
		VerificationStatus:   r.VerificationStatus,
		Prognosis:            r.Prognosis,
		Note:                 r.Note,
		SatuSehatConditionID: r.SatuSehatConditionID,
		DoctorName:           r.DoctorName,
		CreatedAt:            customtime.Time{Time: r.CreatedAt},
		UpdatedAt:            customtime.Time{Time: r.UpdatedAt},
	}
	if r.OnsetDate != nil {
		ct := customtime.Time{Time: *r.OnsetDate}
		res.OnsetDate = &ct
	}
	if r.DeletedAt != nil {
		ct := customtime.Time{Time: *r.DeletedAt}
		res.DeletedAt = &ct
	}
	return res
}

// TrxDiagnosisHistory is used for paginated history queries across visits.
// Also flattened for the same reason as TrxDiagnosisWithDoctor.
type TrxDiagnosisHistory struct {
	ID                   int64       `xorm:"'id'"`
	VisitID              int64       `xorm:"'visit_id'"`
	InstitutionID        int64       `xorm:"'institution_id'"`
	DoctorID             string      `xorm:"'doctor_id'"`
	ICD10Code            string      `xorm:"'icd10_code'"`
	ICD10Display         string      `xorm:"'icd10_display'"`
	Rank                 int16       `xorm:"'rank'"`
	Type                 string      `xorm:"'type'"`
	Case                 string      `xorm:"'case'"`
	ClinicalStatus       string      `xorm:"'clinical_status'"`
	VerificationStatus   string      `xorm:"'verification_status'"`
	Prognosis            string      `xorm:"'prognosis'"`
	Note                 null.String `xorm:"'note' null"`
	OnsetDate            *time.Time  `xorm:"'onset_date' null"`
	SatuSehatConditionID null.String `xorm:"'satusehat_condition_id' null"`
	DeletedAt            *time.Time  `xorm:"'deleted_at' null"`
	CreatedAt            time.Time   `xorm:"'created_at'"`
	UpdatedAt            time.Time   `xorm:"'updated_at'"`
	DoctorName           null.String `xorm:"'doctor_name' null"`
	VisitDate            time.Time   `xorm:"'visit_date'"`
}

// DiagnosisResponse is the JSON envelope for GET /v1/visit/:visit_id/diagnosis.
type DiagnosisResponse struct {
	ID                   int64            `json:"id"`
	VisitID              int64            `json:"visit_id"`
	InstitutionID        int64            `json:"institution_id"`
	DoctorID             string           `json:"doctor_id"`
	ICD10Code            string           `json:"icd10_code"`
	ICD10Display         string           `json:"icd10_display"`
	Rank                 int16            `json:"rank"`
	Type                 string           `json:"type"`
	Case                 string           `json:"case"`
	ClinicalStatus       string           `json:"clinical_status"`
	VerificationStatus   string           `json:"verification_status"`
	Prognosis            string           `json:"prognosis"`
	Note                 null.String      `json:"note"`
	OnsetDate            *customtime.Time `json:"onset_date,omitempty"`
	SatuSehatConditionID null.String      `json:"satusehat_condition_id"`
	DeletedAt            *customtime.Time `json:"deleted_at,omitempty"`
	CreatedAt            customtime.Time  `json:"created_at"`
	UpdatedAt            customtime.Time  `json:"updated_at"`
	DoctorName           null.String      `json:"doctor_name"`
}

// ─── Request DTOs (JSON) ─────────────────────────────────────────────────────

// SaveDiagnosesRequest is the payload for POST /v1/visit/:visit_id/diagnosis.
// prognosis is visit-level in the API contract and will be copied to each row.
type SaveDiagnosesRequest struct {
	Diagnoses []SaveDiagnosisItem `json:"diagnoses" validate:"required,min=1,dive"`
	Prognosis string              `json:"prognosis" validate:"omitempty,oneof=sanam bonam malam dubia_ad_sanam dubia_ad_malam"`
}

type SaveDiagnosisItem struct {
	ID                 *int64           `json:"id"`
	ICD10Code          string           `json:"icd10_code" validate:"required,min=1,max=10"`
	Type               string           `json:"type" validate:"required,oneof=primary secondary comorbidity"`
	Case               string           `json:"case" validate:"required,oneof=new chronic acute_on_chronic"`
	ClinicalStatus     string           `json:"clinical_status" validate:"required,oneof=active recurrence relapse inactive remission resolved"`
	VerificationStatus string           `json:"verification_status" validate:"required,oneof=unconfirmed provisional differential confirmed refuted entered_in_error"`
	OnsetDate          *customtime.Time `json:"onset_date" validate:"omitempty"`
	DoctorID           string           `json:"doctor_id" validate:"required,uuid4"`
	Rank               int16            `json:"rank" validate:"required,min=1"`
	Note               *string          `json:"note"`
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
