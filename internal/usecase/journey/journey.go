package journey

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	patientRepo "github.com/faisalhardin/medilink/internal/entity/repo/patient"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	customtime "github.com/faisalhardin/medilink/pkg/type/time"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
)

const (
	WrapMsgPrefix                = "JourneyUC."
	WrapMsgInsertNewJourneyBoard = WrapMsgPrefix + "InsertNewJourneyBoard"
	WrapMsgListJourneyBoard      = WrapMsgPrefix + "ListJourneyBoard"
	WrapMsgGetJourneyBoardDetail = WrapMsgPrefix + "GetJourneyBoardDetail"
	WrapMsgUpdateJourneyBoard    = WrapMsgPrefix + "UpdateJourneyBoard"
	WrapMsgDeleteJourneyBoard    = WrapMsgPrefix + "DeleteJourneyBoard"

	WrapMsgListJourneyPoints     = WrapMsgPrefix + "ListJourneyPoints "
	WrapMsgInsertNewJourneyPoint = WrapMsgPrefix + "InsertNewJourneyPoint"
	WrapMsgUpdateJourneyPoint    = WrapMsgPrefix + "UpdateJourneyPoint"
	WrapMsgArchiveJourneyPoint   = WrapMsgPrefix + "ArchiveJourneyPoint"

	WrapMsgGetServicePoint       = WrapMsgPrefix + "GetServicePoint"
	WrapMsgListServicePoints     = WrapMsgPrefix + "ListServicePoints"
	WrapMsgInsertNewServicePoint = WrapMsgPrefix + "InsertNewServicePoint"
	WrapMsgUpdateServicePoint    = WrapMsgPrefix + "UpdateServicePoint"
	WrapMsgArchiveServicePoint   = WrapMsgPrefix + "ArchiveServicePoint"
)

type JourneyUC struct {
	JourneyDB    journeyRepo.JourneyDB
	PatientDB    patientRepo.PatientDB
	Transaction  xorm.DBTransactionInterface
	RoomLocation RoomLocationPort
}

func NewJourneyUC(conn *JourneyUC) *JourneyUC {
	return conn
}

func (u *JourneyUC) InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	err = u.JourneyDB.InsertNewJourneyBoard(ctx, journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyBoard)
		return
	}

	return
}

func (u *JourneyUC) ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error) {
	journeyBoards, err = u.JourneyDB.ListJourneyBoard(ctx, params)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyBoard)
		return
	}

	return
}

func (u *JourneyUC) GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoardDetail model.JourneyBoardJoinJourneyPoint, err error) {
	if len(params.ID) == 0 {
		err = errors.New("missing board id")
		return
	}

	journeyBoard, err := u.JourneyDB.GetJourneyBoardByParams(ctx, model.MstJourneyBoard{
		ID: params.ID[0],
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardDetail)
		return
	}

	journeyPoints, _, err := u.JourneyDB.ListJourneyPoints(ctx, model.GetJourneyPointParams{
		IDMstBoard: journeyBoard.ID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardDetail)
		return
	}

	journeyBoardDetail.JourneyPoints = journeyPoints
	journeyBoardDetail.ID = journeyBoard.ID
	journeyBoardDetail.Name = journeyBoard.Name
	journeyBoardDetail.IDMstInstitution = journeyBoard.IDMstInstitution
	journeyBoardDetail.CreateTime = journeyBoard.CreateTime
	journeyBoardDetail.UpdateTime = journeyBoard.UpdateTime

	return
}

func (u *JourneyUC) UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	err = u.JourneyDB.UpdateJourneyBoard(ctx, journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyBoard)
		return
	}

	return
}

func (u *JourneyUC) DeleteJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	err = u.JourneyDB.DeleteJourneyBoard(ctx, journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgDeleteJourneyBoard)
		return
	}
	return
}

