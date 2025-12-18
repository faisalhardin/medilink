package odontogram

import (
	"context"
	"database/sql"

	"github.com/faisalhardin/medilink/internal/entity/model"
	odontogramrepo "github.com/faisalhardin/medilink/internal/entity/repo/odontogram"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

var (
	emptyEventID = uuid.Nil.String()
)

const (
	WrapErrMsgPrefix              = "OdontogramDB."
	WrapMsgInsertEvent            = WrapErrMsgPrefix + "InsertEvent"
	WrapMsgInsertEventsBatch      = WrapErrMsgPrefix + "InsertEventsBatch"
	WrapMsgGetEventsByPatient     = WrapErrMsgPrefix + "GetEventsByPatient"
	WrapMsgGetEventByID           = WrapErrMsgPrefix + "GetEventByID"
	WrapMsgGetMaxSequenceNumber   = WrapErrMsgPrefix + "GetMaxSequenceNumber"
	WrapMsgGetMaxLogicalTimestamp = WrapErrMsgPrefix + "GetMaxLogicalTimestamp"
	WrapMsgGetSnapshot            = WrapErrMsgPrefix + "GetSnapshot"
	WrapMsgUpsertSnapshot         = WrapErrMsgPrefix + "UpsertSnapshot"
)

type OdontogramDB struct {
	DB *xormlib.DBConnect
}

func NewOdontogramDB(db *xormlib.DBConnect) odontogramrepo.OdontogramRepo {
	return &OdontogramDB{DB: db}
}

// InsertEvent inserts a single event with automatic sequence and logical timestamp assignment
func (o *OdontogramDB) InsertEvent(ctx context.Context, event *model.HstOdontogram) error {
	session := xorm.GetDBSession(ctx)
	if session == nil {
		session = o.DB.MasterDB.Context(ctx)
	}

	// Get max sequence number for this patient
	maxSeq, err := o.GetMaxSequenceNumber(ctx, event.InstitutionID, event.PatientID)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertEvent)
	}
	event.SequenceNumber = maxSeq + 1

	// Get max logical timestamp for this patient
	maxTimestamp, err := o.GetMaxLogicalTimestamp(ctx, event.InstitutionID, event.PatientID)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertEvent)
	}
	event.LogicalTimestamp = maxTimestamp + 1

	// Generate UUID if not provided
	if event.EventID == "" {
		newEventID, _ := uuid.NewV4()
		event.EventID = newEventID.String()
	}

	// Insert event
	_, err = session.Insert(event)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertEvent)
	}

	return nil
}

// InsertEventsBatch inserts multiple events atomically
func (o *OdontogramDB) InsertEventsBatch(ctx context.Context, events []*model.HstOdontogram) error {
	if len(events) == 0 {
		return nil
	}

	session := xorm.GetDBSession(ctx)
	if session == nil {
		session = o.DB.MasterDB.Context(ctx)
	}

	// Get initial max values
	patientID := events[0].PatientID
	institutionID := events[0].InstitutionID
	maxSeq, err := o.GetMaxSequenceNumber(ctx, institutionID, patientID)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertEventsBatch)
	}

	maxTimestamp, err := o.GetMaxLogicalTimestamp(ctx, institutionID, patientID)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertEventsBatch)
	}

	// Assign sequence numbers and timestamps
	for i, event := range events {
		event.SequenceNumber = maxSeq + int64(i) + 1
		event.LogicalTimestamp = maxTimestamp + int64(i) + 1

		if event.EventID == "" || event.EventID == emptyEventID {
			newEventID, _ := uuid.NewV4()
			event.EventID = newEventID.String()
		}

		events[i] = event

	}

	_, err = session.Insert(events)
	if err != nil {
		return errors.Wrap(err, WrapMsgInsertEventsBatch)
	}

	return nil
}

// GetEventsByPatient retrieves events for a patient with pagination
func (o *OdontogramDB) GetEventsByPatient(ctx context.Context, params model.GetEventsByPatientParams) ([]model.HstOdontogram, error) {
	session := o.DB.SlaveDB

	query := session.Where("patient_id = ?", params.PatientID)
	query = query.And("institution_id = ?", params.InstitutionID)

	if params.FromSequence > 0 {
		query = query.And("sequence_number >= ?", params.FromSequence)
	}

	if params.ToSequence > 0 {
		query = query.And("sequence_number <= ?", params.ToSequence)
	}

	query = query.OrderBy("logical_timestamp ASC, created_by_staff_id ASC")

	if params.Limit > 0 {
		query = query.Limit(params.Limit, params.Offset)
	}

	var events []model.HstOdontogram
	err := query.Find(&events)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetEventsByPatient)
	}

	return events, nil
}

