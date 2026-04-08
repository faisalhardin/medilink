package recall

import (
	"context"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	wrapMsgInsert           = "RecallDB.Insert"
	wrapMsgUpdate           = "RecallDB.Update"
	wrapMsgGetByID          = "RecallDB.GetByID"
	wrapMsgGetNextByPatient = "RecallDB.GetNextByPatient"
	wrapMsgListUpcoming     = "RecallDB.ListUpcoming"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func NewRecallDB(conn *Conn) *Conn {
	return conn
}

func (c *Conn) Insert(ctx context.Context, r *model.TrxRecall) error {
	result, err := c.DB.MasterDB.SQL(`
		INSERT INTO mdl_trx_recall
		(id_mst_patient, id_mst_institution, scheduled_at, recall_type, notes, created_by_id_mst_staff, id_trx_patient_visit, create_time, update_time)
		VALUES (?, ?, ?::timestamptz, ?, ?, ?, ?, NOW(), NOW())
		RETURNING id, create_time, update_time
	`,
		r.IDMstPatient,
		r.IDMstInstitution,
		r.ScheduledAt.Time().UTC().Format(time.RFC3339),
		r.RecallType,
		r.Notes,
		r.CreatedByIDMstStaff,
		r.IDTrxPatientVisit,
	).QueryInterface()
	if err != nil {
		return errors.Wrap(err, wrapMsgInsert)
	}
	if len(result) > 0 {
		row := result[0]
		r.ID = row["id"].(int64)
		r.CreateTime = row["create_time"].(time.Time)
		r.UpdateTime = row["update_time"].(time.Time)
	}
	return nil
}

func (c *Conn) Update(ctx context.Context, id int64, institutionID int64, req model.UpdateRecallRequest) error {
	session := c.DB.MasterDB.Table(model.TrxRecallTableName).
		Where("id = ?", id).
		And("id_mst_institution = ?", institutionID)

	updates := make(map[string]interface{})
	if req.ScheduledAt != nil {
		updates["scheduled_at"] = req.ScheduledAt.Time().UTC().Format(time.RFC3339)
	}
	if req.RecallType != nil {
		updates["recall_type"] = *req.RecallType
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if len(updates) == 0 {
		return nil
	}

	_, err := session.Update(updates)
	if err != nil {
		return errors.Wrap(err, wrapMsgUpdate)
	}
	return nil
}

func (c *Conn) GetByID(ctx context.Context, id int64, institutionID int64) (model.TrxRecallJoinPatient, error) {
	var rec model.TrxRecallJoinPatient
	ok, err := c.DB.SlaveDB.Table(model.TrxRecallTableName).Alias("mtr").
		Join(database.SQLInner, "mdl_mst_patient_institution mmpi", "mmpi.id = mtr.id_mst_patient and mmpi.delete_time is null").
		Where("mtr.id = ?", id).
		And("mtr.id_mst_institution = ?", institutionID).
		Select("mtr.*, mmpi.uuid, mmpi.name").
		Get(&rec)
	if err != nil {
		return model.TrxRecallJoinPatient{}, errors.Wrap(err, wrapMsgGetByID)
	}
	if !ok {
		return model.TrxRecallJoinPatient{}, nil
	}
	return rec, nil
}

func (c *Conn) GetNextByPatient(ctx context.Context, patientUUID string, institutionID int64) (model.TrxRecallJoinPatient, bool, error) {
	var rec model.TrxRecallJoinPatient
	ok, err := c.DB.SlaveDB.Table(model.TrxRecallTableName).Alias("mtr").
		Join(database.SQLInner, "mdl_mst_patient_institution mmpi", "mmpi.id = mtr.id_mst_patient and mmpi.delete_time is null").
		Where("mmpi.uuid = ?", patientUUID).
		And("mtr.id_mst_institution = ?", institutionID).
		And("mtr.scheduled_at > ?", time.Now()).
		Asc("mtr.scheduled_at").
		Select("mtr.*, mmpi.uuid, mmpi.name").
		Limit(1).
		Get(&rec)
	if err != nil {
		return model.TrxRecallJoinPatient{}, false, errors.Wrap(err, wrapMsgGetNextByPatient)
	}
	return rec, ok, nil
}

func (c *Conn) ListUpcoming(ctx context.Context, params model.GetRecallParams, institutionID int64) ([]model.TrxRecallJoinPatient, error) {
	session := c.DB.SlaveDB.Table(model.TrxRecallTableName).Alias("mtr").
		Join(database.SQLInner, "mdl_mst_patient_institution mmpi", "mmpi.id = mtr.id_mst_patient and mmpi.delete_time is null").
		Where("mtr.id_mst_institution = ?", institutionID)

	if params.PatientUUID != "" {
		session = session.And("mmpi.uuid = ?", params.PatientUUID)
	}
	if params.IDMstPatient > 0 {
		session = session.And("mtr.id_mst_patient = ?", params.IDMstPatient)
	}
	if !params.FromTime.Time().IsZero() {
		session = session.And("mtr.scheduled_at >= ?", params.FromTime.Time())
	}
	if !params.ToTime.Time().IsZero() {
		session = session.And("mtr.scheduled_at <= ?", params.ToTime.Time())
	}
	if params.RecallType != "" {
		session = session.And("mtr.recall_type = ?", params.RecallType)
	}

	session = session.Asc("mtr.scheduled_at")

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}
	session = session.Limit(limit, params.Offset).
		Select("mtr.*, mmpi.uuid, mmpi.name")

	var list []model.TrxRecallJoinPatient
	err := session.Find(&list)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgListUpcoming)
	}
	return list, nil
}