func (u *JourneyUC) validateJourneyBoardOwnership(ctx context.Context, userDetail model.UserJWTPayload, boardID int64) (err error) {

	mstJourneyBoard, err := u.JourneyDB.GetJourneyBoardByParams(ctx, model.MstJourneyBoard{
		ID: boardID,
	})
	if err != nil {
		return
	}

	if mstJourneyBoard.IDMstInstitution != userDetail.InstitutionID {
		return commonerr.SetNewUnauthorizedAPICall()
	}

	return nil
}

func (u *JourneyUC) validateJourneyPointOwnership(ctx context.Context, journeyPointShortID string, userDetail model.UserJWTPayload) (err error) {

	journeyBoard, err := u.JourneyDB.GetJourneyBoardByJourneyPoint(ctx, model.MstJourneyPoint{
		ShortID: journeyPointShortID,
	})
	if err != nil && errors.Is(err, constant.ErrorRowNotFound) {
		err = errors.Wrap(commonerr.SetNewUnauthorizedAPICall(), WrapMsgUpdateJourneyPoint)
		return
	} else if err != nil {
		return
	}

	if journeyBoard.IDMstInstitution != userDetail.InstitutionID {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	return nil
}

func (u *JourneyUC) ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (servicePoint []model.ListJourneyPointResponse, err error) {
	servicePoint, _, err = u.JourneyDB.ListJourneyPoints(ctx, params)
	if err != nil && errors.Is(err, constant.ErrorNoAffectedRow) {
		return nil, nil
	}
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyPoints)
		return
	}

	return
}

func (u *JourneyUC) InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (response model.InsertJourneyPointResponse, err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	err = u.validateJourneyBoardOwnership(ctx, userDetail, journeyPoint.IDMstJourneyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyPoint)
		return
	}

	session, _ := u.Transaction.Begin(ctx)
	defer u.Transaction.Finish(session, &err)
	ctx = xorm.SetDBSession(ctx, session)

	insertJourneyPointRequest := &model.InsertMstJourneyPoint{
		MstJourneyPoint: journeyPoint,
	}

	err = u.JourneyDB.InsertNewJourneyPoint(ctx, insertJourneyPointRequest)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyPoint)
		return
	}

	err = u.JourneyDB.InsertNewMapStaffJourneyPoint(ctx, &model.MapStaffJourneyPoint{
		IDMstStaff:        userDetail.UserID,
		IDMstJourneyPoint: insertJourneyPointRequest.MstJourneyPoint.ID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyBoard)
		return
	}

	response.IsOwned = null.BoolFrom(true)
	response.MstJourneyPoint = *insertJourneyPointRequest.MstJourneyPoint

	return response, nil
}

func (u *JourneyUC) UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = errors.Wrap(commonerr.SetNewUnauthorizedAPICall(), WrapMsgUpdateJourneyPoint)
		return
	}

	err = u.validateJourneyPointOwnership(ctx, journeyPoint.ShortID, userDetail)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		return
	}

	err = u.JourneyDB.UpdateJourneyPoint(ctx, journeyPoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		return
	}

	return
}

func (u *JourneyUC) ArchiveJourneyPoint(ctx context.Context, journeyPoint *model.ArchiveJourneyPointRequest) (err error) {

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = errors.Wrap(commonerr.SetNewUnauthorizedAPICall(), WrapMsgUpdateJourneyPoint)
		return
	}

	err = u.validateJourneyPointOwnership(ctx, journeyPoint.ShortID, userDetail)
	if err != nil {
		err = errors.Wrap(err, WrapMsgArchiveJourneyPoint)
		return
	}

	patientVisits, err := u.PatientDB.GetPatientVisits(ctx, model.GetPatientVisitParams{
		ShortIDMstJourneyPoint: journeyPoint.ShortID,
		IDMstInstitution:       userDetail.InstitutionID,
		CommonRequestPayload: model.CommonRequestPayload{
			FromTime: customtime.Time{time.Now().AddDate(0, 0, -3)},
			ToTime:   customtime.Time{time.Now()},
		},
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgArchiveJourneyPoint)
		return
	}

	if len(patientVisits) > 0 {
		err = commonerr.SetNewBadRequest("invalid", "There are visits that are still active within 3 days")
		return
	}

	err = u.JourneyDB.DeleteJourneyPoint(ctx, &model.MstJourneyPoint{
		ShortID: journeyPoint.ShortID,
	})
	if err != nil {
		err = errors.Wrap(err, WrapMsgArchiveJourneyPoint)
		return
	}

	return
}

