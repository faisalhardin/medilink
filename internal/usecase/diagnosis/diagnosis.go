package diagnosis

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	diagnosisrepo "github.com/faisalhardin/medilink/internal/entity/repo/diagnosis"
	icd10repo "github.com/faisalhardin/medilink/internal/entity/repo/icd10"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	practitionerrepo "github.com/faisalhardin/medilink/internal/entity/repo/practitioner"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

const (
	wrapMsgGetByVisitID = "DiagnosisUC.GetByVisitID"
	wrapMsgSave         = "DiagnosisUC.Save"
	wrapMsgDelete       = "DiagnosisUC.Delete"
)

type DiagnosisUC struct {
	DiagnosisDB    diagnosisrepo.DiagnosisDB
	PatientDB      patientrepo.PatientDB
	ICD10DB        icd10repo.ICD10DB
	PractitionerDB practitionerrepo.PractitionerDB
	Transaction    xormlib.DBTransactionInterface
}

func NewDiagnosisUC(u *DiagnosisUC) *DiagnosisUC {
	return u
}

func (u *DiagnosisUC) GetByVisitID(ctx context.Context, visitID int64) ([]model.DiagnosisResponse, error) {
	userDetail, err := u.authorizeVisit(ctx, visitID)
	if err != nil {
		return nil, err
	}

	rows, dbErr := u.DiagnosisDB.GetActiveByVisitID(ctx, userDetail.InstitutionID, visitID)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, wrapMsgGetByVisitID)
	}

	resp := make([]model.DiagnosisResponse, len(rows))
	for i, row := range rows {
		resp[i] = row.ToResponse()
	}
	return resp, nil
}

