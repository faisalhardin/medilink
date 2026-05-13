package anamnesa

import (
	"context"
	"database/sql"
	"math"
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	anamnesarepo "github.com/faisalhardin/medilink/internal/entity/repo/anamnesa"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	practitionerrepo "github.com/faisalhardin/medilink/internal/entity/repo/practitioner"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/faisalhardin/medilink/internal/library/util/common"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

const (
	wrapMsgGetByVisit         = "AnamnesaUC.GetByVisitID"
	wrapMsgGetDetailedByVisit = "AnamnesaUC.GetDetailedByVisitID"
	wrapMsgUpsert             = "AnamnesaUC.Upsert"
)

type AnamnesaUC struct {
	AnamnesaDB     anamnesarepo.AnamnesaDB
	PatientDB      patientrepo.PatientDB
	PractitionerDB practitionerrepo.PractitionerDB
	Transaction    xormlib.DBTransactionInterface
}

func NewAnamnesaUC(u *AnamnesaUC) *AnamnesaUC {
	return u
}

func (u *AnamnesaUC) GetByVisitID(ctx context.Context, visitID int64) (*model.AnamnesaResponse, error) {
	userDetail, err := u.authorizeVisit(ctx, visitID)
	if err != nil {
		return nil, err
	}
	row, found, dbErr := u.AnamnesaDB.GetByVisitID(ctx, userDetail.InstitutionID, visitID)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, wrapMsgGetByVisit)
	}
	if !found {
		return nil, nil
	}
	resp := row.ToResponse()
	return &resp, nil
}

func (u *AnamnesaUC) GetDetailedByVisitID(ctx context.Context, visitID int64) (*model.AnamnesaDetailedResponse, error) {
	userDetail, err := u.authorizeVisit(ctx, visitID)
	if err != nil {
		return nil, err
	}
	row, found, dbErr := u.AnamnesaDB.GetDetailedByVisitID(ctx, userDetail.InstitutionID, visitID)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, wrapMsgGetDetailedByVisit)
	}
	if !found {
		return nil, nil
	}
	out := row.ToDetailedResponse()
	return out, nil
}