// roomLookupErr is the only translation from "no visible room row" to the wire.
// UserJourneyDB scopes by the JWT clinic and xorm hides archived rows, so a
// missing room, another clinic, an orphan, and an archived room all arrive as
// ErrorNoAffectedRow. They leave as one Set404 so the caller cannot tell which.
func roomLookupErr(err error) error {
	if errors.Is(err, constant.ErrorNoAffectedRow) {
		return commonerr.Set404()
	}
	return err
}

func (u *JourneyUC) GetServicePoint(ctx context.Context, servicePointID int64) (room *model.Room, err error) {
	row, err := u.JourneyDB.GetServicePoint(ctx, model.MstServicePoint{
		ID: servicePointID,
	})
	if err != nil {
		err = errors.Wrap(roomLookupErr(err), WrapMsgGetServicePoint)
		return
	}

	rooms, err := u.toRooms(ctx, []model.MstServicePoint{*row})
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetServicePoint)
		return
	}
	room = &rooms[0]
	return
}

func (u *JourneyUC) ListServicePoints(ctx context.Context, params model.GetServicePointParams) (rooms []model.Room, err error) {
	rows, err := u.JourneyDB.ListServicePoints(ctx, params)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListServicePoints)
		return
	}

	rooms, err = u.toRooms(ctx, rows)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListServicePoints)
		return
	}
	return
}

func (u *JourneyUC) toRooms(ctx context.Context, rows []model.MstServicePoint) ([]model.Room, error) {
	rooms := make([]model.Room, 0, len(rows))
	if len(rows) == 0 {
		return rooms, nil
	}

	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	jpByRoom, err := u.JourneyDB.ListJourneyPointIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	locByRoom, err := roomLocationPort(u.RoomLocation).Read(ctx, rows[0].IDMstInstitution, ids)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		jp := jpByRoom[row.ID]
		if jp == nil {
			jp = []int64{}
		}
		loc := locByRoom[row.ID]
		if loc != nil {
			if _, ok := model.ParseSyncStatus(string(loc.SyncStatus)); !ok {
				loc = nil
			}
		}
		rooms = append(rooms, model.Room{
			ID:              row.ID,
			Name:            row.Name,
			BoardID:         row.IDMstBoard,
			Archived:        row.DeleteTime != nil,
			JourneyPointIDs: jp,
			Location:        loc,
		})
	}
	return rooms, nil
}

const servicePointNameConstraint = "ux_service_point_institution_name"

func roomName(raw string) (string, error) {
	name := strings.TrimSpace(raw)
	if name == "" {
		return "", commonerr.RoomNameRequired()
	}
	return name, nil
}

func roomWriteErr(err error) error {
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) && pgxErr.Code == "23505" && strings.Contains(pgxErr.ConstraintName, servicePointNameConstraint) {
		return commonerr.RoomNameTaken()
	}
	// Master DB uses lib/pq; unique violations arrive as *pq.Error.
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" && strings.Contains(pqErr.Constraint, servicePointNameConstraint) {
		return commonerr.RoomNameTaken()
	}
	return roomLookupErr(err)
}

func (u *JourneyUC) clinicalGuard(ctx context.Context, clinic, roomID int64, clinical bool) error {
	if !clinical {
		return nil
	}
	locked, err := roomLocationPort(u.RoomLocation).Clearing(ctx, clinic, roomID)
	if err != nil || !locked {
		return err
	}
	return commonerr.RoomClearing()
}

func (u *JourneyUC) clinicalEnqueue(ctx context.Context, clinic, roomID int64, clinical bool) error {
	if !clinical {
		return nil
	}
	return roomLocationPort(u.RoomLocation).Enqueue(ctx, clinic, roomID, jobLocationCreate)
}

