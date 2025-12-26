package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	ilog "github.com/faisalhardin/medilink/cmd/log"
	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	liblog "github.com/faisalhardin/medilink/internal/library/common/log"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	journeyrepo "github.com/faisalhardin/medilink/internal/repo/journey"
	odontogramrepo "github.com/faisalhardin/medilink/internal/repo/odontogram"
	patientrepo "github.com/faisalhardin/medilink/internal/repo/patient"
	staffrepo "github.com/faisalhardin/medilink/internal/repo/staff"
)

const (
	repoName = "medilink"
)

func main() {

	cfg, err := config.New(repoName)
	if err != nil {
		log.Fatalf("failed to init the config: %v", err)
	}

	vault, err := config.NewVault()
	if err != nil {
		log.Fatalf("failed to init the vault: %v", err)
	}

	cfg.Vault = vault.Data

	ilog.SetupLogging(cfg)

	db, err := xormlib.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
		return
	}
	defer db.CloseDBConnection()

	// Initialize repositories
	journeyDB := journeyrepo.NewJourneyDB(&journeyrepo.JourneyDB{
		DB: db,
	})

	patientConn := &patientrepo.Conn{DB: db}
	patientDB := patientrepo.NewPatientDB(patientConn)

	odontogramDB := odontogramrepo.NewOdontogramDB(db).(*odontogramrepo.OdontogramDB)

	staffConn := staffrepo.Conn{DB: db}
	staffRepo := staffrepo.New(staffConn)

	// Run the odontogram generation job
	generateOdontogramFromNotes(patientDB, odontogramDB, &staffRepo, journeyDB)

	// Optionally run other jobs
	// fillShortID(journeyDB)
	_ = journeyDB // Suppress unused variable warning

}

func fillShortID(journeyDB *journeyrepo.JourneyDB) {
	journeyPoints, err := journeyDB.ListJourneyPointsWithoutShortID(context.Background(), model.GetJourneyPointParams{
		CommonRequestPayload: model.CommonRequestPayload{
			Limit:  1000,
			Offset: 0,
		},
	})
	if err != nil {
		log.Fatalf("failed to list journey points: %v", err)
	}

	liblog.Info("journey points: %v", journeyPoints)
	for _, journeyPoint := range journeyPoints {
		journeyPoint.BeforeInsert()
		err := journeyDB.UpdateJourneyPoint(context.Background(), &journeyPoint)
		if err != nil {
			log.Fatalf("failed to update journey point: %v", err)
		}
	}

}

// EditorJSData represents the structure of Editor.js JSON
type EditorJSData struct {
	Time    int64           `json:"time"`
	Blocks  []EditorJSBlock `json:"blocks"`
	Version string          `json:"version"`
}

