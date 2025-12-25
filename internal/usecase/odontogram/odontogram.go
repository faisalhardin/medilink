package odontogram

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	odontogramrepo "github.com/faisalhardin/medilink/internal/entity/repo/odontogram"
	patientrepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	odontogramuc "github.com/faisalhardin/medilink/internal/entity/usecase/odontogram"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	authmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// OdontogramUC implements the odontogram use case
type OdontogramUC struct {
	OdontogramDB odontogramrepo.OdontogramRepo
	PatientDB    patientrepo.PatientDB
	Cache        *SnapshotCache
	Transaction  xorm.DBTransactionInterface
}

// New creates a new odontogram use case instance
func New(opt OdontogramUC) odontogramuc.OdontogramUC {
	return &opt
}

// CreateEvents creates one or more odontogram events
func (uc *OdontogramUC) CreateEvents(ctx context.Context, requests []model.CreateOdontogramEventRequest) (*model.CreateOdontogramEventsResponse, error) {
	if len(requests) == 0 {
		return nil, commonerr.SetNewBadRequest("empty request", "At least one event is required")
	}

	// Get user from context
	userDetail, ok := authmodule.GetUserDetailFromCtx(ctx)
	if !ok {
		return nil, commonerr.SetNewUnauthorizedError("unauthorized", "User not authenticated")
	}

	// Get patient ID from first request (all should have same patient)
	patientUUID := requests[0].PatientUUID
	mstPatient, err := uc.PatientDB.GetPatientByParams(ctx, model.MstPatientInstitution{
		UUID:          patientUUID,
		InstitutionID: userDetail.InstitutionID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get patient ID")
	}

	// Validate and prepare events
	events := make([]*model.HstOdontogram, 0, len(requests))
	for _, req := range requests {
		// Validate request
		if err := ValidateEventRequest(req); err != nil {
			return nil, err
		}

		// Ensure all events are for the same patient
		if req.PatientUUID != patientUUID {
			return nil, commonerr.SetNewBadRequest("patient mismatch", "All events must be for the same patient")
		}

		// Marshal event data
		eventDataJSON, err := json.Marshal(req.EventData)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal event data")
		}

		// Create event model
		eventID := req.EventID
		if eventID != "" {
			newEventID, _ := uuid.FromString(eventID)
			eventID = newEventID.String()
		}

		unixTimestamp := req.UnixTimestamp
		if unixTimestamp == 0 {
			unixTimestamp = time.Now().Unix()
		}

		event := &model.HstOdontogram{
			EventID:             eventID,
			PatientID:           mstPatient.ID,
			InstitutionID:       mstPatient.InstitutionID,
			VisitID:             req.VisitID,
			JourneyPointShortID: req.JourneyPointID,
			EventType:           req.EventType,
			ToothID:             req.ToothID,
			EventData:           eventDataJSON,
			CreatedByStaffID:    userDetail.UserID,
			UnixTimestamp:       unixTimestamp,
			CreatedBy:           userDetail.Email,
			CreateTime:          time.Now().Unix(),
		}

		events = append(events, event)
	}

	session, _ := uc.Transaction.Begin(ctx)
	defer uc.Transaction.Finish(session, &err)
	ctx = xorm.SetDBSession(ctx, session)

	// Insert events
	if err := uc.OdontogramDB.InsertEventsBatch(ctx, events); err != nil {
		return nil, errors.Wrap(err, "failed to insert events")
	}

	// Invalidate cache
	if err := uc.Cache.Invalidate(ctx, patientUUID, requests[0].VisitID); err != nil {
		// Log but don't fail - cache invalidation is not critical
		log.Errorf("failed to invalidate cache for patient %s: %v", patientUUID, err)
	}

	// Update snapshot incrementally
	if err := uc.updateSnapshotIncremental(ctx, mstPatient.ID, events); err != nil {
		// Log but don't fail - snapshot can be rebuilt on next read
		log.Errorf("failed to update snapshot incrementally for patient %d: %v", mstPatient, err)
	}

	// Prepare response
	results := make([]model.CreateOdontogramEventResponse, len(events))
	var maxLogicalTimestamp, maxSequenceNumber int64
	for i, event := range events {
		results[i] = model.CreateOdontogramEventResponse{
			EventID:          event.EventID,
			SequenceNumber:   event.SequenceNumber,
			LogicalTimestamp: event.LogicalTimestamp,
			Status:           "created",
		}
		if event.LogicalTimestamp > maxLogicalTimestamp {
			maxLogicalTimestamp = event.LogicalTimestamp
		}
		if event.SequenceNumber > maxSequenceNumber {
			maxSequenceNumber = event.SequenceNumber
		}
	}

	return &model.CreateOdontogramEventsResponse{
		Results:             results,
		MaxLogicalTimestamp: maxLogicalTimestamp,
		MaxSequenceNumber:   maxSequenceNumber,
	}, nil
}