func (u *JourneyUC) InsertNewServicePoint(ctx context.Context, req model.CreateRoomRequest) (room *model.Room, err error) {
	ids := commonerr.RoomIDs{Board: req.BoardID}
	defer func() {
		if room != nil {
			ids.Room = room.ID
		}
		commonerr.LogRoom(commonerr.RoomOpCreate, ids, err)
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}
	// JourneyDB omits id_mst_institution unless the id is above 0.
	if userDetail.InstitutionID <= 0 {
		err = commonerr.Set404()
		return
	}
	ids.Institution = userDetail.InstitutionID

	name, err := roomName(req.Name)
	if err != nil {
		return
	}

	err = u.validateJourneyBoardOwnership(ctx, userDetail, req.BoardID)
	if err != nil {
		err = errors.Wrap(servicePointCreateErr(err), WrapMsgInsertNewServicePoint)
		return
	}

	err = u.clinicalGuard(ctx, userDetail.InstitutionID, 0, req.IsClinical)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewServicePoint)
		return
	}

	row := &model.MstServicePoint{
		Name:             name,
		IDMstBoard:       req.BoardID,
		IDMstInstitution: userDetail.InstitutionID,
	}
	err = u.JourneyDB.InserNewServicePoint(ctx, row)
	if err != nil {
		err = errors.Wrap(roomWriteErr(err), WrapMsgInsertNewServicePoint)
		return
	}

	err = u.clinicalEnqueue(ctx, userDetail.InstitutionID, row.ID, req.IsClinical)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewServicePoint)
		return
	}

	rooms, err := u.toRooms(ctx, []model.MstServicePoint{*row})
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewServicePoint)
		return
	}
	room = &rooms[0]
	return
}

func servicePointCreateErr(err error) error {
	err = roomLookupErr(err)
	var msg *commonerr.ErrorMessage
	if errors.As(err, &msg) && msg != nil && msg.Code == http.StatusUnauthorized {
		return commonerr.Set404()
	}
	return err
}

func (u *JourneyUC) UpdateServicePoint(ctx context.Context, id int64, req model.RenameRoomRequest) (room *model.Room, err error) {
	ids := commonerr.RoomIDs{Room: id}
	defer func() {
		commonerr.LogRoom(commonerr.RoomOpRename, ids, err)
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}
	if userDetail.InstitutionID <= 0 {
		err = commonerr.Set404()
		return
	}
	ids.Institution = userDetail.InstitutionID

	var name string
	if req.Name != nil {
		name, err = roomName(*req.Name)
		if err != nil {
			return
		}
	}

	row, err := u.JourneyDB.GetServicePoint(ctx, model.MstServicePoint{ID: id})
	if err != nil {
		err = errors.Wrap(roomLookupErr(err), WrapMsgUpdateServicePoint)
		return
	}
	ids.Board = row.IDMstBoard
	if req.Name == nil {
		name = row.Name
	}

	err = u.clinicalGuard(ctx, userDetail.InstitutionID, id, req.IsClinical)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateServicePoint)
		return
	}

	err = u.JourneyDB.UpdateServicePoint(ctx, &model.MstServicePoint{ID: id, Name: name})
	if err != nil {
		err = errors.Wrap(roomWriteErr(err), WrapMsgUpdateServicePoint)
		return
	}

	err = u.clinicalEnqueue(ctx, userDetail.InstitutionID, id, req.IsClinical)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateServicePoint)
		return
	}

	row.Name = name
	rooms, err := u.toRooms(ctx, []model.MstServicePoint{*row})
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateServicePoint)
		return
	}
	room = &rooms[0]
	return
}

func (u *JourneyUC) ArchiveServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error) {
	err = u.JourneyDB.DeleteServicePoint(ctx, servicePoint)
	if err != nil {
		err = errors.Wrap(roomLookupErr(err), WrapMsgArchiveServicePoint)
		return
	}

	return
}
