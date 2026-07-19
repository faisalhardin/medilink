package procedure

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	institutionrepo "github.com/faisalhardin/medilink/internal/entity/repo/institution"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	practitionerrepo "github.com/faisalhardin/medilink/internal/entity/repo/practitioner"
	procedurerepo "github.com/faisalhardin/medilink/internal/entity/repo/procedure"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	wrapMsgSearchICD9CM = "ProcedureUC.SearchICD9CM"
	wrapMsgGetByVisitID = "ProcedureUC.GetByVisitID"
	wrapMsgSave         = "ProcedureUC.Save"
	wrapMsgDelete       = "ProcedureUC.Delete"
	wrapMsgGetHistory   = "ProcedureUC.GetPatientHistory"

	defaultHistoryLimit = 20
	maxHistoryLimit     = 100
)

type ProcedureUC struct {
	ProcedureDB     procedurerepo.ProcedureDB
	ICD9CMDB        procedurerepo.ICD9CMDB
	InstitutionRepo institutionrepo.InstitutionDB
	PatientDB       patientrepo.PatientDB
	PractitionerDB  practitionerrepo.PractitionerDB
	Transaction     xormlib.DBTransactionInterface
}

func NewProcedureUC(u *ProcedureUC) *ProcedureUC {
	return u
}

func (u *ProcedureUC) SearchICD9CM(ctx context.Context, q string, limit int) ([]model.ICD9CMOption, error) {
	rows, err := u.ICD9CMDB.Search(ctx, q, limit)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgSearchICD9CM)
	}
	return rows, nil
}

func (u *ProcedureUC) GetByVisitID(ctx context.Context, visitID int64) ([]model.ProcedureEntry, error) {
	userDetail, err := u.authorizeVisit(ctx, visitID)
	if err != nil {
		return nil, err
	}

	rows, dbErr := u.ProcedureDB.GetActiveByVisitID(ctx, userDetail.InstitutionID, visitID)
	if dbErr != nil {
		return nil, errors.Wrap(dbErr, wrapMsgGetByVisitID)
	}

	resp := make([]model.ProcedureEntry, len(rows))
	for i, row := range rows {
		resp[i] = row.ToResponse()
	}
	return resp, nil
}

// Save atomically replaces the full procedure set for a visit.
func (u *ProcedureUC) Save(ctx context.Context, visitID int64, req model.SaveProceduresRequest) (resp model.SaveProceduresSummary, err error) {
	userDetail, authErr := u.authorizeVisit(ctx, visitID)
	if authErr != nil {
		return resp, authErr
	}

	errMsg := commonerr.NewErrorMessage()

	sets := collectInputSets(req.Procedures, errMsg)
	if len(errMsg.ErrorList) > 0 {
		errMsg.SetUnprocessableEntity()
		return resp, errMsg
	}

	productNames, dbErr := u.validateProducts(ctx, userDetail.InstitutionID, sets.productIDs, req.Procedures, errMsg)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}

	if dbErr = u.validateDoctors(ctx, userDetail.InstitutionID, sets.doctorIDs, req.Procedures, errMsg); dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}

	if dbErr = u.validateNurses(ctx, userDetail.InstitutionID, sets.nurseIDs, req.Procedures, errMsg); dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}

	icd9cmDisplays, dbErr := u.validateAndResolveICD9CM(ctx, sets.icd9cmCodes, req.Procedures, errMsg)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}

	if len(errMsg.ErrorList) > 0 {
		errMsg.SetUnprocessableEntity()
		return resp, errMsg
	}

	snapshots, dbErr := u.resolveSnapshotNames(ctx, userDetail.InstitutionID, sets.doctorIDs, sets.nurseIDs)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}
	snapshots.productNames = productNames
	snapshots.icd9cmDisplays = icd9cmDisplays

	existingRows, dbErr := u.ProcedureDB.GetActiveByVisitID(ctx, userDetail.InstitutionID, visitID)
	if dbErr != nil {
		return resp, errors.Wrap(dbErr, wrapMsgSave)
	}

	toInsert, toUpdate, toSoftDelete, diffErr := diffProcedureRows(req.Procedures, existingRows, visitID, userDetail.InstitutionID, snapshots, errMsg)
	if diffErr != nil {
		return resp, diffErr
	}

	if len(errMsg.ErrorList) > 0 {
		errMsg.SetUnprocessableEntity()
		return resp, errMsg
	}

	if dbErr = u.persistAtomically(ctx, userDetail.InstitutionID, visitID, toInsert, toUpdate, toSoftDelete, &err); dbErr != nil {
		return resp, dbErr
	}

	return model.SaveProceduresSummary{
		Saved:   len(toInsert) + len(toUpdate),
		Deleted: len(toSoftDelete),
	}, nil
}