// GetEvents retrieves events for a patient with filtering
func (uc *OdontogramUC) GetEvents(ctx context.Context, params model.GetOdontogramEventsParams) (*model.GetOdontogramEventsResponse, error) {
	// Validate params
	if err := ValidateGetEventsParams(params); err != nil {
		return nil, err
	}

	// Get institution ID
	userDetail, ok := authmodule.GetUserDetailFromCtx(ctx)
	if !ok {
		return nil, commonerr.SetNewUnauthorizedError("unauthorized", "User not authenticated")
	}

	// Get patient ID
	mstPatient, err := uc.PatientDB.GetPatientByParams(ctx, model.MstPatientInstitution{UUID: params.PatientUUID, InstitutionID: userDetail.InstitutionID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get patient ID")
	}

	// Get events
	var events []model.HstOdontogram
	if params.EventID != "" {
		// Get single event
		event, err := uc.OdontogramDB.GetEventByID(ctx, userDetail.InstitutionID, params.EventID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, commonerr.SetNewBadRequest("event not found", "Event not found")
			}
			return nil, errors.Wrap(err, "failed to get event")
		}
		events = []model.HstOdontogram{*event}
	} else {
		// Get filtered events
		params.InstitutionID = userDetail.InstitutionID
		events, err = uc.OdontogramDB.GetEventsByPatientFiltered(ctx, params, mstPatient.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get events")
		}
	}

	// Get max values
	maxSeq, _ := uc.OdontogramDB.GetMaxSequenceNumber(ctx, userDetail.InstitutionID, mstPatient.ID)
	maxTimestamp, _ := uc.OdontogramDB.GetMaxLogicalTimestamp(ctx, userDetail.InstitutionID, mstPatient.ID)

	// Convert to response format
	responseEvents := make([]model.HstOdontogramResponse, len(events))
	for i, event := range events {
		var eventData model.OdontogramEventData
		json.Unmarshal([]byte(event.EventData), &eventData)

		responseEvents[i] = model.HstOdontogramResponse{
			EventID:          event.EventID,
			VisitID:          event.VisitID,
			JourneyPointID:   event.JourneyPointShortID,
			EventType:        event.EventType,
			ToothID:          event.ToothID,
			PatientUUID:      params.PatientUUID,
			EventData:        eventData,
			SequenceNumber:   event.SequenceNumber,
			LogicalTimestamp: event.LogicalTimestamp,
			UnixTimestamp:    event.UnixTimestamp,
			CreatedBy:        event.CreatedBy,
		}
	}

	return &model.GetOdontogramEventsResponse{
		Events:              responseEvents,
		MaxLogicalTimestamp: maxTimestamp,
		MaxSequenceNumber:   maxSeq,
		Total:               len(responseEvents),
	}, nil
}

