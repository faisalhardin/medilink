package patient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	WrapMsgGetPatientByID                   = WrapErrMsgPrefix + "GetPatientByID"
	WrapMsgGetPatients                      = WrapErrMsgPrefix + "GetPatients"
	WrapMsgRecordPatientVisit               = WrapErrMsgPrefix + "RecordPatientVisit"
	WrapMsgGetPatientVisitRecordByPatientID = WrapErrMsgPrefix + "GetPatientVisitRecordByPatientID"
	WrapMsgGetPatientVisitsByID             = WrapErrMsgPrefix + "GetPatientVisitsByID"
	WrapMsgUpdatePatientVisit               = WrapErrMsgPrefix + "UpdatePatientVisit"
	WrapMsgUpdatePatient                    = WrapErrMsgPrefix + "UpdatePatient"
	WrapMsgGetPatientVisits                 = WrapErrMsgPrefix + "GetPatientVisits"
	WrapMsgInsertDtlPatientVisit            = WrapErrMsgPrefix + "InsertDtlPatientVisit"
	WrapMsgUpdateDtlPatientVisit            = WrapErrMsgPrefix + "UpdateDtlPatientVisit"
	WrapMsgGetDtlPatientVisit               = WrapErrMsgPrefix + "GetDtlPatientVisit"
	WrapMsgGetDtlPatientVisitByID           = WrapErrMsgPrefix + "GetDtlPatientVisitByID"
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

func (c *Conn) GetPatientByID(ctx context.Context, patientID int64) (patient model.MstPatientInstitution, err error) {
	session := c.DB.SlaveDB.Table(model.MstPatientInstitutionTableName)
	_, err = session.Where("id = ?", patientID).Get(&patient)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientByID)
		return
	}

	return
}

func (c *Conn) GetPatients(ctx context.Context, params model.GetPatientParams) (patients []model.MstPatientInstitution, err error) {
	patients = []model.MstPatientInstitution{}

	if params.InstitutionID == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}
	session := c.DB.SlaveDB.Table(model.MstPatientInstitutionTableName)

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

	if len(params.NIK) > 0 {
		session.Where("mmpi.nik = ?", params.NIK)
	}

	if len(params.PhoneNumber) > 0 {
		session.Where("mmpi.phone_number ILIKE ?", params.PhoneNumber)
	}

	_, err = session.
		Alias("mmpi").
		Where("mmpi.id_mst_institution = ?", params.InstitutionID).
		FindAndCount(&patients)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatients)
		return
	}

	return
}

func (c *Conn) RecordPatientVisit(ctx context.Context, request *model.TrxPatientVisit) (err error) {
	session := c.DB.MasterDB

	if request.ProductCart == nil {
		emptyArray := []string{}
		request.ProductCart, _ = json.Marshal(emptyArray)
	}

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

func (c *Conn) GetPatientVisitsByID(ctx context.Context, visitID int64) (mstPatientVisits model.TrxPatientVisit, err error) {
	session := c.DB.SlaveDB.Table(model.TrxPatientVisitTableName).Alias("mtpv")

	err = session.Where("mtpv.id = ?", visitID).
		Find(&mstPatientVisits)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisitsByID)
		return
	}

	return
}

func (c *Conn) GetPatientVisits(ctx context.Context, params model.GetPatientVisitParams) (trxPatientVisit []model.GetPatientVisitResponse, err error) {
	if params.IDMstInstitution == 0 {
		err = commonerr.SetNewNoInstitutionError()
		return
	}

	session := c.DB.MasterDB.Table(model.TrxPatientVisitTableName)

	if len(params.PatientUUID) > 0 {
		session.Where("mmpi.uuid = ?", params.PatientUUID)
	}
	if params.PatientID > 0 {
		session.Where("mtpv.id_mst_patient = ?", params.PatientID)
	}

	if params.IDPatientVisit > 0 {
		session.Where("mtpv.id = ?", params.IDPatientVisit)
	}

	if params.IDMstJourneyBoard > 0 {
		session.Where("mtpv.id_mst_journey_board = ?", params.IDMstJourneyBoard)
	}
	if !params.FromTime.IsZero() && !params.ToTime.IsZero() {
		session.Where("mtpv.create_time between ? and ?", params.FromTime.Format(time.RFC3339), params.ToTime.Format(time.RFC3339))
	}

	session.
		Join(database.SQLInner, "mdl_mst_patient_institution mmpi", "mtpv.id_mst_patient = mmpi.id and mmpi.delete_time is null").
		Join(database.SQLLeft, "mdl_mst_service_point mmsp", "mmsp.id = mtpv.id_mst_service_point").
		Join(database.SQLLeft, "mdl_mst_journey_point mmjp", "mmjp.id = mtpv.id_mst_journey_point")

	err = session.Alias("mtpv").
		Where("mtpv.id_mst_institution = ?", params.IDMstInstitution).
		Select("mtpv.id, mtpv.action, mtpv.create_time, mtpv.update_time, mtpv.id_mst_journey_point, mtpv.id_mst_service_point, mtpv.mst_journey_point_id_update_unix_time, mtpv.product_cart, mmpi.id, mmpi.name, mmsp.id, mmsp.name, mmpi.sex, mmpi.uuid,  mmjp.id, mmjp.name").
		Find(&trxPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisits)
		return
	}

	return
}