// GetEventsByPatientFiltered retrieves events with additional filtering
func (o *OdontogramDB) GetEventsByPatientFiltered(ctx context.Context, params model.GetOdontogramEventsParams, patientID int64) ([]model.HstOdontogram, error) {
	session := o.DB.SlaveDB

	query := session.Where("patient_id = ?", patientID)

	query = query.And("institution_id = ?", params.InstitutionID)

	if params.ToothID != "" {
		query = query.And("tooth_id = ?", params.ToothID)
	}

	if params.EventType != "" {
		query = query.And("event_type = ?", params.EventType)
	}

	if params.VisitID > 0 {
		query = query.And("visit_id = ?", params.VisitID)
	}

	if params.FromSequence > 0 {
		query = query.And("sequence_number >= ?", params.FromSequence)
	}

	if params.ToSequence > 0 {
		query = query.And("sequence_number <= ?", params.ToSequence)
	}

	query = query.OrderBy("logical_timestamp ASC, created_by_staff_id ASC")

	limit := params.Limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	query = query.Limit(limit, params.Offset)

	var events []model.HstOdontogram
	err := query.Find(&events)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetEventsByPatient)
	}

	return events, nil
}

// GetEventByID retrieves a single event by ID
func (o *OdontogramDB) GetEventByID(ctx context.Context, institutionID int64, eventID string) (*model.HstOdontogram, error) {
	session := o.DB.SlaveDB

	var event model.HstOdontogram
	found, err := session.Where("event_id = ?", eventID).
		And("institution_id = ?", institutionID).
		Get(&event)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetEventByID)
	}

	if !found {
		return nil, sql.ErrNoRows
	}

	return &event, nil
}

// GetMaxSequenceNumber returns the maximum sequence number for a patient
func (o *OdontogramDB) GetMaxSequenceNumber(ctx context.Context, institutionID, patientID int64) (int64, error) {
	session := o.DB.SlaveDB

	var result struct {
		MaxSeq int64 `xorm:"max_seq"`
	}

	_, err := session.SQL("SELECT COALESCE(MAX(sequence_number), 0) as max_seq FROM mdl_hst_odontogram WHERE patient_id = ? and institution_id = ?", patientID, institutionID).Get(&result)
	if err != nil {
		return 0, errors.Wrap(err, WrapMsgGetMaxSequenceNumber)
	}

	return result.MaxSeq, nil
}

// GetMaxLogicalTimestamp returns the maximum logical timestamp for a patient
func (o *OdontogramDB) GetMaxLogicalTimestamp(ctx context.Context, institutionID, patientID int64) (int64, error) {
	session := o.DB.SlaveDB

	var result struct {
		MaxTimestamp int64 `xorm:"max_timestamp"`
	}

	_, err := session.SQL("SELECT COALESCE(MAX(logical_timestamp), 0) as max_timestamp FROM mdl_hst_odontogram WHERE patient_id = ? and institution_id = ?", patientID, institutionID).Get(&result)
	if err != nil {
		return 0, errors.Wrap(err, WrapMsgGetMaxLogicalTimestamp)
	}

	return result.MaxTimestamp, nil
}

// GetSnapshot retrieves the snapshot for a patient
func (o *OdontogramDB) GetSnapshot(ctx context.Context, institutionID, patientID int64) (*model.MstPatientOdontogram, error) {
	session := o.DB.SlaveDB

	var snapshot model.MstPatientOdontogram
	found, err := session.
		Where("institution_id = ?", institutionID).
		And("patient_id = ?", patientID).Get(&snapshot)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetSnapshot)
	}

	if !found {
		return nil, sql.ErrNoRows
	}

	return &snapshot, nil
}

// UpsertSnapshot updates or inserts a snapshot
func (o *OdontogramDB) UpsertSnapshot(ctx context.Context, snapshot *model.MstPatientOdontogram) error {
	session := xorm.GetDBSession(ctx)
	if session == nil {
		session = o.DB.MasterDB.Context(ctx)
	}

	// Generate UUID if not provided
	if snapshot.ID == "" {
		snapshotID, _ := uuid.NewV4()
		snapshot.ID = snapshotID.String()
	}

	// Try to get existing snapshot
	existing, err := o.GetSnapshot(ctx, snapshot.InstitutionID, snapshot.PatientID)
	if err != nil && err != sql.ErrNoRows {
		return errors.Wrap(err, WrapMsgUpsertSnapshot)
	}

	if existing != nil {
		// Update existing
		snapshot.ID = existing.ID
		_, err = session.Where("patient_id = ? and institution_id = ?", snapshot.PatientID, snapshot.InstitutionID).Update(snapshot)
		if err != nil {
			return errors.Wrap(err, WrapMsgUpsertSnapshot)
		}
	} else {
		// Insert new
		_, err = session.Insert(snapshot)
		if err != nil {
			return errors.Wrap(err, WrapMsgUpsertSnapshot)
		}
	}

	return nil
}
