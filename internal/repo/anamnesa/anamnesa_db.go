package anamnesa

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/model"
	anamnesarepo "github.com/faisalhardin/medilink/internal/entity/repo/anamnesa"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix          = "AnamnesaDB."
	WrapMsgGetByVisit         = WrapErrMsgPrefix + "GetByVisitID"
	WrapMsgGetDetailedByVisit = WrapErrMsgPrefix + "GetDetailedByVisitID"
	WrapMsgUpsert             = WrapErrMsgPrefix + "Upsert"
	WrapMsgGenerateUUID       = WrapErrMsgPrefix + "GenerateUUID"
)

type Conn struct {
	DB *xormlib.DBConnect
}

// NewAnamnesaDB returns an AnamnesaDB implementation bound to the xorm connection.
func NewAnamnesaDB(db *xormlib.DBConnect) anamnesarepo.AnamnesaDB {
	return &Conn{DB: db}
}

// writeSession returns the active TX session from ctx, falling back to the
// master engine. Mutations must use this so outbox enqueue stays atomic with
// the anamnesa write.
func (c *Conn) writeSession(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

// GetByVisitID loads the (at most one) anamnesa row for the visit. Uses
// xorm's struct mapping so the generated gcs_total column is populated via
// the "<-" read-only tag on the model.
func (c *Conn) GetByVisitID(ctx context.Context, institutionID, visitID int64) (*model.TrxAnamnesa, bool, error) {
	row := &model.TrxAnamnesa{}
	ok, err := c.DB.SlaveDB.Context(ctx).
		Table(model.TRX_ANAMNESA_TABLE).
		Where("institution_id = ?", institutionID).
		And("visit_id = ?", visitID).
		Get(row)
	if err != nil {
		return nil, false, errors.Wrap(err, WrapMsgGetByVisit)
	}
	if !ok {
		return nil, false, nil
	}
	return row, true, nil
}

// GetDetailedByVisitID loads the anamnesa row with doctor and nurse names via LEFT JOIN.
func (c *Conn) GetDetailedByVisitID(ctx context.Context, institutionID, visitID int64) (*model.TrxAnamnesaDetailRow, bool, error) {
	row := &model.TrxAnamnesaDetailRow{}
	const sql = `
		SELECT a.*,
			md.name AS doctor_name,
			mn.name AS nurse_name
		FROM mdl_trx_anamnesa a
		LEFT JOIN mdl_mst_doctor md
			ON md.id = a.doctor_id
			AND md.institution_id = a.institution_id
			AND md.active = TRUE
		LEFT JOIN mdl_mst_nurse mn
			ON mn.id = a.nurse_id
			AND mn.institution_id = a.institution_id
			AND mn.active = TRUE
		WHERE a.institution_id = ?
		  AND a.visit_id = ?
	`
	ok, err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, visitID).Get(row)
	if err != nil {
		return nil, false, errors.Wrap(err, WrapMsgGetDetailedByVisit)
	}
	if !ok {
		return nil, false, nil
	}

	fmt.Println("row", row)
	return row, true, nil
}

