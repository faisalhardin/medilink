package diagnosis

import (
	"context"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/model"
	diagnosisrepo "github.com/faisalhardin/medilink/internal/entity/repo/diagnosis"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix           = "DiagnosisDB."
	WrapMsgGetActive           = WrapErrMsgPrefix + "GetActiveByVisitID"
	WrapMsgGetActiveByVisitIDs = WrapErrMsgPrefix + "GetActiveByVisitIDs"
	WrapMsgSoftDelIDs          = WrapErrMsgPrefix + "SoftDeleteByIDs"
	WrapMsgSoftDelOne = WrapErrMsgPrefix + "SoftDeleteByID"
	WrapMsgBulkInsert = WrapErrMsgPrefix + "BulkInsert"
	WrapMsgBulkUpdate = WrapErrMsgPrefix + "BulkUpdate"
)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewDiagnosisDB returns a DiagnosisDB implementation bound to the xorm connection.
func NewDiagnosisDB(db *xormlib.DBConnect) diagnosisrepo.DiagnosisDB {
	return &Conn{DB: db}
}

// writeSession returns the active TX session if one was put on ctx by the
// usecase; otherwise it returns a fresh master-engine session. Mutating repo
// methods MUST use this helper so diagnosis + outbox enqueue can land atomically.
func (c *Conn) writeSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

// GetActiveByVisitID returns the read-model used by GET /v1/visit/:visit_id/diagnosis.
// icd10_display lives on the diagnosis row as a write-time snapshot, so the
// only join we still need is the doctor master (for doctor_name).
// Uses the slave DB because reads never participate in write transactions.
func (c *Conn) GetActiveByVisitID(ctx context.Context, institutionID, visitID int64) ([]model.TrxDiagnosisWithDoctor, error) {
	const sql = `
		SELECT
			d.id, d.visit_id, d.institution_id, d.doctor_id,
			d.icd10_code, d.icd10_display, d.rank,
			d.type, d."case", d.clinical_status, d.verification_status,
			d.prognosis, d.note, d.onset_date, d.satusehat_condition_id,
			d.deleted_at, d.created_at, d.updated_at,
			md.name AS doctor_name
		FROM mdl_trx_diagnosis d
		LEFT JOIN mdl_mst_doctor md ON md.id = d.doctor_id
		WHERE d.institution_id = ?
		  AND d.visit_id = ?
		  AND d.deleted_at IS NULL
		ORDER BY d.created_at ASC
	`

	var rows []model.TrxDiagnosisWithDoctor
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, visitID).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetActive)
	}
	return rows, nil
}

// GetActiveByVisitIDs returns diagnoses for all given visits in one round-trip.
func (c *Conn) GetActiveByVisitIDs(ctx context.Context, institutionID int64, visitIDs []int64) ([]model.TrxDiagnosisWithDoctor, error) {
	if len(visitIDs) == 0 {
		return nil, nil
	}

	args := make([]interface{}, 0, len(visitIDs)+1)
	args = append(args, institutionID)
	for _, id := range visitIDs {
		args = append(args, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(visitIDs)), ",")
	sql := `
		SELECT
			d.id, d.visit_id, d.institution_id, d.doctor_id,
			d.icd10_code, d.icd10_display, d.rank,
			d.type, d."case", d.clinical_status, d.verification_status,
			d.prognosis, d.note, d.onset_date, d.satusehat_condition_id,
			d.deleted_at, d.created_at, d.updated_at,
			md.name AS doctor_name
		FROM mdl_trx_diagnosis d
		LEFT JOIN mdl_mst_doctor md ON md.id = d.doctor_id
		WHERE d.institution_id = ?
		  AND d.visit_id IN (` + placeholders + `)
		  AND d.deleted_at IS NULL
		ORDER BY d.visit_id, d.created_at ASC
	`

	var rows []model.TrxDiagnosisWithDoctor
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, args...).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetActiveByVisitIDs)
	}
	return rows, nil
}