func (u *AnamnesaUC) Upsert(ctx context.Context, visitID int64, req model.UpsertAnamnesaRequest) (resp model.UpsertAnamnesaResponse, err error) {
	userDetail, authErr := u.authorizeVisit(ctx, visitID)
	if authErr != nil {
		return resp, authErr
	}

	if req.DoctorID != "" {
		missingDoctors, dbErr := u.PractitionerDB.MissingDoctorIDs(ctx, userDetail.InstitutionID, []string{req.DoctorID})
		if dbErr != nil {
			return resp, errors.Wrap(dbErr, wrapMsgUpsert)
		}
		if len(missingDoctors) > 0 {
			return resp, commonerr.SetNewUnprocessableEntityError("doctor_id", "doctor_id not found in institution")
		}
	}

	if req.NurseID != "" {
		missingNurse, dbErr := u.PractitionerDB.MissingNurseIDs(ctx, userDetail.InstitutionID, []string{req.NurseID})
		if dbErr != nil {
			return resp, errors.Wrap(dbErr, wrapMsgUpsert)
		}
		if len(missingNurse) > 0 {
			return resp, commonerr.SetNewUnprocessableEntityError("nurse_id", "nurse_id not found in institution")
		}
	}

	row := &model.TrxAnamnesa{
		VisitID:       visitID,
		InstitutionID: userDetail.InstitutionID,
		DoctorID:      common.NullableString(req.DoctorID),
		NurseID:       common.NullableString(req.NurseID),
		ChiefComplaint: null.String{
			String: req.ChiefComplaint,
			Valid:  req.ChiefComplaint != "",
		},
		SecondaryComplaint: req.SecondaryComplaint,
		HistoryOfIllness:   req.HistoryOfIllness,

		IllnessYears:  req.IllnessDuration.Years,
		IllnessMonths: req.IllnessDuration.Months,
		IllnessDays:   req.IllnessDuration.Days,

		VSSystolic:               common.NullInt16ToPointer(req.VitalSigns.Systolic),
		VSDiastolic:              common.NullInt16ToPointer(req.VitalSigns.Diastolic),
		VSPulse:                  common.NullInt16ToPointer(req.VitalSigns.Pulse),
		VSTemperature:            common.NullFloat32ToPointer(req.VitalSigns.Temperature),
		VSRespiratoryRate:        common.NullInt16ToPointer(req.VitalSigns.RespiratoryRate),
		VSOxygenSaturation:       common.NullInt16ToPointer(req.VitalSigns.OxygenSaturation),
		VSWeight:                 common.NullFloat32ToPointer(req.VitalSigns.Weight),
		VSHeight:                 common.NullFloat32ToPointer(req.VitalSigns.Height),
		VSHeightMeasurement:      req.VitalSigns.HeightMeasurement,
		VSAbdominalCircumference: common.NullFloat32ToPointer(req.VitalSigns.AbdominalCircumference),
		VSConsciousness:          req.VitalSigns.Consciousness,
		VSHeartRhythm:            req.VitalSigns.HeartRhythm,
		VSPregnancyStatus:        common.NullBoolToPointer(req.VitalSigns.PregnancyStatus),
		VSTriage:                 req.VitalSigns.Triage,

		GCSEye:    common.NullInt16ToPointer(req.GCS.Eye),
		GCSVerbal: common.NullInt16ToPointer(req.GCS.Verbal),
		GCSMotor:  common.NullInt16ToPointer(req.GCS.Motor),

		PainHasPain:  common.NullBoolToPointer(req.PainAssessment.HasPain),
		PainTrigger:  sql.NullString(req.PainAssessment.Trigger), //req.PainAssessment.Trigger,
		PainQuality:  sql.NullString(req.PainAssessment.Quality),
		PainLocation: sql.NullString(req.PainAssessment.Location),
		PainScale:    common.NullInt16ToPointer(req.PainAssessment.Scale),
		PainPattern:  sql.NullString(req.PainAssessment.Pattern),
		// FallRisk and Lifestyle are intentionally not mapped — not persisted.
	}

	row.VSMAP = computeMAP(common.NullInt16ToPointer(req.VitalSigns.Systolic), common.NullInt16ToPointer(req.VitalSigns.Diastolic))
	row.VSBMI = computeBMI(common.NullFloat32ToPointer(req.VitalSigns.Weight), common.NullFloat32ToPointer(req.VitalSigns.Height))
	row.VSBMIResult = computeBMIResult(row.VSBMI)

	session, beginErr := u.Transaction.Begin(ctx)
	if beginErr != nil {
		return resp, errors.Wrap(beginErr, wrapMsgUpsert)
	}
	defer u.Transaction.Finish(session, &err)
	txCtx := xormlib.SetDBSession(ctx, session)

	if dbErr := u.AnamnesaDB.Upsert(txCtx, row); dbErr != nil {
		err = errors.Wrap(dbErr, wrapMsgUpsert)
		return resp, err
	}

	resp.ID = row.ID
	return resp, nil
}

func (u *AnamnesaUC) authorizeVisit(ctx context.Context, visitID int64) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}

	visit, err := u.PatientDB.GetPatientVisitsByID(ctx, visitID)
	if err != nil {
		return model.UserJWTPayload{}, err
	}
	if visit.ID == 0 || visit.IDMstInstitution != userDetail.InstitutionID {
		return model.UserJWTPayload{}, commonerr.SetNewError(http.StatusNotFound, "visit_not_found", "visit was not found in this institution")
	}
	return userDetail, nil
}

func computeMAP(systolic, diastolic *int16) *int16 {
	if systolic == nil || diastolic == nil {
		return nil
	}
	v := int16(math.Round(float64(*diastolic) + float64(*systolic-*diastolic)/3.0))
	return &v
}

func computeBMI(weightKg, heightCm *float32) *float32 {
	if weightKg == nil || heightCm == nil || *heightCm <= 0 {
		return nil
	}
	heightM := float64(*heightCm) / 100.0
	bmi := float64(*weightKg) / (heightM * heightM)
	rounded := float32(math.Round(bmi*100) / 100)
	return &rounded
}

func computeBMIResult(bmi *float32) null.String {
	if bmi == nil {
		return null.String{}
	}
	v := float64(*bmi)
	label := "obesity"
	switch {
	case v < 18.5:
		label = "underweight"
	case v < 25:
		label = "normal"
	case v < 30:
		label = "overweight"
	}
	return null.String{String: label, Valid: true}
}