// Upsert inserts or overwrites the anamnesa row keyed by (institution_id, visit_id).
// The SQL is hand-written because xorm's Insert does not emit ON CONFLICT, and
// we rely on the unique index from the migration for correctness.
func (c *Conn) Upsert(ctx context.Context, row *model.TrxAnamnesa) error {
	if row == nil {
		return nil
	}
	if row.ID == "" {
		newID, err := uuid.NewV7()
		if err != nil {
			return errors.Wrap(err, WrapMsgGenerateUUID)
		}
		row.ID = newID.String()
	}

	const sql = `
		INSERT INTO mdl_trx_anamnesa (
			id, visit_id, institution_id, nurse_id, doctor_id,
			chief_complaint, secondary_complaint, history_of_illness,
			illness_years, illness_months, illness_days,
			vs_systolic, vs_diastolic, vs_pulse, vs_temperature,
			vs_respiratory_rate, vs_oxygen_saturation,
			vs_map, vs_weight, vs_height, vs_bmi, vs_bmi_result,
			vs_height_measurement, vs_abdominal_circumference, vs_consciousness, vs_heart_rhythm, vs_pregnancy_status, vs_triage,
			gcs_eye, gcs_verbal, gcs_motor,
			pain_has_pain, pain_trigger, pain_quality, pain_location, pain_scale, pain_pattern,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?, ?, ?,
			NOW(), NOW()
		)
		ON CONFLICT (institution_id, visit_id) DO UPDATE SET
			nurse_id             = EXCLUDED.nurse_id,
			doctor_id            = EXCLUDED.doctor_id,
			chief_complaint      = EXCLUDED.chief_complaint,
			secondary_complaint  = EXCLUDED.secondary_complaint,
			history_of_illness   = EXCLUDED.history_of_illness,
			illness_years        = EXCLUDED.illness_years,
			illness_months       = EXCLUDED.illness_months,
			illness_days         = EXCLUDED.illness_days,
			vs_systolic          = EXCLUDED.vs_systolic,
			vs_diastolic         = EXCLUDED.vs_diastolic,
			vs_pulse             = EXCLUDED.vs_pulse,
			vs_temperature       = EXCLUDED.vs_temperature,
			vs_respiratory_rate  = EXCLUDED.vs_respiratory_rate,
			vs_oxygen_saturation = EXCLUDED.vs_oxygen_saturation,
			vs_map               = EXCLUDED.vs_map,
			vs_weight            = EXCLUDED.vs_weight,
			vs_height            = EXCLUDED.vs_height,
			vs_bmi               = EXCLUDED.vs_bmi,
			vs_bmi_result        = EXCLUDED.vs_bmi_result,
			vs_height_measurement = EXCLUDED.vs_height_measurement,
			vs_abdominal_circumference = EXCLUDED.vs_abdominal_circumference,
			vs_consciousness     = EXCLUDED.vs_consciousness,
			vs_heart_rhythm      = EXCLUDED.vs_heart_rhythm,
			vs_pregnancy_status  = EXCLUDED.vs_pregnancy_status,
			vs_triage            = EXCLUDED.vs_triage,
			gcs_eye              = EXCLUDED.gcs_eye,
			gcs_verbal           = EXCLUDED.gcs_verbal,
			gcs_motor            = EXCLUDED.gcs_motor,
			pain_has_pain        = EXCLUDED.pain_has_pain,
			pain_trigger         = EXCLUDED.pain_trigger,
			pain_quality         = EXCLUDED.pain_quality,
			pain_location        = EXCLUDED.pain_location,
			pain_scale           = EXCLUDED.pain_scale,
			pain_pattern         = EXCLUDED.pain_pattern,
			updated_at           = NOW()
		RETURNING id, created_at, updated_at, gcs_total
	`

	res, err := c.writeSession(ctx).SQL(sql,
		row.ID,
		row.VisitID,
		row.InstitutionID,
		row.NurseID,
		row.DoctorID,
		row.ChiefComplaint,
		row.SecondaryComplaint,
		row.HistoryOfIllness,
		row.IllnessYears,
		row.IllnessMonths,
		row.IllnessDays,
		row.VSSystolic,
		row.VSDiastolic,
		row.VSPulse,
		row.VSTemperature,
		row.VSRespiratoryRate,
		row.VSOxygenSaturation,
		row.VSMAP,
		row.VSWeight,
		row.VSHeight,
		row.VSBMI,
		row.VSBMIResult,
		row.VSHeightMeasurement,
		row.VSAbdominalCircumference,
		row.VSConsciousness,
		row.VSHeartRhythm,
		row.VSPregnancyStatus,
		row.VSTriage,
		row.GCSEye,
		row.GCSVerbal,
		row.GCSMotor,
		row.PainHasPain,
		row.PainTrigger,
		row.PainQuality,
		row.PainLocation,
		row.PainScale,
		row.PainPattern,
	).QueryInterface()
	if err != nil {
		return errors.Wrap(err, WrapMsgUpsert)
	}

	if len(res) > 0 {
		r := res[0]
		if id, ok := r["id"].(string); ok {
			row.ID = id
		}
		row.CreatedAt = xormlib.ToTime(r["created_at"])
		row.UpdatedAt = xormlib.ToTime(r["updated_at"])
		if v := xormlib.ToInt16Ptr(r["gcs_total"]); v != nil {
			row.GCSTotal = v
		}
	}
	return nil
}