// SoftDeleteByIDs flips deleted_at = NOW() for a batch of ids, scoped by
// (institution_id, visit_id). Rows already soft-deleted are filtered by the
// partial WHERE so this is idempotent.
func (c *Conn) SoftDeleteByIDs(ctx context.Context, institutionID, visitID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(ids)+2)
	args = append(args, institutionID, visitID)
	for _, id := range ids {
		args = append(args, id)
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	sql := `
		UPDATE mdl_trx_diagnosis
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE institution_id = ?
		  AND visit_id = ?
		  AND deleted_at IS NULL
		  AND id IN (` + placeholders + `)
	`

	_, err := c.writeSession(ctx).Exec(append([]interface{}{sql}, args...)...)
	if err != nil {
		return errors.Wrap(err, WrapMsgSoftDelIDs)
	}
	return nil
}

// SoftDeleteByID is the single-row variant used by the DELETE endpoint.
// Returns found=false for idempotency when the row is missing or already deleted.
func (c *Conn) SoftDeleteByID(ctx context.Context, institutionID, visitID int64, diagnosisID int64) (bool, error) {
	const sql = `
		UPDATE mdl_trx_diagnosis
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE institution_id = ?
		  AND visit_id = ?
		  AND id = ?
		  AND deleted_at IS NULL
	`

	res, err := c.writeSession(ctx).Exec(sql, institutionID, visitID, diagnosisID)
	if err != nil {
		return false, errors.Wrap(err, WrapMsgSoftDelOne)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, errors.Wrap(err, WrapMsgSoftDelOne)
	}
	return affected > 0, nil
}

// `case` is a SQL reserved word — we always double-quote the column.
func (c *Conn) BulkInsert(ctx context.Context, rows []model.TrxDiagnosis) error {
	if len(rows) == 0 {
		return nil
	}

	const perRowCols = 13
	placeholders := make([]string, 0, len(rows))
	args := make([]interface{}, 0, len(rows)*perRowCols)

	for i := range rows {
		r := &rows[i]
		placeholders = append(placeholders,
			"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		)
		args = append(args,
			r.VisitID,
			r.InstitutionID,
			r.DoctorID,
			r.ICD10Code,
			r.ICD10Display,
			r.Rank,
			r.Type,
			r.Case,
			r.ClinicalStatus,
			r.VerificationStatus,
			r.Prognosis,
			r.Note,
			r.OnsetDate,
		)
	}

	sql := `
		INSERT INTO mdl_trx_diagnosis
		(visit_id, institution_id, doctor_id, icd10_code, icd10_display, rank,
		 type, "case", clinical_status, verification_status, prognosis,
		 note, onset_date, created_at, updated_at)
		VALUES ` + strings.Join(placeholders, ", ")

	_, err := c.writeSession(ctx).Exec(append([]interface{}{sql}, args...)...)
	if err != nil {
		return errors.Wrap(err, WrapMsgBulkInsert)
	}
	return nil
}

// BulkUpdate overwrites the mutable columns for each row. We issue one UPDATE
// per row rather than a bulk CTE for readability and to keep the error path
// obvious in the usecase; diagnosis writes are low-cardinality per visit.
func (c *Conn) BulkUpdate(ctx context.Context, rows []model.TrxDiagnosis) error {
	if len(rows) == 0 {
		return nil
	}

	const sql = `
		UPDATE mdl_trx_diagnosis
		SET icd10_code          = ?,
		    icd10_display       = ?,
		    rank                = ?,
		    type                = ?,
		    "case"              = ?,
		    clinical_status     = ?,
		    verification_status = ?,
		    prognosis           = ?,
		    note                = ?,
		    onset_date          = ?,
		    updated_at          = NOW()
		WHERE institution_id = ?
		  AND visit_id       = ?
		  AND id             = ?
		  AND deleted_at IS NULL
	`

	session := c.writeSession(ctx)
	for i := range rows {
		r := &rows[i]
		_, err := session.Exec(sql,
			r.ICD10Code,
			r.ICD10Display,
			r.Rank,
			r.Type,
			r.Case,
			r.ClinicalStatus,
			r.VerificationStatus,
			r.Prognosis,
			r.Note,
			r.OnsetDate,
			r.InstitutionID,
			r.VisitID,
			r.ID,
		)
		if err != nil {
			return errors.Wrap(err, WrapMsgBulkUpdate)
		}
	}
	return nil
}
