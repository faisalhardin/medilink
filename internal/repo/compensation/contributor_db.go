package compensation

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	compensationrepo "github.com/faisalhardin/medilink/internal/entity/repo/compensation"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	wrapMsgDetectForVisit          = "ContributorDB.DetectForVisit"
	wrapMsgDetectForVisits         = "ContributorDB.DetectForVisits"
	wrapMsgDetectStaffForPeriod    = "ContributorDB.DetectStaffForPeriod"
	wrapMsgDetectForPeriodStaff    = "ContributorDB.DetectForPeriodStaff"
	wrapMsgUpsertManualContributor = "ContributorDB.UpsertManualContributor"
	wrapMsgDeleteManualContributor = "ContributorDB.DeleteManualContributor"

	// detect*SQL use visit_id = ANY(?) so DetectForVisit and DetectForVisits share one shape.
	// Single-visit callers pass pq.Array([]int64{visitID}).

	detectProcedureSQL = `
		SELECT
			'procedure' AS source_type,
			p.visit_id AS visit_id,
			md.staff_uuid::text AS staff_id,
			s.name AS name,
			p.id AS clinical_id,
			p.id AS procedure_id,
			NULL::bigint AS diagnosis_id,
			p.product_id AS product_id,
			p.product_name AS label
		FROM mdl_trx_visit_procedure p
		INNER JOIN mdl_mst_doctor md
			ON md.id = p.doctor_id
			AND md.staff_uuid IS NOT NULL
		INNER JOIN mdl_mst_staff s
			ON s.uuid = md.staff_uuid::text
			AND s.delete_time IS NULL
		WHERE p.institution_id = ?
		  AND p.visit_id = ANY(?)
		  AND p.deleted_at IS NULL

		UNION ALL

		SELECT
			'procedure' AS source_type,
			p.visit_id AS visit_id,
			mn.staff_uuid::text AS staff_id,
			s.name AS name,
			p.id AS clinical_id,
			p.id AS procedure_id,
			NULL::bigint AS diagnosis_id,
			p.product_id AS product_id,
			p.product_name AS label
		FROM mdl_trx_visit_procedure p
		INNER JOIN mdl_mst_nurse mn
			ON mn.id = p.nurse_id
			AND mn.staff_uuid IS NOT NULL
		INNER JOIN mdl_mst_staff s
			ON s.uuid = mn.staff_uuid::text
			AND s.delete_time IS NULL
		WHERE p.institution_id = ?
		  AND p.visit_id = ANY(?)
		  AND p.deleted_at IS NULL
		  AND p.nurse_id IS NOT NULL
	`

	detectDiagnosisSQL = `
		SELECT
			'diagnosis' AS source_type,
			d.visit_id AS visit_id,
			md.staff_uuid::text AS staff_id,
			s.name AS name,
			d.id AS clinical_id,
			NULL::bigint AS procedure_id,
			d.id AS diagnosis_id,
			NULL::bigint AS product_id,
			d.icd10_display AS label
		FROM mdl_trx_diagnosis d
		INNER JOIN mdl_mst_doctor md
			ON md.id = d.doctor_id
			AND md.staff_uuid IS NOT NULL
		INNER JOIN mdl_mst_staff s
			ON s.uuid = md.staff_uuid::text
			AND s.delete_time IS NULL
		WHERE d.institution_id = ?
		  AND d.visit_id = ANY(?)
		  AND d.deleted_at IS NULL
	`

	detectAnamnesaSQL = `
		SELECT
			'anamnesa' AS source_type,
			a.visit_id AS visit_id,
			md.staff_uuid::text AS staff_id,
			s.name AS name,
			0::bigint AS clinical_id,
			NULL::bigint AS procedure_id,
			NULL::bigint AS diagnosis_id,
			NULL::bigint AS product_id,
			NULL::text AS label
		FROM mdl_trx_anamnesa a
		INNER JOIN mdl_mst_doctor md
			ON md.id = a.doctor_id
			AND md.staff_uuid IS NOT NULL
		INNER JOIN mdl_mst_staff s
			ON s.uuid = md.staff_uuid::text
			AND s.delete_time IS NULL
		WHERE a.institution_id = ?
		  AND a.visit_id = ANY(?)
		  AND a.doctor_id IS NOT NULL

		UNION ALL

		SELECT
			'anamnesa' AS source_type,
			a.visit_id AS visit_id,
			mn.staff_uuid::text AS staff_id,
			s.name AS name,
			0::bigint AS clinical_id,
			NULL::bigint AS procedure_id,
			NULL::bigint AS diagnosis_id,
			NULL::bigint AS product_id,
			NULL::text AS label
		FROM mdl_trx_anamnesa a
		INNER JOIN mdl_mst_nurse mn
			ON mn.id = a.nurse_id
			AND mn.staff_uuid IS NOT NULL
		INNER JOIN mdl_mst_staff s
			ON s.uuid = mn.staff_uuid::text
			AND s.delete_time IS NULL
		WHERE a.institution_id = ?
		  AND a.visit_id = ANY(?)
		  AND a.nurse_id IS NOT NULL
	`

	detectMapSQL = `
		SELECT
			'manual' AS source_type,
			m.visit_id AS visit_id,
			m.staff_id::text AS staff_id,
			s.name AS name,
			m.id AS clinical_id,
			NULL::bigint AS procedure_id,
			NULL::bigint AS diagnosis_id,
			NULL::bigint AS product_id,
			NULL::text AS label
		FROM mdl_map_visit_contributor m
		INNER JOIN mdl_mst_staff s
			ON s.uuid = m.staff_id::text
			AND s.delete_time IS NULL
		WHERE m.institution_id = ?
		  AND m.visit_id = ANY(?)
		  AND m.delete_time IS NULL
	`

	// detectForPeriodStaffSQL is the period+staff window with the same source columns as detect*SQL.
	// Staff name is unused by generate, so mdl_mst_staff is not joined.
	detectForPeriodStaffSQL = `
		WITH period_visits AS (
			SELECT id
			FROM mdl_trx_patient_visit
			WHERE id_mst_institution = ?
			  AND delete_time IS NULL
			  AND create_time >= ?
			  AND create_time < ?
		)
		SELECT
			'procedure' AS source_type,
			p.visit_id AS visit_id,
			md.staff_uuid::text AS staff_id,
			''::text AS name,
			p.id AS clinical_id,
			p.id AS procedure_id,
			NULL::bigint AS diagnosis_id,
			p.product_id AS product_id,
			p.product_name AS label
		FROM mdl_trx_visit_procedure p
		INNER JOIN period_visits v ON v.id = p.visit_id
		INNER JOIN mdl_mst_doctor md
			ON md.id = p.doctor_id
			AND md.staff_uuid IS NOT NULL
		WHERE p.institution_id = ?
		  AND p.deleted_at IS NULL
		  AND md.staff_uuid = ?

		UNION ALL

		SELECT
			'procedure' AS source_type,
			p.visit_id AS visit_id,
			mn.staff_uuid::text AS staff_id,
			''::text AS name,
			p.id AS clinical_id,
			p.id AS procedure_id,
			NULL::bigint AS diagnosis_id,
			p.product_id AS product_id,
			p.product_name AS label
		FROM mdl_trx_visit_procedure p
		INNER JOIN period_visits v ON v.id = p.visit_id
		INNER JOIN mdl_mst_nurse mn
			ON mn.id = p.nurse_id
			AND mn.staff_uuid IS NOT NULL
		WHERE p.institution_id = ?
		  AND p.deleted_at IS NULL
		  AND p.nurse_id IS NOT NULL
		  AND mn.staff_uuid = ?

		UNION ALL

		SELECT
			'diagnosis' AS source_type,
			d.visit_id AS visit_id,
			md.staff_uuid::text AS staff_id,
			''::text AS name,
			d.id AS clinical_id,
			NULL::bigint AS procedure_id,
			d.id AS diagnosis_id,
			NULL::bigint AS product_id,
			d.icd10_display AS label
		FROM mdl_trx_diagnosis d
		INNER JOIN period_visits v ON v.id = d.visit_id
		INNER JOIN mdl_mst_doctor md
			ON md.id = d.doctor_id
			AND md.staff_uuid IS NOT NULL
		WHERE d.institution_id = ?
		  AND d.deleted_at IS NULL
		  AND md.staff_uuid = ?

		UNION ALL

		SELECT
			'anamnesa' AS source_type,
			a.visit_id AS visit_id,
			md.staff_uuid::text AS staff_id,
			''::text AS name,
			0::bigint AS clinical_id,
			NULL::bigint AS procedure_id,
			NULL::bigint AS diagnosis_id,
			NULL::bigint AS product_id,
			NULL::text AS label
		FROM mdl_trx_anamnesa a
		INNER JOIN period_visits v ON v.id = a.visit_id
		INNER JOIN mdl_mst_doctor md
			ON md.id = a.doctor_id
			AND md.staff_uuid IS NOT NULL
		WHERE a.institution_id = ?
		  AND a.doctor_id IS NOT NULL
		  AND md.staff_uuid = ?

		UNION ALL

		SELECT
			'anamnesa' AS source_type,
			a.visit_id AS visit_id,
			mn.staff_uuid::text AS staff_id,
			''::text AS name,
			0::bigint AS clinical_id,
			NULL::bigint AS procedure_id,
			NULL::bigint AS diagnosis_id,
			NULL::bigint AS product_id,
			NULL::text AS label
		FROM mdl_trx_anamnesa a
		INNER JOIN period_visits v ON v.id = a.visit_id
		INNER JOIN mdl_mst_nurse mn
			ON mn.id = a.nurse_id
			AND mn.staff_uuid IS NOT NULL
		WHERE a.institution_id = ?
		  AND a.nurse_id IS NOT NULL
		  AND mn.staff_uuid = ?

		UNION ALL

		SELECT
			'manual' AS source_type,
			m.visit_id AS visit_id,
			m.staff_id::text AS staff_id,
			''::text AS name,
			m.id AS clinical_id,
			NULL::bigint AS procedure_id,
			NULL::bigint AS diagnosis_id,
			NULL::bigint AS product_id,
			NULL::text AS label
		FROM mdl_map_visit_contributor m
		INNER JOIN period_visits v ON v.id = m.visit_id
		WHERE m.institution_id = ?
		  AND m.delete_time IS NULL
		  AND m.staff_id = ?
	`

	// detectStaffForPeriodSQL aggregates contributing staff across visits in a payday window.
	// raw emits staff_uuid only (no per-arm staff join) so the planner cannot cross-product
	// clinical rows against all staff; name + roles resolve once after GROUP BY.
	detectStaffForPeriodSQL = `
		WITH period_visits AS (
			SELECT id
			FROM mdl_trx_patient_visit
			WHERE id_mst_institution = ?
			  AND delete_time IS NULL
			  AND create_time >= ?
			  AND create_time < ?
		),
		raw AS (
			SELECT md.staff_uuid AS staff_uuid, p.visit_id AS visit_id
			FROM mdl_trx_visit_procedure p
			INNER JOIN period_visits v ON v.id = p.visit_id
			INNER JOIN mdl_mst_doctor md
				ON md.id = p.doctor_id
				AND md.staff_uuid IS NOT NULL
			WHERE p.institution_id = ?
			  AND p.deleted_at IS NULL

			UNION ALL

			SELECT mn.staff_uuid AS staff_uuid, p.visit_id AS visit_id
			FROM mdl_trx_visit_procedure p
			INNER JOIN period_visits v ON v.id = p.visit_id
			INNER JOIN mdl_mst_nurse mn
				ON mn.id = p.nurse_id
				AND mn.staff_uuid IS NOT NULL
			WHERE p.institution_id = ?
			  AND p.deleted_at IS NULL
			  AND p.nurse_id IS NOT NULL

			UNION ALL

			SELECT md.staff_uuid AS staff_uuid, d.visit_id AS visit_id
			FROM mdl_trx_diagnosis d
			INNER JOIN period_visits v ON v.id = d.visit_id
			INNER JOIN mdl_mst_doctor md
				ON md.id = d.doctor_id
				AND md.staff_uuid IS NOT NULL
			WHERE d.institution_id = ?
			  AND d.deleted_at IS NULL

			UNION ALL

			SELECT md.staff_uuid AS staff_uuid, a.visit_id AS visit_id
			FROM mdl_trx_anamnesa a
			INNER JOIN period_visits v ON v.id = a.visit_id
			INNER JOIN mdl_mst_doctor md
				ON md.id = a.doctor_id
				AND md.staff_uuid IS NOT NULL
			WHERE a.institution_id = ?
			  AND a.doctor_id IS NOT NULL

			UNION ALL

			SELECT mn.staff_uuid AS staff_uuid, a.visit_id AS visit_id
			FROM mdl_trx_anamnesa a
			INNER JOIN period_visits v ON v.id = a.visit_id
			INNER JOIN mdl_mst_nurse mn
				ON mn.id = a.nurse_id
				AND mn.staff_uuid IS NOT NULL
			WHERE a.institution_id = ?
			  AND a.nurse_id IS NOT NULL

			UNION ALL

			SELECT m.staff_id AS staff_uuid, m.visit_id AS visit_id
			FROM mdl_map_visit_contributor m
			INNER JOIN period_visits v ON v.id = m.visit_id
			WHERE m.institution_id = ?
			  AND m.delete_time IS NULL
		),
		per_staff AS (
			SELECT staff_uuid, COUNT(DISTINCT visit_id) AS visit_count
			FROM raw
			GROUP BY staff_uuid
		)
		SELECT
			s.uuid AS staff_id,
			s.name AS name,
			ps.visit_count AS visit_count,
			COALESCE(
				jsonb_agg(DISTINCT r.name) FILTER (WHERE r.name IS NOT NULL),
				'[]'::jsonb
			) AS roles
		FROM per_staff ps
		INNER JOIN mdl_mst_staff s
			ON s.uuid = ps.staff_uuid::text
			AND s.delete_time IS NULL
		LEFT JOIN mdl_map_role_staff mrs ON mrs.id_mst_staff = s.id
		LEFT JOIN mdl_mst_role r ON r.id = mrs.id_mst_role AND r.delete_time IS NULL
		GROUP BY s.uuid, s.name, ps.visit_count
		ORDER BY s.name ASC, s.uuid ASC
	`
)

