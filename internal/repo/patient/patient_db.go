package patient

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix                        = "Conn."
	WrapMsgRegisterNewPatient               = WrapErrMsgPrefix + "RegisterNewPatient"
	WrapMsgRecordPatientVisit               = WrapErrMsgPrefix + "RecordPatientVisit"
	WrapMsgGetPatientVisitRecordByPatientID = WrapErrMsgPrefix + "GetPatientVisitRecordByPatientID"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func NewPatientDB(conn *Conn) *Conn {
	return conn
}

func (c *Conn) RegisterNewPatient(ctx context.Context, patient *model.MstPatientInstitution) (err error) {
	if patient.InstitutionID == 0 {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	session := c.DB.MasterDB

	_, err = session.Table(model.MST_PATIENT_INSTITUTION).InsertOne(patient)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}

func (c *Conn) GetPatients(ctx context.Context, params model.GetPatientParams) (patients model.GetPatientResponse, err error) {

	session := c.DB.MasterDB.Table(model.MST_PATIENT_INSTITUTION)

	if !params.DateOfBirth.IsZero() {
		session.Where("mmpi.date_of_birth = ?", params.DateOfBirth.Format(constant.DateFormatYYYYMMDDDashed))
	}
	if len(params.PatientUUIDs) > 0 {
		session.Where("mmpi.uuid = any(?)", params.PatientUUIDs)
	}
	if params.InstitutionID > 0 {
		session.Where("mmpi.institution_id = ?", params)
	}

	_, err = session.
		Alias("mmpi").
		FindAndCount(params)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRegisterNewPatient)
		return
	}

	return
}

func (c *Conn) RecordPatientVisit(ctx context.Context, request *model.MstPatientVisit) (err error) {
	session := c.DB.MasterDB

	_, err = session.
		Table(model.MST_PATIENT_VISIT).
		InsertOne(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgRecordPatientVisit)
		return
	}

	return
}

func (c *Conn) GetPatientVisitsRecordByPatientID(ctx context.Context, patientID int64) (mstPatientVisits []model.MstPatientVisit, err error) {
	session := c.DB.SlaveDB.Table(model.MST_PATIENT_VISIT).Alias("mmpv")

	err = session.Where("mmpv.patient_id = ?", patientID).
		Find(&mstPatientVisits)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetPatientVisitRecordByPatientID)
		return
	}

	return
}