// GetSnapshot retrieves the current or historical snapshot for a patient
func (uc *OdontogramUC) GetSnapshot(ctx context.Context, params model.GetOdontogramSnapshotParams) (*model.GetOdontogramSnapshotResponse, error) {

	userDetail, ok := authmodule.GetUserDetailFromCtx(ctx)
	if !ok {
		return nil, commonerr.SetNewUnauthorizedError("unauthorized", "User not authenticated")
	}

	// Get patient ID
	mstPatient, err := uc.PatientDB.GetPatientByParams(ctx, model.MstPatientInstitution{
		UUID:          params.PatientUUID,
		InstitutionID: userDetail.InstitutionID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get patient ID")
	}

	// Check if requesting historical snapshot
	if params.SequenceNumber > 0 {
		return uc.getHistoricalSnapshot(ctx, mstPatient.ID, userDetail.InstitutionID, params.SequenceNumber)
	}

	// Try L1 cache (in-memory)
	cached, err := uc.Cache.Get(ctx, params.PatientUUID, params.VisitID)
	if err == nil {
		return &model.GetOdontogramSnapshotResponse{
			Snapshot:            cached.Snapshot,
			MaxLogicalTimestamp: cached.MaxLogicalTimestamp,
			MaxSequenceNumber:   cached.MaxSequenceNumber,
			LastUpdated:         cached.LastUpdated,
		}, nil
	}

	// Execute database calls in parallel using errgroup
	var (
		maxSeq   int64
		snapshot *model.MstPatientOdontogram
	)

	g, gctx := errgroup.WithContext(ctx)

	// Get max sequence number in parallel
	g.Go(func() error {
		var err error
		maxSeq, err = uc.OdontogramDB.GetMaxSequenceNumber(gctx, userDetail.InstitutionID, mstPatient.ID)
		if err != nil {
			return errors.Wrap(err, "failed to get max sequence number")
		}
		return nil
	})

	// Get snapshot in parallel
	g.Go(func() error {
		var err error
		snapshot, err = uc.OdontogramDB.GetSnapshot(gctx, userDetail.InstitutionID, mstPatient.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "failed to get snapshot")
		}
		return nil
	})

	// Wait for both goroutines to complete and check for errors
	if err := g.Wait(); err != nil {
		return nil, err
	}

	if snapshot != nil && snapshot.LastEventSequence == maxSeq {
		var snapshotData model.OdontogramSnapshot
		if err := json.Unmarshal([]byte(snapshot.Snapshot), &snapshotData); err == nil {

			uc.Cache.Set(ctx, params.PatientUUID, params.VisitID, snapshotData, snapshot.MaxLogicalTimestamp, snapshot.LastEventSequence, snapshot.LastUpdated)

			return &model.GetOdontogramSnapshotResponse{
				Snapshot:            snapshotData,
				MaxLogicalTimestamp: snapshot.MaxLogicalTimestamp,
				MaxSequenceNumber:   snapshot.LastEventSequence,
				LastUpdated:         snapshot.LastUpdated,
			}, nil
		}
	} else if snapshot == nil && maxSeq == 0 {
		return &model.GetOdontogramSnapshotResponse{
			Snapshot:            model.OdontogramSnapshot{},
			MaxLogicalTimestamp: 0,
			MaxSequenceNumber:   0,
			LastUpdated:         0,
		}, nil
	}

	return uc.rebuildAndCacheSnapshot(ctx, mstPatient.ID, userDetail.InstitutionID, params.PatientUUID, params.VisitID)
}