var _ compensationrepo.ContributorDB = (*Conn)(nil)

// NewContributorDB returns a ContributorDB implementation bound to the xorm connection.
func NewContributorDB(db *xormlib.DBConnect) compensationrepo.ContributorDB {
	return &Conn{DB: db}
}

type detectedRow struct {
	SourceType  string         `xorm:"source_type"`
	VisitID     int64          `xorm:"visit_id"`
	StaffID     string         `xorm:"staff_id"`
	Name        string         `xorm:"name"`
	ClinicalID  int64          `xorm:"clinical_id"`
	ProcedureID sql.NullInt64  `xorm:"procedure_id"`
	DiagnosisID sql.NullInt64  `xorm:"diagnosis_id"`
	ProductID   sql.NullInt64  `xorm:"product_id"`
	Label       sql.NullString `xorm:"label"`
}

func (c *Conn) DetectForVisit(ctx context.Context, institutionID, visitID int64) ([]compensationrepo.DetectedAttribution, error) {
	out, err := c.DetectForVisits(ctx, institutionID, []int64{visitID})
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgDetectForVisit)
	}
	return out, nil
}

func (c *Conn) DetectForVisits(ctx context.Context, institutionID int64, visitIDs []int64) ([]compensationrepo.DetectedAttribution, error) {
	if len(visitIDs) == 0 {
		return []compensationrepo.DetectedAttribution{}, nil
	}

	visitIDsArg := pq.Array(visitIDs)
	type query struct {
		sql  string
		args []interface{}
	}
	queries := []query{
		{detectProcedureSQL, []interface{}{institutionID, visitIDsArg, institutionID, visitIDsArg}},
		{detectDiagnosisSQL, []interface{}{institutionID, visitIDsArg}},
		{detectAnamnesaSQL, []interface{}{institutionID, visitIDsArg, institutionID, visitIDsArg}},
		{detectMapSQL, []interface{}{institutionID, visitIDsArg}},
	}

	out := make([]compensationrepo.DetectedAttribution, 0)
	for _, q := range queries {
		rows, err := c.detectRows(ctx, q.sql, q.args...)
		if err != nil {
			return nil, errors.Wrap(err, wrapMsgDetectForVisits)
		}
		for _, row := range rows {
			out = append(out, row.toAttribution())
		}
	}
	return out, nil
}