// EditorJSBlock represents a block in Editor.js
type EditorJSBlock struct {
	ID   string                 `json:"id"`
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// OdontogramBlockData represents the data in an odontogram block
type OdontogramBlockData struct {
	Teeth map[string]ToothInfo `json:"teeth"`
}

// ToothInfo represents tooth information in odontogram block
type ToothInfo struct {
	ID             string        `json:"id"`
	Surfaces       []SurfaceInfo `json:"surfaces"`
	GeneralNotes   string        `json:"generalNotes"`
	WholeToothCode []string      `json:"wholeToothCode"`
}

// SurfaceInfo represents surface information for a tooth
type SurfaceInfo struct {
	Surface string `json:"surface"`
	Code    string `json:"code"`
	Notes   string `json:"notes,omitempty"`
}

// parseOdontogramBlocks extracts odontogram blocks from Editor.js notes JSON
func parseOdontogramBlocks(notes json.RawMessage) ([]OdontogramBlockData, int64, error) {
	var editorData EditorJSData
	if err := json.Unmarshal(notes, &editorData); err != nil {
		return nil, 0, err
	}

	var odontogramBlocks []OdontogramBlockData
	for _, block := range editorData.Blocks {
		if block.Type == "odontogram" {
			// Convert map[string]interface{} to OdontogramBlockData
			blockDataJSON, err := json.Marshal(block.Data)
			if err != nil {
				liblog.Error("failed to marshal block data: %v", err)
				continue
			}

			var odontogramData OdontogramBlockData
			if err := json.Unmarshal(blockDataJSON, &odontogramData); err != nil {
				liblog.Error("failed to unmarshal odontogram data: %v", err)
				continue
			}

			odontogramBlocks = append(odontogramBlocks, odontogramData)
		}
	}

	// Return the time field from Editor.js data (in milliseconds)
	return odontogramBlocks, editorData.Time, nil
}

// convertToothDataToEvents converts tooth data to HstOdontogram events
func convertToothDataToEvents(
	toothData OdontogramBlockData,
	visitInfo model.TrxPatientVisit,
	staffEmail string,
	staffID int64,
	journeyPointShortID string,
	unixTimestamp int64,
) []*model.HstOdontogram {
	var events []*model.HstOdontogram
	currentTime := time.Now().Unix()

	for toothID, tooth := range toothData.Teeth {
		// Create event for general notes if present
		if tooth.GeneralNotes != "" {
			eventData := model.OdontogramEventData{
				GeneralNotes: tooth.GeneralNotes,
			}
			eventDataJSON, err := json.Marshal(eventData)
			if err != nil {
				liblog.Error("failed to marshal general notes event data for tooth %s: %v", toothID, err)
				continue
			}

			event := &model.HstOdontogram{
				InstitutionID:       visitInfo.IDMstInstitution,
				PatientID:           visitInfo.IDMstPatient,
				VisitID:             visitInfo.ID,
				JourneyPointShortID: journeyPointShortID,
				EventType:           constant.EventTypeToothGeneralNoteUpdate,
				ToothID:             toothID,
				EventData:           eventDataJSON,
				CreatedByStaffID:    staffID,
				UnixTimestamp:       unixTimestamp,
				CreatedBy:           staffEmail,
				CreateTime:          currentTime,
			}
			events = append(events, event)
		}

		// Create events for whole tooth codes
		for _, code := range tooth.WholeToothCode {
			eventData := model.OdontogramEventData{
				WholeToothCode: []string{code},
			}
			eventDataJSON, err := json.Marshal(eventData)
			if err != nil {
				liblog.Error("failed to marshal tooth code event data for tooth %s: %v", toothID, err)
				continue
			}

			event := &model.HstOdontogram{
				InstitutionID:       visitInfo.IDMstInstitution,
				PatientID:           visitInfo.IDMstPatient,
				VisitID:             visitInfo.ID,
				JourneyPointShortID: journeyPointShortID,
				EventType:           constant.EventTypeToothCodeInsert,
				ToothID:             toothID,
				EventData:           eventDataJSON,
				CreatedByStaffID:    staffID,
				UnixTimestamp:       unixTimestamp,
				CreatedBy:           staffEmail,
				CreateTime:          currentTime,
			}
			events = append(events, event)
		}

		// Create events for surfaces
		for _, surface := range tooth.Surfaces {
			// Create event for surface code
			surfaceCodeData := model.OdontogramEventData{
				Surface:     surface.Surface,
				SurfaceCode: surface.Code,
			}
			surfaceCodeJSON, err := json.Marshal(surfaceCodeData)
			if err != nil {
				liblog.Error("failed to marshal surface code event data for tooth %s: %v", toothID, err)
				continue
			}

			surfaceCodeEvent := &model.HstOdontogram{
				InstitutionID:       visitInfo.IDMstInstitution,
				PatientID:           visitInfo.IDMstPatient,
				VisitID:             visitInfo.ID,
				JourneyPointShortID: journeyPointShortID,
				EventType:           constant.EventTypeToothSurfaceCodeSet,
				ToothID:             toothID,
				EventData:           surfaceCodeJSON,
				CreatedByStaffID:    staffID,
				UnixTimestamp:       unixTimestamp,
				CreatedBy:           staffEmail,
				CreateTime:          currentTime,
			}
			events = append(events, surfaceCodeEvent)

			// Create separate event for surface notes if present
			if surface.Notes != "" {
				surfaceNoteData := model.OdontogramEventData{
					Surface:      surface.Surface,
					SurfaceNotes: surface.Notes,
				}
				surfaceNoteJSON, err := json.Marshal(surfaceNoteData)
				if err != nil {
					liblog.Error("failed to marshal surface note event data for tooth %s: %v", toothID, err)
					continue
				}

				surfaceNoteEvent := &model.HstOdontogram{
					InstitutionID:       visitInfo.IDMstInstitution,
					PatientID:           visitInfo.IDMstPatient,
					VisitID:             visitInfo.ID,
					JourneyPointShortID: journeyPointShortID,
					EventType:           constant.EventTypeToothSurfaceNoteUpdate,
					ToothID:             toothID,
					EventData:           surfaceNoteJSON,
					CreatedByStaffID:    staffID,
					UnixTimestamp:       unixTimestamp,
					CreatedBy:           staffEmail,
					CreateTime:          currentTime,
				}
				events = append(events, surfaceNoteEvent)
			}
		}
	}

	return events
}

// generateOdontogramFromNotes processes detail visits with odontogram blocks and generates events
func generateOdontogramFromNotes(
	patientDB *patientrepo.Conn,
	odontogramDB *odontogramrepo.OdontogramDB,
	staffRepo *staffrepo.Conn,
	journeyDB *journeyrepo.JourneyDB,
) {
	// Get the underlying DB connection for transaction management
	db := patientDB.DB
	const batchSize = 100
	const systemEmail = "system@medilink.com"
	const systemStaffID = 0

	offset := 0
	totalProcessed := 0
	totalSkipped := 0
	totalEvents := 0
	totalErrors := 0

	liblog.Info("Starting odontogram event generation from visit notes...")

	for {
		// Query detail visits with odontogram blocks
		dtlVisits, err := patientDB.ListDtlPatientVisitWithOdontogram(context.Background(), batchSize, offset)
		if err != nil {
			log.Fatalf("failed to list detail visits with odontogram: %v", err)
		}

		if len(dtlVisits) == 0 {
			liblog.Info("No more detail visits to process")
			break
		}

		liblog.Info("Processing batch of %d detail visits (offset: %d)", len(dtlVisits), offset)

		for _, dtlVisit := range dtlVisits {
			// Process each detail visit in its own transaction
			err := processDetailVisitWithTransaction(
				db,
				patientDB,
				odontogramDB,
				staffRepo,
				journeyDB,
				dtlVisit,
				systemEmail,
				systemStaffID,
				&totalProcessed,
				&totalSkipped,
				&totalEvents,
				&totalErrors,
			)
			if err != nil {
				// Error already logged in processDetailVisitWithTransaction
				continue
			}
		}

		offset += batchSize
	}

	liblog.Info("Odontogram event generation complete!")
	liblog.Info("Total detail visits processed: %d", totalProcessed)
	liblog.Info("Total detail visits skipped (already have events): %d", totalSkipped)
	liblog.Info("Total events created: %d", totalEvents)
	liblog.Info("Total errors: %d", totalErrors)
}

// processDetailVisitWithTransaction processes a single detail visit within a transaction
func processDetailVisitWithTransaction(
	db *xormlib.DBConnect,
	patientDB *patientrepo.Conn,
	odontogramDB *odontogramrepo.OdontogramDB,
	staffRepo *staffrepo.Conn,
	journeyDB *journeyrepo.JourneyDB,
	dtlVisit model.DtlPatientVisit,
	systemEmail string,
	systemStaffID int,
	totalProcessed *int,
	totalSkipped *int,
	totalEvents *int,
	totalErrors *int,
) error {
	// Start transaction
	session := db.MasterDB.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		liblog.Error("failed to begin transaction for detail visit %d: %v", dtlVisit.ID, err)
		*totalErrors++
		return err
	}

	// Create context with session
	ctx := xorm.SetDBSession(context.Background(), session)

	// Get visit info to get patient_id, institution_id
	visitInfo, err := patientDB.GetPatientVisitsByID(ctx, dtlVisit.IDTrxPatientVisit)
	if err != nil {
		liblog.Error("failed to get visit info for detail visit %d: %v", dtlVisit.ID, err)
		session.Rollback()
		*totalErrors++
		return err
	}

	journeyPoint, err := journeyDB.GetJourneyPoint(ctx, model.MstJourneyPoint{
		ID:               dtlVisit.IDMstJourneyPoint,
		IDMstInstitution: visitInfo.IDMstInstitution,
	})
	if err != nil {
		liblog.Error("failed to get journey point for detail visit %d: %v", dtlVisit.ID, err)
		session.Rollback()
		*totalErrors++
		return err
	}

	journeyPointShortID := journeyPoint.ShortID

	// Check if events already exist for this visit
	hasEvents, err := odontogramDB.HasEventsForVisit(ctx, visitInfo.IDMstInstitution, visitInfo.ID)
	if err != nil {
		liblog.Error("failed to check existing events for visit %d: %v", visitInfo.ID, err)
		session.Rollback()
		*totalErrors++
		return err
	}

	if hasEvents {
		liblog.Info("Visit %d already has odontogram events, skipping", visitInfo.ID)
		session.Rollback()
		*totalSkipped++
		return nil
	}

	contributors := []string{}
	err = json.Unmarshal(dtlVisit.Contributors, &contributors)
	if err != nil {
		liblog.Error("failed to unmarshal contributors for detail visit %d: %v", dtlVisit.ID, err)
		session.Rollback()
		*totalErrors++
		return err
	}

	// Get staff email
	staffEmail := systemEmail
	staffID := int64(systemStaffID)
	if len(contributors) > 0 {
		staffDetail, err := staffRepo.GetUserDetailByEmail(ctx, contributors[len(contributors)-1])
		if err != nil {
			liblog.Errorf("failed to get staff detail for staff ID %d, using system email: %v", dtlVisit.ActionBy, err)
		} else {
			staffEmail = staffDetail.Staff.Email
			staffID = staffDetail.Staff.ID
		}
	}

	// Parse notes to extract odontogram blocks
	odontogramBlocks, editorTime, err := parseOdontogramBlocks(dtlVisit.Notes)
	if err != nil {
		liblog.Error("failed to parse odontogram blocks for detail visit %d: %v", dtlVisit.ID, err)
		session.Rollback()
		*totalErrors++
		return err
	}

	if len(odontogramBlocks) == 0 {
		liblog.Infof("No odontogram blocks found in detail visit %d", dtlVisit.ID)
		session.Rollback()
		return nil
	}

	// Get journey point short ID

	// Convert odontogram blocks to events
	var allEvents []*model.HstOdontogram
	for _, odontogramBlock := range odontogramBlocks {
		// Use editor time if available, otherwise use create_time
		unixTimestamp := editorTime
		if unixTimestamp == 0 {
			unixTimestamp = dtlVisit.CreateTime.Unix()
		}

		events := convertToothDataToEvents(
			odontogramBlock,
			visitInfo,
			staffEmail,
			staffID,
			journeyPointShortID,
			unixTimestamp,
		)
		allEvents = append(allEvents, events...)
	}

	if len(allEvents) == 0 {
		liblog.Infof("No events generated for detail visit %d", dtlVisit.ID)
		session.Rollback()
		return nil
	}

	// Insert events in batch within the transaction
	err = odontogramDB.InsertEventsBatch(ctx, allEvents)
	if err != nil {
		liblog.Error("failed to insert events for detail visit %d: %v", dtlVisit.ID, err)
		session.Rollback()
		*totalErrors++
		return err
	}

	// Commit transaction
	if err := session.Commit(); err != nil {
		liblog.Error("failed to commit transaction for detail visit %d: %v", dtlVisit.ID, err)
		*totalErrors++
		return err
	}

	liblog.Info("Successfully inserted %d events for detail visit %d (visit %d)", len(allEvents), dtlVisit.ID, visitInfo.ID)
	*totalProcessed++
	*totalEvents += len(allEvents)

	return nil
}