// rebuildAndCacheSnapshot rebuilds snapshot from events and updates both L1 and L2 caches
func (uc *OdontogramUC) rebuildAndCacheSnapshot(ctx context.Context, patientID int64, institutionID int64, patientUUID string, visitID int64) (*model.GetOdontogramSnapshotResponse, error) {
	// Rebuild from events
	snapshotData, maxTimestamp, maxSeq, err := uc.BuildSnapshot(ctx, model.GetEventsByPatientParams{
		PatientID:     patientID,
		InstitutionID: institutionID,
		FromSequence:  0,
		ToSequence:    0,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to build snapshot")
	}

	lastUpdated := time.Now().Unix()

	// Update L2 cache (Postgres)
	snapshotJSON, _ := json.Marshal(snapshotData)
	err = uc.OdontogramDB.UpsertSnapshot(ctx, &model.MstPatientOdontogram{
		PatientID:           patientID,
		InstitutionID:       institutionID,
		Snapshot:            snapshotJSON,
		LastEventSequence:   maxSeq,
		MaxLogicalTimestamp: maxTimestamp,
		LastUpdated:         lastUpdated,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert snapshot")
	}

	// Update L1 cache
	_ = uc.Cache.Set(ctx, patientUUID, visitID, *snapshotData, maxTimestamp, maxSeq, lastUpdated)

	return &model.GetOdontogramSnapshotResponse{
		Snapshot:            *snapshotData,
		MaxLogicalTimestamp: maxTimestamp,
		MaxSequenceNumber:   maxSeq,
		LastUpdated:         lastUpdated,
	}, nil
}

// BuildSnapshot builds a snapshot from events up to a specific sequence number
func (uc *OdontogramUC) BuildSnapshot(ctx context.Context, params model.GetEventsByPatientParams) (*model.OdontogramSnapshot, int64, int64, error) {

	if params.InstitutionID == 0 {
		userDetail, ok := authmodule.GetUserDetailFromCtx(ctx)
		if !ok {
			return nil, 0, 0, commonerr.SetNewUnauthorizedError("unauthorized", "User not authenticated")
		}
		params.InstitutionID = userDetail.InstitutionID
	}

	// Get all events up to sequence number
	events, err := uc.OdontogramDB.GetEventsByPatientFiltered(ctx, model.GetOdontogramEventsParams{
		InstitutionID: params.InstitutionID,
		PatientID:     params.PatientID,
		FromSequence:  params.FromSequence,
		ToSequence:    params.ToSequence,
		VisitID:       params.VisitID,
	}, params.PatientID)
	if err != nil {
		return nil, 0, 0, errors.Wrap(err, "failed to get events")
	}

	// Build snapshot using CRDT
	builder := NewSnapshotBuilder(events)
	snapshot, err := builder.Build()
	if err != nil {
		return nil, 0, 0, errors.Wrap(err, "failed to build snapshot")
	}

	// Get max values
	var maxTimestamp, maxSeq int64
	for _, event := range events {
		if event.LogicalTimestamp > maxTimestamp {
			maxTimestamp = event.LogicalTimestamp
		}
		if event.SequenceNumber > maxSeq {
			maxSeq = event.SequenceNumber
		}
	}

	return snapshot, maxTimestamp, maxSeq, nil
}

// getHistoricalSnapshot gets a snapshot at a specific point in time
func (uc *OdontogramUC) getHistoricalSnapshot(ctx context.Context, patientID int64, institutionID int64, sequenceNumber int64) (*model.GetOdontogramSnapshotResponse, error) {
	snapshot, maxTimestamp, maxSeq, err := uc.BuildSnapshot(ctx, model.GetEventsByPatientParams{
		PatientID:     patientID,
		InstitutionID: institutionID,
		FromSequence:  sequenceNumber,
		ToSequence:    sequenceNumber,
	})
	if err != nil {
		return nil, err
	}

	return &model.GetOdontogramSnapshotResponse{
		Snapshot:            *snapshot,
		MaxLogicalTimestamp: maxTimestamp,
		MaxSequenceNumber:   maxSeq,
		LastUpdated:         time.Now().Unix(),
	}, nil
}

// updateSnapshotIncremental updates the snapshot incrementally with new events
func (uc *OdontogramUC) updateSnapshotIncremental(ctx context.Context, patientID int64, newEvents []*model.HstOdontogram) error {

	if len(newEvents) == 0 {
		return nil
	}

	// Get institution ID
	userDetail, ok := authmodule.GetUserDetailFromCtx(ctx)
	if !ok {
		return commonerr.SetNewUnauthorizedError("unauthorized", "User not authenticated")
	}

	// Get current snapshot
	existingSnapshot, err := uc.OdontogramDB.GetSnapshot(ctx, userDetail.InstitutionID, patientID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	var currentSnapshot model.OdontogramSnapshot
	allEvents := make([]model.HstOdontogram, 0)

	for _, event := range newEvents {
		allEvents = append(allEvents, *event)
	}

	earliestEventSequence := int64(math.MaxInt64)
	for _, event := range newEvents {
		if event.SequenceNumber < earliestEventSequence {
			earliestEventSequence = event.SequenceNumber
		}
	}

	if existingSnapshot != nil {
		err = json.Unmarshal([]byte(existingSnapshot.Snapshot), &currentSnapshot)
		if err != nil {
			existingSnapshot = nil
		}

	} // find if there is a gap in the events
	if existingSnapshot != nil && existingSnapshot.LastEventSequence != earliestEventSequence-1 {
		// Get events since last snapshot
		events, err := uc.OdontogramDB.GetEventsByPatient(ctx, model.GetEventsByPatientParams{
			PatientID:     patientID,
			InstitutionID: userDetail.InstitutionID,
			FromSequence:  existingSnapshot.LastEventSequence + 1,
			ToSequence:    earliestEventSequence - 1, // to cater replication gap
		})
		if err != nil {
			return err
		}
		allEvents = append(allEvents, events...)

	}
	// if there is no existing snapshot, get all events
	if existingSnapshot == nil {
		// No existing snapshot, get all events
		events, err := uc.OdontogramDB.GetEventsByPatient(ctx, model.GetEventsByPatientParams{
			PatientID:     patientID,
			InstitutionID: userDetail.InstitutionID,
			FromSequence:  0,
			ToSequence:    earliestEventSequence - 1, // to cater replication gap
		})
		if err != nil {
			return err
		}
		allEvents = append(allEvents, events...)
	}

	// Build snapshot
	builder := NewSnapshotBuilder(allEvents)
	if existingSnapshot != nil {
		builder.FromSnapshot(&currentSnapshot)
	}
	snapshot, err := builder.Build()
	if err != nil {
		return err
	}

	var lastEventSequence int64
	var lastEventLogicalTimestamp int64
	if len(builder.events) > 0 {
		lastEventSequence = builder.events[len(builder.events)-1].SequenceNumber
		lastEventLogicalTimestamp = builder.events[len(builder.events)-1].LogicalTimestamp
	} else {
		lastEventSequence = 0
		lastEventLogicalTimestamp = 0
	}

	// Update snapshot in database
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	return uc.OdontogramDB.UpsertSnapshot(ctx, &model.MstPatientOdontogram{
		InstitutionID:       userDetail.InstitutionID,
		PatientID:           patientID,
		Snapshot:            snapshotJSON,
		LastEventSequence:   lastEventSequence,
		MaxLogicalTimestamp: lastEventLogicalTimestamp,
		LastUpdated:         time.Now().Unix(),
	})
}