type periodStaffDetectRow struct {
	StaffID    string `xorm:"staff_id"`
	Name       string `xorm:"name"`
	VisitCount int64  `xorm:"visit_count"`
	Roles      []byte `xorm:"roles"`
}

func (c *Conn) DetectStaffForPeriod(ctx context.Context, institutionID int64, periodStart, periodEndExclusive time.Time) ([]compensationrepo.PeriodStaffDetection, error) {
	var rows []periodStaffDetectRow
	err := c.DB.SlaveDB.Context(ctx).SQL(
		detectStaffForPeriodSQL,
		institutionID, periodStart, periodEndExclusive,
		institutionID,
		institutionID,
		institutionID,
		institutionID,
		institutionID,
		institutionID,
	).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgDetectStaffForPeriod)
	}

	out := make([]compensationrepo.PeriodStaffDetection, 0, len(rows))
	for _, row := range rows {
		roles := []string{}
		if len(row.Roles) > 0 {
			if err := json.Unmarshal(row.Roles, &roles); err != nil {
				return nil, errors.Wrap(err, wrapMsgDetectStaffForPeriod)
			}
			if roles == nil {
				roles = []string{}
			}
		}
		out = append(out, compensationrepo.PeriodStaffDetection{
			StaffID:    row.StaffID,
			Name:       row.Name,
			Roles:      roles,
			VisitCount: row.VisitCount,
		})
	}
	return out, nil
}

