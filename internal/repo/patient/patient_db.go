package patient

import (
	"context"
	"fmt"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/constant"
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
	WrapMsgUpdatePatient                    = WrapErrMsgPrefix + "UpdatePatient"
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

func (c *Conn) GetPatients(ctx context.Context, params model.GetPatientParams) (patients []model.GetPatientResponse, err error) {

	if params.InstitutionID == 0 {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}
	session := c.DB.MasterDB.Table(model.MST_PATIENT_INSTITUTION)

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

func (c *Conn) UpdatePatient(ctx context.Context, request *model.UpdatePatientRequest) (err error) {

	session := c.DB.MasterDB.Table(model.MST_PATIENT_INSTITUTION)

	_, err = session.
		Where("uuid = ?", request.UUID).
		Update(request)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdatePatient)
		return
	}

	return
}
