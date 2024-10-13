package patient

import (
	"context"
	"fmt"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix                        = "Conn."
	WrapMsgRegisterNewPatient               = WrapErrMsgPrefix + "RegisterNewPatient"
	WrapMsgRecordPatientVisit               = WrapErrMsgPrefix + "RecordPatientVisit"
	WrapMsgGetPatientVisitRecordByPatientID = WrapErrMsgPrefix + "GetPatientVisitRecordByPatientID"
	WrapMsgUpdatePatientVisit               = WrapErrMsgPrefix + "UpdatePatientVisit"
	WrapMsgUpdatePatient                    = WrapErrMsgPrefix + "UpdatePatient"
	WrapMsgGetPatientVisits                 = WrapErrMsgPrefix + "GetPatientVisits"
	WrapMsgInsertDtlPatientVisit            = WrapErrMsgPrefix + "InsertDtlPatientVisit"
	WrapMsgUpdateDtlPatientVisit            = WrapErrMsgPrefix + "UpdateDtlPatientVisit"
	WrapMsgGetDtlPatientVisit               = WrapErrMsgPrefix + "GetDtlPatientVisit"
	WrapMsgInsertTrxVisitProduct            = WrapErrMsgPrefix + "InsertTrxVisitProduct"
	WrapMsgUpdateTrxVisitProduct            = WrapErrMsgPrefix + "UpdateTrxVisitProduct"
	WrapMsgDeleteTrxVisitProduct            = WrapErrMsgPrefix + "DeleteTrxVisitProduct"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func NewPatientDB(conn *Conn) *Conn {
	return conn
}

func (c *Conn) RegisterNewPatient(ctx context.Context, patient *model.MstPatientInstitution) (err error) {
	if patient.InstitutionID == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}

	session := c.DB.MasterDB

	_, err = session.Table(model.MstPatientInstitutionTableName).InsertOne(patient)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}

func (c *Conn) GetPatients(ctx context.Context, params model.GetPatientParams) (patients []model.MstPatientInstitution, err error) {

	if params.InstitutionID == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}
	session := c.DB.MasterDB.Table(model.MstPatientInstitutionTableName)

	if !params.DateOfBirth.Time().IsZero() {
		session.Where("mmpi.date_of_birth::date = ?", params.DateOfBirth.Time().Format(constant.DateFormatYYYYMMDDDashed))
	}
	if len(params.PatientUUIDs) > 0 {
		session.Where("mmpi.uuid = any(?)", pq.Array(params.PatientUUIDs))
	}
	if len(params.Name) > 0 {
		splitNames := strings.Split(params.Name, " ")
		nameQuery := []string{}
		for _, name := range splitNames {
			nameQuery = append(nameQuery, fmt.Sprintf("%%%s%%", name))
		}
		session.Where("mmpi.name ILIKE ANY(?)", pq.Array(nameQuery))
	}

	_, err = session.
		Alias("mmpi").
		Where("mmpi.id_mst_institution = ?", params.InstitutionID).
		FindAndCount(&patients)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}

func (c *Conn) RecordPatientVisit(ctx context.Context, request *model.TrxPatientVisit) (err error) {
	session := c.DB.MasterDB

	_, err = session.
		Table(model.TrxPatientVisitTableName).
		InsertOne(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRecordPatientVisit)
		return
	}

	return
}

func (c *Conn) GetPatientVisitsRecordByPatientID(ctx context.Context, patientID int64) (mstPatientVisits []model.TrxPatientVisit, err error) {
	session := c.DB.SlaveDB.Table(model.TrxPatientVisitTableName).Alias("mtpv")

	err = session.Where("mtpv.patient_id = ?", patientID).
		Find(&mstPatientVisits)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisitRecordByPatientID)
		return
	}

	return
}

