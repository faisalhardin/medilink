package anamnesa

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	anamnesarepo "github.com/faisalhardin/medilink/internal/entity/repo/anamnesa"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix    = "AnamnesaDB."
	WrapMsgGetByVisit   = WrapErrMsgPrefix + "GetByVisitID"
	WrapMsgUpsert       = WrapErrMsgPrefix + "Upsert"
	WrapMsgGenerateUUID = WrapErrMsgPrefix + "GenerateUUID"
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
			id, visit_id, institution_id, nurse_id,
			chief_complaint, history_of_illness,
			vs_systolic, vs_diastolic, vs_pulse, vs_temperature,
			vs_respiratory_rate, vs_oxygen_saturation,
			vs_map, vs_weight, vs_height, vs_bmi, vs_bmi_result,
			gcs_eye, gcs_verbal, gcs_motor,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?, ?,
			?, ?, ?,
			NOW(), NOW()
		)
		ON CONFLICT (institution_id, visit_id) DO UPDATE SET
			nurse_id             = EXCLUDED.nurse_id,
			chief_complaint      = EXCLUDED.chief_complaint,
			history_of_illness   = EXCLUDED.history_of_illness,
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
			gcs_eye              = EXCLUDED.gcs_eye,
			gcs_verbal           = EXCLUDED.gcs_verbal,
			gcs_motor            = EXCLUDED.gcs_motor,
			updated_at           = NOW()
		RETURNING id, created_at, updated_at, gcs_total
	`

	res, err := c.writeSession(ctx).SQL(sql,
		row.ID,
		row.VisitID,
		row.InstitutionID,
		row.NurseID,
		row.ChiefComplaint,
		row.HistoryOfIllness,
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
		row.GCSEye,
		row.GCSVerbal,
		row.GCSMotor,
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