func (c *Conn) DetectForPeriodStaff(ctx context.Context, institutionID int64, staffID string, periodStart, periodEndExclusive time.Time) ([]compensationrepo.DetectedAttribution, error) {
	rows, err := c.detectRows(
		ctx,
		detectForPeriodStaffSQL,
		institutionID, periodStart, periodEndExclusive,
		institutionID, staffID,
		institutionID, staffID,
		institutionID, staffID,
		institutionID, staffID,
		institutionID, staffID,
		institutionID, staffID,
	)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgDetectForPeriodStaff)
	}
	out := make([]compensationrepo.DetectedAttribution, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toAttribution())
	}
	return out, nil
}

func (c *Conn) detectRows(ctx context.Context, sqlText string, args ...interface{}) ([]detectedRow, error) {
	var rows []detectedRow
	err := c.DB.SlaveDB.Context(ctx).SQL(sqlText, args...).Find(&rows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r detectedRow) toAttribution() compensationrepo.DetectedAttribution {
	attr := compensationrepo.DetectedAttribution{
		Type:          model.ContributionSourceType(r.SourceType),
		VisitID:       r.VisitID,
		StaffID:       r.StaffID,
		Name:          r.Name,
		ClinicalRowID: r.ClinicalID,
		ProductID:     r.ProductID,
		Label:         r.Label,
	}
	if r.ProcedureID.Valid {
		attr.ProcedureID = r.ProcedureID.Int64
	}
	if r.DiagnosisID.Valid {
		attr.DiagnosisID = r.DiagnosisID.Int64
	}
	return attr
}

func (c *Conn) UpsertManualContributor(ctx context.Context, row model.MapVisitContributor) error {
	execAffected := func(sqlText string, args ...interface{}) (int64, error) {
		res, err := c.writeSession(ctx).Exec(append([]interface{}{sqlText}, args...)...)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}

	const restoreSQL = `
		UPDATE mdl_map_visit_contributor t
		SET delete_time = NULL, added_by = ?
		FROM (
			SELECT id
			FROM mdl_map_visit_contributor
			WHERE visit_id = ? AND staff_id = ? AND delete_time IS NOT NULL
			ORDER BY id DESC
			LIMIT 1
		) d
		WHERE t.id = d.id
		  AND NOT EXISTS (
			SELECT 1
			FROM mdl_map_visit_contributor live
			WHERE live.visit_id = t.visit_id
			  AND live.staff_id = t.staff_id
			  AND live.delete_time IS NULL
		  )
	`
	restored, err := execAffected(restoreSQL, row.AddedBy, row.VisitID, row.StaffID)
	if err != nil {
		return errors.Wrap(err, wrapMsgUpsertManualContributor)
	}
	if restored > 0 {
		return nil
	}

	const insertSQL = `
		INSERT INTO mdl_map_visit_contributor (visit_id, staff_id, institution_id, added_by)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (visit_id, staff_id) WHERE delete_time IS NULL
		DO NOTHING
	`
	inserted, err := execAffected(insertSQL, row.VisitID, row.StaffID, row.InstitutionID, row.AddedBy)
	if err != nil {
		return errors.Wrap(err, wrapMsgUpsertManualContributor)
	}
	if inserted == 0 {
		return compensationrepo.ErrContributorAlreadyAdded
	}
	return nil
}

func (c *Conn) DeleteManualContributor(ctx context.Context, institutionID, visitID int64, staffID string) (bool, error) {
	const deleteSQL = `
		UPDATE mdl_map_visit_contributor
		SET delete_time = NOW()
		WHERE visit_id = ? AND staff_id = ? AND institution_id = ?
		  AND delete_time IS NULL
	`
	res, err := c.writeSession(ctx).Exec(deleteSQL, visitID, staffID, institutionID)
	if err != nil {
		return false, errors.Wrap(err, wrapMsgDeleteManualContributor)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, errors.Wrap(err, wrapMsgDeleteManualContributor)
	}
	return affected > 0, nil
}