func (c *Conn) GetPatientVisits(ctx context.Context, params model.GetPatientVisitParams) (trxPatientVisit []model.TrxPatientVisit, err error) {
	if params.IDMstInstitution == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}

	session := c.DB.MasterDB.Table(model.TrxPatientVisitTableName)

	if len(params.PatientUUID) > 0 {
		session.Join(database.SQLInner, "mdl_mst_patient_institution mmpi", "mmpi.id = mtpv.id_mst_patient and mmpi.delete_time is null").
			Where("mmpi.uuid = ?", params.PatientUUID)
	}
	if params.PatientID > 0 {
		session.Where("mtpv.id_mst_patient = ?", params.PatientID)
	}

	if params.IDPatientVisit > 0 {
		session.Where("mtpv.id = ?", params.IDPatientVisit)
	}

	err = session.Alias("mtpv").
		Where("mtpv.id_mst_institution = ?", params.IDMstInstitution).
		Find(&trxPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	return
}

func (c *Conn) UpdatePatientVisit(ctx context.Context, trxVisit model.TrxPatientVisit) (err error) {

	session := c.DB.MasterDB.Table(model.TrxPatientVisitTableName)

	_, err = session.
		ID(trxVisit.ID).
		Where("id_mst_institution = ?", trxVisit.IDMstInstitution).
		Cols("action", "status", "notes").
		Update(trxVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdatePatientVisit)
		return
	}

	return
}

func (c *Conn) UpdatePatient(ctx context.Context, request *model.UpdatePatientRequest) (err error) {

	session := c.DB.MasterDB.Table(model.MstPatientInstitutionTableName)

	_, err = session.
		Where("uuid = ?", request.UUID).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdatePatient)
		return
	}

	return
}

func (c *Conn) InsertDtlPatientVisit(ctx context.Context, request *model.DtlPatientVisit) (err error) {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		Table(model.DtlPatientVisitTableName).
		InsertOne(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRecordPatientVisit)
		return
	}

	return
}

func (c *Conn) UpdateDtlPatientVisit(ctx context.Context, request *model.DtlPatientVisit) (err error) {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		ID(request.ID).
		Table(model.DtlPatientVisitTableName).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateDtlPatientVisit)
		return
	}

	return
}

func (c *Conn) GetDtlPatientVisit(ctx context.Context, params model.DtlPatientVisit) (dtlPatientVisit []model.DtlPatientVisit, err error) {
	session := c.DB.SlaveDB.Table(model.DtlPatientVisitTableName)

	if params.IDTrxPatientVisit > 0 {
		session.Where("mdpv.id_trx_patient_visit = ?", params.IDTrxPatientVisit)
	}

	if params.ID > 0 {
		session.Where("mdpv.id = ?", params.ID)
	}

	err = session.Alias("mdpv").
		Find(&dtlPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetDtlPatientVisit)
		return
	}

	return
}

func (c *Conn) InsertTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error) {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		Table(model.TrxVisitProductTableName).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertTrxVisitProduct)
		return
	}

	return nil
}

func (c *Conn) UpdateTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error) {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		Table(model.TrxVisitProductTableName).
		ID(request.ID).
		Where("id_mst_institution = ?", request.IDMstInstitution).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateTrxVisitProduct)
		return
	}

	return
}

func (c *Conn) DeleteTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error) {
	session := xormlib.GetDBSession(ctx)
	if session == nil {
		session = c.DB.MasterDB.Context(ctx)
	}

	_, err = session.
		Table(model.TrxVisitProductTableName).
		ID(request.ID).
		Where("id_mst_institution = ?", request.IDMstInstitution).
		Delete(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgDeleteTrxVisitProduct)
		return
	}

	return
}

func (c *Conn) GetTrxVisitProduct(ctx context.Context, params model.TrxVisitProduct) (trxVisitProduct []model.TrxVisitProduct, err error) {
	session := c.DB.SlaveDB.Table(model.TrxVisitProductTableName)

	if params.IDTrxPatientVisit > 0 {
		session.Where("mtvp.id_trx_patient_visit = ?", params.IDTrxPatientVisit)
	}

	if params.ID > 0 {
		session.Where("mtvp.id = ?", params.ID)
	}

	err = session.Alias("mtvp").
		Where("id_mst_institution = ?", params.IDMstInstitution).
		Find(&trxVisitProduct)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetDtlPatientVisit)
		return
	}

	return
}