func (u *ProcedureUC) Delete(ctx context.Context, visitID, procedureID int64) error {
	userDetail, err := u.authorizeVisit(ctx, visitID)
	if err != nil {
		return err
	}
	found, dbErr := u.ProcedureDB.SoftDeleteByID(ctx, userDetail.InstitutionID, visitID, procedureID)
	if dbErr != nil {
		return errors.Wrap(dbErr, wrapMsgDelete)
	}
	if !found {
		return commonerr.SetNewError(http.StatusNotFound, "procedure_not_found", "procedure row was not found for this visit")
	}
	return nil
}

func (u *ProcedureUC) GetPatientHistory(ctx context.Context, patientUUID string, limit, offset int) (model.ProcedureHistoryResponse, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.ProcedureHistoryResponse{}, commonerr.SetNewUnauthorizedAPICall()
	}

	if limit <= 0 {
		limit = defaultHistoryLimit
	}
	if limit > maxHistoryLimit {
		limit = maxHistoryLimit
	}
	if offset < 0 {
		offset = 0
	}

	rows, total, dbErr := u.ProcedureDB.GetPatientProcedureHistory(ctx, userDetail.InstitutionID, patientUUID, limit, offset)
	if dbErr != nil {
		return model.ProcedureHistoryResponse{}, errors.Wrap(dbErr, wrapMsgGetHistory)
	}

	if rows == nil {
		rows = []model.ProcedureHistoryEntry{}
	}

	return model.ProcedureHistoryResponse{
		Data:   rows,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// ─── Save helpers ─────────────────────────────────────────────────────────────

// inputSets holds unique IDs extracted from the payload for batch validation.
type inputSets struct {
	productIDs    []int64
	doctorIDs     []string
	nurseIDs      []string
	icd9cmCodes   []string
	seenUpdateIDs map[int64]struct{}
}

// snapshotMaps holds the resolved display names written into each DB row at save time.
type snapshotMaps struct {
	productNames   map[int64]string
	doctorNames    map[string]string
	nurseNames     map[string]string
	icd9cmDisplays map[string]string
}

// collectInputSets scans each payload row, collecting unique IDs for batch lookups
// and appending basic field-level errors (required, duplicate).
func collectInputSets(rows []model.SaveProcedureRow, errMsg *commonerr.ErrorMessage) inputSets {
	productIDSet := make(map[int64]struct{})
	doctorIDSet := make(map[string]struct{})
	nurseIDSet := make(map[string]struct{})
	icd9cmCodeSet := make(map[string]struct{})
	seenUpdateIDs := make(map[int64]struct{})

	for i, row := range rows {
		if row.ProductID.Valid && row.ProductID.Int64 > 0 {
			productIDSet[row.ProductID.Int64] = struct{}{}
		}
		if row.DoctorID == "" {
			errMsg.Append(fmt.Sprintf("procedures[%d].doctor_id", i), "doctor_id is required")
		} else {
			doctorIDSet[row.DoctorID] = struct{}{}
		}
		if row.NurseID.Valid && row.NurseID.String != "" {
			nurseIDSet[row.NurseID.String] = struct{}{}
		}
		if !row.ICD9CMCode.Valid || row.ICD9CMCode.String == "" {
			errMsg.Append(fmt.Sprintf("procedures[%d].icd9cm_code", i), "icd9cm_code is required")
		} else {
			icd9cmCodeSet[row.ICD9CMCode.String] = struct{}{}
		}
		if row.ID.Valid {
			if _, exists := seenUpdateIDs[row.ID.Int64]; exists {
				errMsg.Append(fmt.Sprintf("procedures[%d].id", i), "duplicate procedure id in payload")
			}
			seenUpdateIDs[row.ID.Int64] = struct{}{}
		}
	}

	return inputSets{
		productIDs:    int64SetKeys(productIDSet),
		doctorIDs:     stringSetKeys(doctorIDSet),
		nurseIDs:      stringSetKeys(nurseIDSet),
		icd9cmCodes:   stringSetKeys(icd9cmCodeSet),
		seenUpdateIDs: seenUpdateIDs,
	}
}

// validateProducts verifies that every product_id belongs to the institution and has
// is_treatment = true. Returns a map of id → name for snapshotting.
func (u *ProcedureUC) validateProducts(
	ctx context.Context,
	institutionID int64,
	productIDs []int64,
	rows []model.SaveProcedureRow,
	errMsg *commonerr.ErrorMessage,
) (map[int64]string, error) {
	productNames := make(map[int64]string)
	if len(productIDs) == 0 {
		return productNames, nil
	}

	products, err := u.InstitutionRepo.FindTrxInstitutionProductByParams(ctx, model.FindTrxInstitutionProductParams{
		IDs:              productIDs,
		IDMstInstitution: institutionID,
		IsTreatment:      true,
	})
	if err != nil {
		return nil, err
	}

	found := make(map[int64]string, len(products))
	for _, p := range products {
		found[p.ID] = p.Name
	}
	for _, pid := range productIDs {
		if name, ok := found[pid]; ok {
			productNames[pid] = name
		} else {
			for i, row := range rows {
				if row.ProductID.Valid && row.ProductID.Int64 == pid {
					errMsg.Append(fmt.Sprintf("procedures[%d].product_id", i), "product_id does not exist as a treatment product in this institution")
				}
			}
		}
	}
	return productNames, nil
}

// validateDoctors checks that every doctor_id in the payload exists in the institution.
func (u *ProcedureUC) validateDoctors(
	ctx context.Context,
	institutionID int64,
	doctorIDs []string,
	rows []model.SaveProcedureRow,
	errMsg *commonerr.ErrorMessage,
) error {
	missing, err := u.PractitionerDB.MissingDoctorIDs(ctx, institutionID, doctorIDs)
	if err != nil {
		return err
	}
	missingSet := stringSliceToSet(missing)
	for i, row := range rows {
		if _, miss := missingSet[row.DoctorID]; miss {
			errMsg.Append(fmt.Sprintf("procedures[%d].doctor_id", i), "doctor_id does not exist")
		}
	}
	return nil
}

// validateNurses checks that every nurse_id in the payload exists in the institution.
// Skips the DB call when no nurse IDs were provided.
func (u *ProcedureUC) validateNurses(
	ctx context.Context,
	institutionID int64,
	nurseIDs []string,
	rows []model.SaveProcedureRow,
	errMsg *commonerr.ErrorMessage,
) error {
	if len(nurseIDs) == 0 {
		return nil
	}
	missing, err := u.PractitionerDB.MissingNurseIDs(ctx, institutionID, nurseIDs)
	if err != nil {
		return err
	}
	missingSet := stringSliceToSet(missing)
	for i, row := range rows {
		if row.NurseID.Valid && row.NurseID.String != "" {
			if _, miss := missingSet[row.NurseID.String]; miss {
				errMsg.Append(fmt.Sprintf("procedures[%d].nurse_id", i), "nurse_id does not exist")
			}
		}
	}
	return nil
}

// validateAndResolveICD9CM checks that each code exists in mdl_ref_icd9cm and returns
// a map of code → display for snapshotting.
func (u *ProcedureUC) validateAndResolveICD9CM(
	ctx context.Context,
	codes []string,
	rows []model.SaveProcedureRow,
	errMsg *commonerr.ErrorMessage,
) (map[string]string, error) {
	displays := make(map[string]string)
	for _, code := range codes {
		node, err := u.ICD9CMDB.GetByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if node == nil {
			for i, row := range rows {
				if row.ICD9CMCode.Valid && row.ICD9CMCode.String == code {
					errMsg.Append(fmt.Sprintf("procedures[%d].icd9cm_code", i), "icd9cm_code not found in reference table")
				}
			}
		} else {
			displays[code] = node.Display
		}
	}
	return displays, nil
}

// resolveSnapshotNames fetches the display names for doctors and nurses so they can be
// written into snapshot columns on the transaction row.
func (u *ProcedureUC) resolveSnapshotNames(
	ctx context.Context,
	institutionID int64,
	doctorIDs []string,
	nurseIDs []string,
) (snapshotMaps, error) {
	snap := snapshotMaps{
		doctorNames: make(map[string]string),
		nurseNames:  make(map[string]string),
	}

	doctors, err := u.PractitionerDB.GetDoctorsByIDs(ctx, institutionID, doctorIDs)
	if err != nil {
		return snap, err
	}
	for _, d := range doctors {
		snap.doctorNames[d.ID] = d.Name
	}

	if len(nurseIDs) > 0 {
		nurses, err := u.PractitionerDB.GetNursesByIDs(ctx, institutionID, nurseIDs)
		if err != nil {
			return snap, err
		}
		for _, n := range nurses {
			snap.nurseNames[n.ID] = n.Name
		}
	}

	return snap, nil
}

// diffProcedureRows compares the payload against existing rows and produces three
// disjoint slices: rows to insert, rows to update, and IDs to soft-delete.
func diffProcedureRows(
	payload []model.SaveProcedureRow,
	existingRows []model.TrxVisitProcedure,
	visitID int64,
	institutionID int64,
	snap snapshotMaps,
	errMsg *commonerr.ErrorMessage,
) (toInsert, toUpdate []model.TrxVisitProcedure, toSoftDelete []int64, err error) {
	existingMap := make(map[int64]model.TrxVisitProcedure, len(existingRows))
	for _, row := range existingRows {
		existingMap[row.ID] = row
	}

	requestedIDs := make(map[int64]struct{})
	for i, item := range payload {
		row := buildTrxRow(item, visitID, institutionID, snap)

		if !item.ID.Valid {
			toInsert = append(toInsert, row)
			continue
		}
		if _, ok := existingMap[item.ID.Int64]; !ok {
			errMsg.Append(fmt.Sprintf("procedures[%d].id", i), "procedure id not found for this visit")
			continue
		}
		row.ID = item.ID.Int64
		requestedIDs[row.ID] = struct{}{}
		toUpdate = append(toUpdate, row)
	}

	for id := range existingMap {
		if _, keep := requestedIDs[id]; !keep {
			toSoftDelete = append(toSoftDelete, id)
		}
	}
	return
}

// persistAtomically runs soft-delete, bulk-insert, and bulk-update inside a single DB
// transaction. The named err pointer is used by the deferred Finish to roll back on panic.
func (u *ProcedureUC) persistAtomically(
	ctx context.Context,
	institutionID, visitID int64,
	toInsert, toUpdate []model.TrxVisitProcedure,
	toSoftDelete []int64,
	errPtr *error,
) error {
	session, beginErr := u.Transaction.Begin(ctx)
	if beginErr != nil {
		return errors.Wrap(beginErr, wrapMsgSave)
	}
	defer u.Transaction.Finish(session, errPtr)
	txCtx := xormlib.SetDBSession(ctx, session)

	if err := u.ProcedureDB.SoftDeleteByIDs(txCtx, institutionID, visitID, toSoftDelete); err != nil {
		*errPtr = errors.Wrap(err, wrapMsgSave)
		return *errPtr
	}
	if err := u.ProcedureDB.BulkInsert(txCtx, toInsert); err != nil {
		*errPtr = errors.Wrap(err, wrapMsgSave)
		return *errPtr
	}
	if err := u.ProcedureDB.BulkUpdate(txCtx, toUpdate); err != nil {
		*errPtr = errors.Wrap(err, wrapMsgSave)
		return *errPtr
	}
	return nil
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (u *ProcedureUC) authorizeVisit(ctx context.Context, visitID int64) (model.UserJWTPayload, error) {
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

// buildTrxRow constructs a TrxVisitProcedure from a request row and resolved snapshot maps.
func buildTrxRow(
	item model.SaveProcedureRow,
	visitID int64,
	institutionID int64,
	snap snapshotMaps,
) model.TrxVisitProcedure {
	row := model.TrxVisitProcedure{
		VisitID:       visitID,
		InstitutionID: institutionID,
		DoctorID:      item.DoctorID,
		DoctorName:    snap.doctorNames[item.DoctorID],
		Rank:          item.Rank,
	}

	if item.ProductID.Valid && item.ProductID.Int64 > 0 {
		row.ProductID = sql.NullInt64{Int64: item.ProductID.Int64, Valid: true}
		row.ProductName = sql.NullString{String: snap.productNames[item.ProductID.Int64], Valid: true}
	}
	if item.NurseID.Valid && item.NurseID.String != "" {
		row.NurseID = sql.NullString{String: item.NurseID.String, Valid: true}
		row.NurseName = sql.NullString{String: snap.nurseNames[item.NurseID.String], Valid: true}
	}
	if item.PlannedAt.Valid && item.PlannedAt.String != "" {
		if t, err := time.Parse(time.RFC3339, item.PlannedAt.String); err == nil {
			row.PlannedAt = &t
		}
	}
	if item.Category.Valid {
		row.Category = sql.NullString{String: item.Category.String, Valid: true}
	}
	if item.Duration.Valid {
		row.Duration = sql.NullString{String: item.Duration.String, Valid: true}
	}
	if item.ICD9CMCode.Valid && item.ICD9CMCode.String != "" {
		row.ICD9CMCode = sql.NullString{String: item.ICD9CMCode.String, Valid: true}
		row.ICD9CMDisplay = sql.NullString{String: snap.icd9cmDisplays[item.ICD9CMCode.String], Valid: true}
	}
	if item.Description.Valid {
		row.Description = sql.NullString{String: item.Description.String, Valid: true}
	}
	if item.Notes.Valid {
		row.Notes = sql.NullString{String: item.Notes.String, Valid: true}
	}
	return row
}

func int64SetKeys(m map[int64]struct{}) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func stringSetKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func stringSliceToSet(s []string) map[string]struct{} {
	out := make(map[string]struct{}, len(s))
	for _, v := range s {
		out[v] = struct{}{}
	}
	return out
}
