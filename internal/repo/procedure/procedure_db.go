package procedure

import (
	"context"
	"fmt"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/model"
	procedurerepo "github.com/faisalhardin/medilink/internal/entity/repo/procedure"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix    = "ProcedureDB."
	WrapMsgGetActive    = WrapErrMsgPrefix + "GetActiveByVisitID"
	WrapMsgBulkInsert   = WrapErrMsgPrefix + "BulkInsert"
	WrapMsgBulkUpdate   = WrapErrMsgPrefix + "BulkUpdate"
	WrapMsgSoftDelIDs   = WrapErrMsgPrefix + "SoftDeleteByIDs"
	WrapMsgSoftDelOne   = WrapErrMsgPrefix + "SoftDeleteByID"
	WrapMsgHistory      = WrapErrMsgPrefix + "GetPatientProcedureHistory"
	WrapMsgHistoryCount = WrapErrMsgPrefix + "GetPatientProcedureHistoryCount"
)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewProcedureDB returns a ProcedureDB implementation bound to the xorm connection.
func NewProcedureDB(db *xormlib.DBConnect) procedurerepo.ProcedureDB {
	return &Conn{DB: db}
}

// writeSession returns the active TX session if one was put on ctx by the
// usecase; otherwise returns a fresh master-engine session.
func (c *Conn) writeSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

// GetActiveByVisitID returns all non-soft-deleted procedure rows for a visit,
// ordered by rank ASC. Snapshots are stored on the row so no JOINs are needed.
func (c *Conn) GetActiveByVisitID(ctx context.Context, institutionID, visitID int64) ([]model.TrxVisitProcedure, error) {
	const sql = `
		SELECT id, visit_id, institution_id, product_id, product_name,
		       doctor_id, doctor_name, nurse_id, nurse_name,
		       planned_at, category, duration, icd9cm_code, icd9cm_display,
		       description, notes, rank,
		       created_at, updated_at, deleted_at
		FROM mdl_trx_visit_procedure
		WHERE institution_id = ?
		  AND visit_id = ?
		  AND deleted_at IS NULL
		ORDER BY rank ASC
	`

	var rows []model.TrxVisitProcedure
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, visitID).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetActive)
	}
	return rows, nil
}

// BulkInsert persists a batch of new procedure rows.
func (c *Conn) BulkInsert(ctx context.Context, rows []model.TrxVisitProcedure) error {
	if len(rows) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(rows))
	args := make([]interface{}, 0, len(rows)*16)

	for i := range rows {
		r := &rows[i]
		placeholders = append(placeholders,
			"(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())",
		)
		args = append(args,
			r.VisitID, r.InstitutionID,
			r.ProductID, r.ProductName,
			r.DoctorID, r.DoctorName,
			r.NurseID, r.NurseName,
			r.PlannedAt, r.Category, r.Duration,
			r.ICD9CMCode, r.ICD9CMDisplay,
			r.Description, r.Notes, r.Rank,
		)
	}

	sql := `
		INSERT INTO mdl_trx_visit_procedure
		(visit_id, institution_id, product_id, product_name,
		 doctor_id, doctor_name, nurse_id, nurse_name,
		 planned_at, category, duration, icd9cm_code, icd9cm_display,
		 description, notes, rank,
		 created_at, updated_at)
		VALUES ` + strings.Join(placeholders, ", ")

	_, err := c.writeSession(ctx).Exec(append([]interface{}{sql}, args...)...)
	if err != nil {
		return errors.Wrap(err, WrapMsgBulkInsert)
	}
	return nil
}

// BulkUpdate overwrites the mutable columns for each row.
func (c *Conn) BulkUpdate(ctx context.Context, rows []model.TrxVisitProcedure) error {
	if len(rows) == 0 {
		return nil
	}

	const sql = `
		UPDATE mdl_trx_visit_procedure
		SET product_id        = ?,
		    product_name      = ?,
		    doctor_id         = ?,
		    doctor_name       = ?,
		    nurse_id          = ?,
		    nurse_name        = ?,
		    planned_at        = ?,
		    category          = ?,
		    duration          = ?,
		    icd9cm_code       = ?,
		    icd9cm_display    = ?,
		    description       = ?,
		    notes             = ?,
		    rank              = ?,
		    updated_at        = NOW()
		WHERE institution_id = ?
		  AND visit_id       = ?
		  AND id             = ?
		  AND deleted_at IS NULL
	`

	session := c.writeSession(ctx)
	for i := range rows {
		r := &rows[i]
		_, err := session.Exec(sql,
			r.ProductID, r.ProductName,
			r.DoctorID, r.DoctorName,
			r.NurseID, r.NurseName,
			r.PlannedAt, r.Category, r.Duration,
			r.ICD9CMCode, r.ICD9CMDisplay,
			r.Description, r.Notes, r.Rank,
			r.InstitutionID, r.VisitID, r.ID,
		)
		if err != nil {
			return errors.Wrap(err, WrapMsgBulkUpdate)
		}
	}
	return nil
}

// SoftDeleteByIDs marks a batch of rows deleted, scoped by institution + visit.
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
		UPDATE mdl_trx_visit_procedure
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE institution_id = ?
		  AND visit_id        = ?
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
func (c *Conn) SoftDeleteByID(ctx context.Context, institutionID, visitID, procedureID int64) (bool, error) {
	const sql = `
		UPDATE mdl_trx_visit_procedure
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE institution_id = ?
		  AND visit_id        = ?
		  AND id              = ?
		  AND deleted_at IS NULL
	`

	res, err := c.writeSession(ctx).Exec(sql, institutionID, visitID, procedureID)
	if err != nil {
		return false, errors.Wrap(err, WrapMsgSoftDelOne)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, errors.Wrap(err, WrapMsgSoftDelOne)
	}
	return affected > 0, nil
}

// GetPatientProcedureHistory returns paginated cross-visit procedure records
// for a patient, ordered by visit date descending. Institution-scoped for tenant safety.
func (c *Conn) GetPatientProcedureHistory(ctx context.Context, institutionID int64, patientUUID string, limit, offset int) ([]model.ProcedureHistoryEntry, int, error) {
	const baseSQL = `
		FROM mdl_trx_visit_procedure p
		JOIN mdl_trx_patient_visit pv ON pv.id = p.visit_id
		JOIN mdl_mst_patient_institution pat ON pat.id = pv.id_mst_patient
		WHERE pat.uuid = ?
		  AND p.institution_id = ?
		  AND p.deleted_at IS NULL
	`

	var total int
	countSQL := "SELECT COUNT(*) " + baseSQL
	_, err := c.DB.SlaveDB.Context(ctx).SQL(countSQL, patientUUID, institutionID).Get(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, WrapMsgHistoryCount)
	}

	dataSQL := fmt.Sprintf(`
		SELECT p.visit_id,
		       TO_CHAR(pv.create_time AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS visit_date,
		       p.product_name,
		       p.icd9cm_code,
		       p.icd9cm_display,
		       p.doctor_name
		%s
		ORDER BY pv.create_time DESC
		LIMIT ? OFFSET ?
	`, baseSQL)

	var rows []model.ProcedureHistoryEntry
	err = c.DB.SlaveDB.Context(ctx).SQL(dataSQL, patientUUID, institutionID, limit, offset).Find(&rows)
	if err != nil {
		return nil, 0, errors.Wrap(err, WrapMsgHistory)
	}

	return rows, total, nil
}