func (u *DiagnosisUC) Save(ctx context.Context, visitID int64, req model.SaveDiagnosesRequest) (resp model.SaveDiagnosesResponse, err error) {
	userDetail, authErr := u.authorizeVisit(ctx, visitID)
	if authErr != nil {
		return resp, authErr
	}

	prognosis := req.Prognosis
	if prognosis == "" {
		prognosis = model.PrognosisMalam
	}

	now := time.Now().UTC()
	errMsg := commonerr.NewErrorMessage()

	codeSet := make(map[string]struct{})
	doctorSet := make(map[string]struct{})
	seenUpdateIDs := make(map[int64]struct{})

	for i, item := range req.Diagnoses {
		if item.Type == model.DiagnosisTypePrimary && item.Rank != 1 {
			errMsg.Append(fmt.Sprintf("diagnoses on row %d", i+1), "primary diagnosis must have rank 1")
		}
		if item.OnsetDate != nil && !item.OnsetDate.Time.IsZero() && item.OnsetDate.Time.After(now) {
			errMsg.Append(fmt.Sprintf("diagnoses on row %d", i+1), "onset_date cannot be in the future")
		}
		codeSet[item.ICD10Code] = struct{}{}
		doctorSet[item.DoctorID] = struct{}{}
		if item.ID != nil {
			if _, exists := seenUpdateIDs[*item.ID]; exists {
				errMsg.Append(fmt.Sprintf("diagnoses on row %d", i+1), "duplicate diagnosis id in payload")
			}
			seenUpdateIDs[*item.ID] = struct{}{}
		}
	}

	primaryCount := 0
	for _, item := range req.Diagnoses {
		if item.Type == model.DiagnosisTypePrimary {
			primaryCount++
		}
	}
	if primaryCount != 1 {
		errMsg.Append("diagnoses", "exactly one diagnosis with type=primary is required")
	}

	codes := keys(codeSet)
	icd10Rows, dbErr := u.ICD10DB.GetByCodes(ctx, codes)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}
	displayByCode := make(map[string]string, len(icd10Rows))
	for _, r := range icd10Rows {
		displayByCode[r.Code] = r.Display
	}
	for i, item := range req.Diagnoses {
		if _, ok := displayByCode[item.ICD10Code]; !ok {
			errMsg.Append(fmt.Sprintf("diagnoses[%d].icd10_code", i), "icd10_code not found in reference table")
		}
	}

	doctorIDs := keys(doctorSet)
	missingDoctors, dbErr := u.PractitionerDB.MissingDoctorIDs(ctx, userDetail.InstitutionID, doctorIDs)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}
	if len(missingDoctors) > 0 {
		missingSet := make(map[string]struct{}, len(missingDoctors))
		for _, id := range missingDoctors {
			missingSet[id] = struct{}{}
		}
		for i, item := range req.Diagnoses {
			if _, miss := missingSet[item.DoctorID]; miss {
				errMsg.Append(fmt.Sprintf("diagnoses[%d].doctor_id", i), "doctor_id not found in institution")
			}
		}
	}

	existingRows, dbErr := u.DiagnosisDB.GetActiveByVisitID(ctx, userDetail.InstitutionID, visitID)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}

	existingMap := make(map[int64]model.TrxDiagnosis, len(existingRows))
	for _, row := range existingRows {
		existingMap[row.ID] = row.AsTrxDiagnosis()
	}

	toInsert := make([]model.TrxDiagnosis, 0)
	toUpdate := make([]model.TrxDiagnosis, 0)
	requestedIDs := make(map[int64]struct{})

	for i, item := range req.Diagnoses {
		row := model.TrxDiagnosis{
			VisitID:            visitID,
			InstitutionID:      userDetail.InstitutionID,
			DoctorID:           item.DoctorID,
			ICD10Code:          item.ICD10Code,
			ICD10Display:       displayByCode[item.ICD10Code],
			Rank:               item.Rank,
			Type:               item.Type,
			Case:               item.Case,
			ClinicalStatus:     item.ClinicalStatus,
			VerificationStatus: item.VerificationStatus,
			Prognosis:          prognosis,
		}
		if item.Note != nil {
			row.Note = null.String{String: strings.TrimSpace(*item.Note), Valid: true}
		}
		if item.OnsetDate != nil && !item.OnsetDate.Time.IsZero() {
			t := item.OnsetDate.Time
			row.OnsetDate = &t
		}

		if item.ID == nil {
			toInsert = append(toInsert, row)
			continue
		}
		if _, ok := existingMap[*item.ID]; !ok {
			errMsg.Append(fmt.Sprintf("diagnoses[%d].id", i), "diagnosis id not found for this visit")
			continue
		}
		row.ID = *item.ID
		requestedIDs[row.ID] = struct{}{}
		toUpdate = append(toUpdate, row)
	}

	if len(errMsg.ErrorList) > 0 {
		errMsg.SetUnprocessableEntity()
		return resp, errMsg
	}

	toSoftDelete := make([]int64, 0)
	for id := range existingMap {
		if _, keep := requestedIDs[id]; !keep {
			toSoftDelete = append(toSoftDelete, id)
		}
	}

	session, beginErr := u.Transaction.Begin(ctx)
	if beginErr != nil {
		return resp, errors.Wrap(beginErr, wrapMsgSave)
	}
	defer u.Transaction.Finish(session, &err)
	txCtx := xormlib.SetDBSession(ctx, session)

	if dbErr = u.DiagnosisDB.SoftDeleteByIDs(txCtx, userDetail.InstitutionID, visitID, toSoftDelete); dbErr != nil {
		err = errors.Wrap(dbErr, wrapMsgSave)
		return resp, err
	}
	if dbErr = u.DiagnosisDB.BulkInsert(txCtx, toInsert); dbErr != nil {
		err = errors.Wrap(dbErr, wrapMsgSave)
		return resp, err
	}
	if dbErr = u.DiagnosisDB.BulkUpdate(txCtx, toUpdate); dbErr != nil {
		err = errors.Wrap(dbErr, wrapMsgSave)
		return resp, err
	}

	resp = model.SaveDiagnosesResponse{
		Saved:   len(toInsert) + len(toUpdate),
		Deleted: len(toSoftDelete),
	}
	return resp, nil
}

func (u *DiagnosisUC) Delete(ctx context.Context, visitID, diagnosisID int64) error {
	userDetail, err := u.authorizeVisit(ctx, visitID)
	if err != nil {
		return err
	}
	found, dbErr := u.DiagnosisDB.SoftDeleteByID(ctx, userDetail.InstitutionID, visitID, diagnosisID)
	if dbErr != nil {
		return errors.Wrap(dbErr, wrapMsgDelete)
	}
	if !found {
		return commonerr.SetNewError(http.StatusNotFound, "diagnosis_not_found", "diagnosis row was not found for this visit")
	}
	return nil
}

func (u *DiagnosisUC) authorizeVisit(ctx context.Context, visitID int64) (model.UserJWTPayload, error) {
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

func keys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