func (c *Conn) UpdatePatientVisit(ctx context.Context, updateRequest model.UpdatePatientVisitRequest) (trxVisit model.TrxPatientVisit, err error) {
	session := c.DB.MasterDB.Table(model.TrxPatientVisitTableName)

	trxVisit = model.TrxPatientVisit{
		ID:               updateRequest.ID,
		IDMstInstitution: updateRequest.IDMstInstitution,
	}

	if updateRequest.IDMstJourneyBoard.Valid {
		trxVisit.IDMstJourneyBoard = updateRequest.IDMstJourneyBoard.Int64
		session.Cols("id_mst_journey_board")
	}

	if updateRequest.IDMstJourneyPoint.Valid {
		trxVisit.IDMstJourneyPoint = updateRequest.IDMstJourneyPoint.Int64
		session.Cols("id_mst_journey_point")
	}

	if updateRequest.IDMstServicePoint.Valid {
		trxVisit.IDMstServicePoint = updateRequest.IDMstServicePoint.Int64
		session.Cols("id_mst_service_point")
	}

	if updateRequest.UpdateTimeMstJourneyPointID.Valid {
		trxVisit.UpdateTimeMstJourneyPointID = updateRequest.UpdateTimeMstJourneyPointID.Int64
		session.Cols("mst_journey_point_id_update_unix_time")
	}

	if updateRequest.ProductCart != nil {
		b, _ := json.Marshal(updateRequest.ProductCart)
		trxVisit.ProductCart = b
		session.Cols("product_cart")
	}

	_, err = session.
		ID(trxVisit.ID).
		Where("id_mst_institution = ?", trxVisit.IDMstInstitution).
		Update(&trxVisit)
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
		err = errors.Wrap(err, WrapMsgInsertDtlPatientVisit)
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
		Cols("notes", "contributors").
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateDtlPatientVisit)
		return
	}

	return
}

func (c *Conn) GetDtlPatientVisit(ctx context.Context, params model.GetDtlPatientVisitParams) (dtlPatientVisit []model.DtlPatientVisit, err error) {
	dtlPatientVisit = []model.DtlPatientVisit{}

	session := c.DB.SlaveDB.Table(model.DtlPatientVisitTableName)
	if len(params.IDsTrxPatientVisit) > 0 {
		session.Where("mdpv.id_trx_patient_visit = any(?)", pq.Array(params.IDsTrxPatientVisit))
	}
	if len(params.IDs) > 0 {
		session.Where("mdpv.id = ANY(?)", pq.Array(params.IDs))
	}
	if params.Limit > 0 {
		session.Limit(params.Limit, params.Offset)
	}

	err = session.Alias("mdpv").
		Find(&dtlPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetDtlPatientVisit)
		return
	}

	return
}

func (c *Conn) GetDtlPatientVisitByID(ctx context.Context, id int64) (dtlPatientVisit model.DtlPatientVisit, err error) {
	session := c.DB.SlaveDB.Table(model.DtlPatientVisitTableName)

	_, err = session.Alias("mdpv").
		Where("id = ?", id).
		Get(&dtlPatientVisit)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetDtlPatientVisitByID)
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
		InsertOne(request)
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

func (c *Conn) UpsertTrxVisitProduct(ctx context.Context, request *model.TrxVisitProduct) (err error) {
	if request.ID > 0 && request.Quantity > 0 {
		return c.UpdateTrxVisitProduct(ctx, request)
	}
	if request.ID > 0 && request.Quantity == 0 {
		return c.DeleteTrxVisitProduct(ctx, request)
	}
	return c.InsertTrxVisitProduct(ctx, request)
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

func (c *Conn) GetTrxVisitProduct(ctx context.Context, params model.GetVisitProductRequest) (trxVisitProduct []model.TrxVisitProduct, err error) {
	session := c.DB.SlaveDB.Table(model.TrxVisitProductTableName)

	if params.VisitID > 0 {
		session.Where("mtvp.id_trx_patient_visit = ?", params.VisitID)
	}

	if params.VisitProductID > 0 {
		session.Where("mtvp.id = ?", params.VisitProductID)
	}

	err = session.Alias("mtvp").
		Where("id_mst_institution = ?", params.InstitutionID).
		Find(&trxVisitProduct)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetDtlPatientVisit)
		return
	}

	return
}
