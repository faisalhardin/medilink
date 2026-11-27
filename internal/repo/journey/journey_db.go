package journey

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"xorm.io/builder"
)

const (
	WrapMsgPrefix = "JourneyDB."

	WrapMsgInsertNewJourneyBoard         = "InsertNewJourneyBoard"
	WrapMsgListJourneyBoard              = "ListJourneyBoard"
	WrapMsgGetJourneyBoardByID           = "GetJourneyBoardByID"
	WrapMsgGetJourneyBoardDetail         = "GetJourneyBoardDetail"
	WrapMsgUpdateJourneyBoard            = "UpdateJourneyBoard"
	WrapMsgDeleteJourneyBoard            = "DeleteJourneyBoard"
	WrapMsgGetJourneyBoardByJourneyPoint = "GetJourneyBoardByJourneyPoint"

	WrapMsgInserNewJourneyPoint = "InsertNewJourneyPoint"
	WrapMsgListJourneyPoints    = "ListJourneyPoints"
	WrapMsgGetJourneyPoint      = "GetJourneyPoints"
	WrapMsgUpdateJourneyPoint   = "UpdateJourneyPoint"
	WrapMsgDeleteJourneyPoint   = "DeleteJourneyPoint"

	WrapMsgInserNewServicePoint = "InserNewServicePoint"
	WrapMsgListServicePoints    = "ListServicePoints"
	WrapMsgListJourneyPointIDs  = "ListJourneyPointIDs"
	WrapMsgListLinkNodes        = "ListLinkNodes"
	WrapMsgLockJourneyBoard     = "LockJourneyBoard"
	WrapMsgLockRoom             = "LockRoom"
	WrapMsgLockRoomShare        = "LockRoomShare"
	WrapMsgDetachRoomLinks      = "DetachRoomLinks"
	WrapMsgReplaceLinks         = "ReplaceLinks"
	WrapMsgListColumnRooms      = "ListColumnRooms"
	WrapMsgGetServicePoint      = "GetServicePoint"
	WrapMsgUpdateServicePoint   = "UpdateServicePoint"
	WrapMsgDeleteServicePoint   = "DeleteServicePoint"

	WrapMsgGetServicePointMappedByJourneyPoints = WrapMsgPrefix + "GetServicePointMappedByJourneyPoints"
	WrapMsgInsertNewMapStaffJourneyPoint        = WrapMsgPrefix + "InsertNewMapStaffJourneyPoint"
)

type JourneyDB struct {
	DB *xormlib.DBConnect
}

func NewJourneyDB(conn *JourneyDB) *JourneyDB {
	return conn
}

// session returns the active transaction, or a master session when the caller
// is not inside one. Link reads that follow a board lock must use this.
func (c *JourneyDB) session(ctx context.Context) *xorm.Session {
	if s := xormlib.GetDBSession(ctx); s != nil {
		return s
	}
	return c.DB.MasterDB.Context(ctx)
}

func (c *JourneyDB) InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	session := c.DB.MasterDB.Table(database.MstJourneyBoardTable)
	_, err = session.InsertOne(journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyBoard)
		return
	}

	return
}

// for customer use, wrap this with institution validation
func (c *JourneyDB) ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyBoardTable)

	if len(params.ID) > 0 {
		session.Where("id = ANY(?)", pq.Array(params.ID))
	}
	if len(params.Name) > 0 {
		nameParams := []string{}
		for _, name := range params.Name {
			nameParams = append(nameParams, fmt.Sprintf("%%%s%%", name))
		}
		session.Where("name = ANY(?)", pq.Array(nameParams))
	}

	if params.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", params.IDMstInstitution)
	}

	err = session.Find(&journeyBoards)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyBoard)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyBoardByParams(ctx context.Context, params model.MstJourneyBoard) (resp model.MstJourneyBoard, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyBoardTable)

	if params.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", params.IDMstInstitution)
	}

	found, err := session.
		ID(params.ID).
		Get(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardByID)
		return
	}

	if !found {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgGetJourneyBoardByID)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyBoardByJourneyPoint(ctx context.Context, journeyPoint model.MstJourneyPoint) (resp model.MstJourneyBoard, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyBoardTable)

	if journeyPoint.ID > 0 {
		session.Where("mmjp.id = ?", journeyPoint.ID)
	} else if journeyPoint.ShortID != "" {
		session.Where("mmjp.short_id = ?", journeyPoint.ShortID)
	} else {
		err = errors.New("missing journey point id or short id")
		return
	}
	found, err := session.Alias("mmjb").
		Join(database.SQLInner, "mdl_mst_journey_point mmjp", "mmjb.id = mmjp.id_mst_journey_board and mmjp.delete_time is null and mmjb.delete_time is null").
		Select("mmjb.*").
		Get(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardByJourneyPoint)
		return
	}

	if !found {
		err = errors.Wrap(constant.ErrorRowNotFound, WrapMsgGetJourneyBoardByJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (resp model.JourneyBoardJoinJourneyPoint, found bool, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyBoardTable)

	if len(params.ID) == 0 {
		err = errors.New("missing board id")
		return
	}
	boardID := params.ID[0]

	if params.IDMstInstitution > 0 {
		session.Where("mmjb.id_mst_institution = ?", params.IDMstInstitution)
	}

	session.
		Alias("mmjb").
		Join(database.SQLLeft, "mdl_mst_journey_point mmjp", `mmjp.id_mst_journey_board = mmjb.id and mmjb.delete_time is null and mmjp.delete_time is null`)
	found, err = session.
		Select("mmjb.*, json_agg(json_build_object('id', mmjp.short_id, 'name', mmjp.name, 'position', mmjp.position, 'id_mst_journey_board', mmjp.id_mst_journey_board, 'create_time', mmjp.create_time, 'update_time', mmjp.update_time) ORDER BY mmjp.position ASC) as mst_journey_point").
		GroupBy("mmjb.id").
		Where("mmjb.id = ?", boardID).
		Where("mmjp.delete_time is null").
		Get(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardDetail)
		return
	}

	return
}

func (c *JourneyDB) UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	session := c.DB.MasterDB.Table(database.MstJourneyBoardTable)

	if journeyBoard.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", journeyBoard.IDMstInstitution)
	}
	count, err := session.
		Where("id = ?", journeyBoard.ID).
		Update(journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyBoard)
		return
	}
	if count == 0 {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgUpdateJourneyBoard)
		return
	}

	return
}

func (c *JourneyDB) DeleteJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	session := c.DB.MasterDB.Table(database.MstJourneyBoardTable)
	count, err := session.
		Delete(journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgDeleteJourneyBoard)
		return
	}
	if count == 0 {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgDeleteJourneyBoard)
		return
	}

	return
}

func (c *JourneyDB) InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.InsertMstJourneyPoint) (err error) {
	// Generate short ID before insertion
	journeyPoint.MstJourneyPoint.BeforeInsert()

	session := c.DB.MasterDB.Table(database.MstJourneyPointTable)
	sqlResult, err := session.SQL(
		`WITH latest_position AS (
			SELECT COALESCE(
				(SELECT j.position + 100
				FROM mdl_mst_journey_point j
				WHERE j.id_mst_journey_board = ? AND j.delete_time IS NULL
				ORDER BY j.position DESC LIMIT 1),
				100
			) AS next_position
		)
		INSERT INTO mdl_mst_journey_point (name, short_id, id_mst_journey_board, id_mst_institution, position, create_time, update_time)
		SELECT ?, ?, ?, ?, lp.next_position, NOW(), NOW()
		FROM latest_position lp
		RETURNING id, name, short_id, id_mst_journey_board, position
		`,
		journeyPoint.MstJourneyPoint.IDMstJourneyBoard,
		journeyPoint.MstJourneyPoint.Name,
		journeyPoint.MstJourneyPoint.ShortID,
		journeyPoint.MstJourneyPoint.IDMstJourneyBoard,
		journeyPoint.MstJourneyPoint.IDMstInstitution,
	).QueryInterface()
	if err != nil {
		err = errors.Wrap(err, WrapMsgInserNewJourneyPoint)
		return
	}

	for _, column := range sqlResult {
		columnID := column["id"].(int64)
		position := int(column["position"].(int64))
		shortID := column["short_id"].(string)

		journeyPoint.MstJourneyPoint.ID = columnID
		journeyPoint.MstJourneyPoint.Position = position
		journeyPoint.MstJourneyPoint.ShortID = shortID
	}

	return
}

func (c *JourneyDB) ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.ListJourneyPointResponse, count int64, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyPointTable).Alias("mjp")

	if params.ID > 0 {
		session.Where("id = ?", params.ID)
	} else if len(params.IDs) > 0 {
		session.Where("id = ANY(?)", pq.Array(params.IDs))
	}
	if len(params.Name) > 0 {
		substringNames := []string{}
		for _, name := range params.Name {
			substringNames = append(substringNames, fmt.Sprintf("%%%s%%", name))
		}
		session.Where("name = ANY(?)", pq.Array(substringNames))
	}
	if params.StaffID > 0 {
		session.
			Join(database.SQLLeft, database.MapStaffJourneyPoint+" msjp", "msjp.id_mst_journey_point = mjp.id")
	}

	// for select logic statement
	if params.StaffID > 0 {
		session.Select("mjp.*, msjp.id_mst_staff is not null as is_owned")
	} else {
		session.Select("mjp.*, false as is_owned")
	}

	count, err = session.
		Where("id_mst_journey_board = ?", params.IDMstBoard).
		FindAndCount(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyPoints)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyPoint(ctx context.Context, param model.MstJourneyPoint) (resp *model.MstJourneyPoint, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyPointTable)
	resp = &param

	_, err = session.
		Get(resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) ListJourneyPointsWithoutShortID(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyPointTable)

	if params.Limit > 0 {
		session.Limit(params.Limit, params.Offset)
	}

	err = session.
		Where("short_id = ? or short_id is null", "").
		Unscoped().
		Find(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyPoints)
		return
	}

	return
}

// GetJourneyPointByShortID retrieves a journey point by its short ID
func (c *JourneyDB) GetJourneyPointByShortID(ctx context.Context, shortID string) (resp *model.MstJourneyPoint, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyPointTable)
	resp = &model.MstJourneyPoint{}

	_, err = session.
		Where("short_id = ?", shortID).
		Get(resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {
	session := c.DB.MasterDB.Table(database.MstJourneyPointTable)

	if journeyPoint.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", journeyPoint.IDMstInstitution).Omit("id_mst_institution")
	}

	if journeyPoint.ID > 0 {
		session.Where("id = ?", journeyPoint.ID)
	} else if journeyPoint.ShortID != "" {
		session.Where("short_id = ?", journeyPoint.ShortID).Omit("short_id")
	} else {
		err = errors.New("missing short id or id")
		return
	}

	_, err = session.
		Unscoped().
		Update(journeyPoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) DeleteJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {
	session := c.DB.MasterDB.Table(database.MstJourneyPointTable)

	count, err := session.
		Delete(journeyPoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		return
	}

	if count == 0 {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgUpdateJourneyPoint)
		return
	}
	return
}

func (c *JourneyDB) InserNewServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {
	session := c.DB.MasterDB.Table(database.MstServicePointTable)

	_, err = session.InsertOne(mstServicePoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInserNewServicePoint)
		return
	}

	return

}

// ListServicePoints returns no rows without a clinic, so a session with no
// clinic can never list another clinic's rooms.
func (c *JourneyDB) ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error) {
	if params.IDMstInstitution <= 0 {
		return []model.MstServicePoint{}, nil
	}

	session := c.DB.SlaveDB.Table(database.MstServicePointTable)
	if params.IncludeArchived {
		session.Unscoped()
	}

	err = session.
		Where(servicePointListCond(params)).
		Limit(params.Limit, params.Offset).
		OrderBy("id ASC").
		Find(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListServicePoints)
		return
	}
	return
}

func servicePointListCond(params model.GetServicePointParams) builder.Cond {
	cond := builder.And(builder.Eq{"id_mst_institution": params.IDMstInstitution})
	if len(params.ID) > 0 {
		cond = cond.And(builder.Expr("id = ANY(?)", pq.Array(params.ID)))
	}
	if len(params.Name) > 0 {
		substringNames := []string{}
		for _, name := range params.Name {
			substringNames = append(substringNames, fmt.Sprintf("%%%s%%", name))
		}
		cond = cond.And(builder.Expr("name ilike ANY(?)", pq.Array(substringNames)))
	}
	if params.IDMstBoard > 0 {
		cond = cond.And(builder.Eq{"id_mst_journey_board": params.IDMstBoard})
	}
	return cond
}

func (c *JourneyDB) ListJourneyPointIDs(ctx context.Context, roomIDs []int64) (resp map[int64][]int64, err error) {
	resp = map[int64][]int64{}
	if len(roomIDs) == 0 {
		return
	}

	rows := []model.MapServicePointJourneyPoint{}
	err = c.DB.SlaveDB.Table(database.MapServicePointJourneyPoint).Alias("m").
		Join(database.SQLInner, database.MstJourneyPointTable+" jp", "jp.id = m.id_mst_journey_point AND jp.delete_time IS NULL").
		Join(database.SQLInner, database.MstServicePointTable+" sp", "sp.id = m.id_mst_service_point AND jp.id_mst_journey_board = sp.id_mst_journey_board").
		Select("m.id_mst_service_point, m.id_mst_journey_point").
		Where("m.id_mst_service_point = ANY(?)", pq.Array(roomIDs)).
		OrderBy("m.id_mst_journey_point ASC").
		Find(&rows)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyPointIDs)
		return
	}

	for _, row := range rows {
		resp[row.IDMstServicePoint] = append(resp[row.IDMstServicePoint], row.IDMstJourneyPoint)
	}
	return
}

func (c *JourneyDB) ListLinkNodes(ctx context.Context, query model.LinkQuery) (nodes []model.LinkNode, err error) {
	nodes = []model.LinkNode{}
	if query.Clinic <= 0 || len(query.IDs) == 0 {
		return
	}

	var sqlText string
	switch query.Kind {
	case model.LinkRoom:
		sqlText = `SELECT id, id_mst_journey_board AS board_id
			FROM ` + database.MstServicePointTable + `
			WHERE id_mst_institution = ?
			  AND delete_time IS NULL
			  AND id = ANY(?)`
	case model.LinkColumn:
		sqlText = `SELECT jp.id, jp.id_mst_journey_board AS board_id
			FROM ` + database.MstJourneyPointTable + ` jp
			INNER JOIN ` + database.MstJourneyBoardTable + ` b
			  ON b.id = jp.id_mst_journey_board
			 AND b.delete_time IS NULL
			 AND b.id_mst_institution = ?
			WHERE jp.delete_time IS NULL
			  AND jp.id = ANY(?)`
	default:
		return
	}

	err = c.session(ctx).SQL(sqlText, query.Clinic, pq.Array(query.IDs)).Find(&nodes)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListLinkNodes)
		return
	}
	if nodes == nil {
		nodes = []model.LinkNode{}
	}
	return
}

func (c *JourneyDB) LockJourneyBoard(ctx context.Context, boardID int64) (err error) {
	var id int64
	has, err := c.session(ctx).SQL(
		`SELECT id FROM `+database.MstJourneyBoardTable+` WHERE id = ? FOR UPDATE`,
		boardID,
	).Get(&id)
	if err != nil {
		return errors.Wrap(err, WrapMsgLockJourneyBoard)
	}
	if !has {
		return errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgLockJourneyBoard)
	}
	return nil
}

func (c *JourneyDB) lockRoom(ctx context.Context, clinicID, roomID int64, mode, wrap string) (*model.MstServicePoint, error) {
	if clinicID <= 0 {
		return nil, errors.Wrap(constant.ErrorNoAffectedRow, wrap)
	}
	row := &model.MstServicePoint{}
	has, err := c.session(ctx).SQL(
		`SELECT id, name, id_mst_journey_board, id_mst_institution, create_time, update_time, delete_time
		 FROM `+database.MstServicePointTable+`
		 WHERE id = ? AND id_mst_institution = ? AND delete_time IS NULL
		 `+mode,
		roomID, clinicID,
	).Get(row)
	if err != nil {
		return nil, errors.Wrap(err, wrap)
	}
	if !has {
		return nil, errors.Wrap(constant.ErrorNoAffectedRow, wrap)
	}
	return row, nil
}

// LockRoom takes FOR UPDATE on the active clinic-scoped room row. BE-5 archive holds this lock for the whole txn.
func (c *JourneyDB) LockRoom(ctx context.Context, clinicID, roomID int64) (*model.MstServicePoint, error) {
	return c.lockRoom(ctx, clinicID, roomID, "FOR UPDATE", WrapMsgLockRoom)
}

// LockRoomShare takes FOR SHARE on the same row. BE-7 move/first-step uses this so it waits on archive and cannot land on an archived room.
func (c *JourneyDB) LockRoomShare(ctx context.Context, clinicID, roomID int64) (*model.MstServicePoint, error) {
	return c.lockRoom(ctx, clinicID, roomID, "FOR SHARE", WrapMsgLockRoomShare)
}

func (c *JourneyDB) DetachRoomLinks(ctx context.Context, clinicID, roomID int64) (err error) {
	if clinicID <= 0 {
		return errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgDetachRoomLinks)
	}
	_, err = c.session(ctx).Exec(`
		DELETE FROM `+database.MapServicePointJourneyPoint+` AS m
		WHERE m.id_mst_service_point = ?
		  AND EXISTS (
		    SELECT 1 FROM `+database.MstServicePointTable+` sp
		    WHERE sp.id = m.id_mst_service_point
		      AND sp.id_mst_institution = ?
		  )`, roomID, clinicID)
	if err != nil {
		return errors.Wrap(err, WrapMsgDetachRoomLinks)
	}
	return nil
}

func (c *JourneyDB) ReplaceLinks(ctx context.Context, set model.LinkSet) (err error) {
	targets := set.Targets
	if targets == nil {
		targets = []int64{}
	}

	var deleteSQL, insertSQL string
	switch set.Side {
	case model.LinkRoom:
		deleteSQL = `DELETE FROM ` + database.MapServicePointJourneyPoint + ` AS m
			WHERE m.id_mst_service_point = ?
			  AND EXISTS (
			    SELECT 1 FROM ` + database.MstServicePointTable + ` sp
			    WHERE sp.id = m.id_mst_service_point
			      AND sp.id_mst_institution = ?
			  )
			  AND m.id_mst_journey_point IN (
			    SELECT jp.id FROM ` + database.MstJourneyPointTable + ` jp
			    WHERE jp.id_mst_journey_board = ?
			  )`
		insertSQL = `INSERT INTO ` + database.MapServicePointJourneyPoint + ` (id_mst_service_point, id_mst_journey_point)
			SELECT sp.id, jp.id
			FROM unnest(CAST(? AS bigint[])) AS t(id)
			JOIN ` + database.MstServicePointTable + ` sp
			  ON sp.id = ?
			 AND sp.delete_time IS NULL
			 AND sp.id_mst_institution = ?
			 AND sp.id_mst_journey_board = ?
			JOIN ` + database.MstJourneyPointTable + ` jp
			  ON jp.id = t.id
			 AND jp.delete_time IS NULL
			 AND jp.id_mst_journey_board = sp.id_mst_journey_board
			ON CONFLICT (id_mst_service_point, id_mst_journey_point) DO NOTHING`
		if _, err = c.session(ctx).Exec(deleteSQL, set.Subject, set.Clinic, set.Board); err != nil {
			return errors.Wrap(err, WrapMsgReplaceLinks)
		}
		_, err = c.session(ctx).Exec(insertSQL, pq.Array(targets), set.Subject, set.Clinic, set.Board)
	case model.LinkColumn:
		deleteSQL = `DELETE FROM ` + database.MapServicePointJourneyPoint + ` AS m
			WHERE m.id_mst_journey_point = ?
			  AND m.id_mst_service_point IN (
			    SELECT sp.id FROM ` + database.MstServicePointTable + ` sp
			    WHERE sp.id_mst_journey_board = ?
			      AND sp.id_mst_institution = ?
			  )`
		insertSQL = `INSERT INTO ` + database.MapServicePointJourneyPoint + ` (id_mst_service_point, id_mst_journey_point)
			SELECT sp.id, jp.id
			FROM unnest(CAST(? AS bigint[])) AS t(id)
			JOIN ` + database.MstJourneyPointTable + ` jp
			  ON jp.id = ?
			 AND jp.delete_time IS NULL
			 AND jp.id_mst_journey_board = ?
			JOIN ` + database.MstServicePointTable + ` sp
			  ON sp.id = t.id
			 AND sp.delete_time IS NULL
			 AND sp.id_mst_institution = ?
			 AND sp.id_mst_journey_board = jp.id_mst_journey_board
			ON CONFLICT (id_mst_service_point, id_mst_journey_point) DO NOTHING`
		if _, err = c.session(ctx).Exec(deleteSQL, set.Subject, set.Board, set.Clinic); err != nil {
			return errors.Wrap(err, WrapMsgReplaceLinks)
		}
		_, err = c.session(ctx).Exec(insertSQL, pq.Array(targets), set.Subject, set.Board, set.Clinic)
	default:
		return errors.Wrap(errors.New("unknown link side"), WrapMsgReplaceLinks)
	}
	if err != nil {
		return errors.Wrap(err, WrapMsgReplaceLinks)
	}
	return nil
}

type columnRoomRow struct {
	ID   int64  `xorm:"id"`
	Name string `xorm:"name"`
}

func (c *JourneyDB) ListColumnRooms(ctx context.Context, clinic, columnID int64) (rooms []model.PickerRoom, err error) {
	rooms = []model.PickerRoom{}
	if clinic <= 0 {
		return
	}

	rows := []columnRoomRow{}
	err = c.session(ctx).SQL(`
		SELECT sp.id, sp.name
		FROM `+database.MapServicePointJourneyPoint+` m
		INNER JOIN `+database.MstServicePointTable+` sp
		  ON sp.id = m.id_mst_service_point
		 AND sp.delete_time IS NULL
		 AND sp.id_mst_institution = ?
		INNER JOIN `+database.MstJourneyPointTable+` jp
		  ON jp.id = m.id_mst_journey_point
		 AND jp.delete_time IS NULL
		 AND jp.id_mst_journey_board = sp.id_mst_journey_board
		INNER JOIN `+database.MstJourneyBoardTable+` b
		  ON b.id = jp.id_mst_journey_board
		 AND b.delete_time IS NULL
		 AND b.id_mst_institution = ?
		WHERE m.id_mst_journey_point = ?
		ORDER BY sp.id ASC`,
		clinic, clinic, columnID).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgListColumnRooms)
	}

	rooms = make([]model.PickerRoom, 0, len(rows))
	for _, row := range rows {
		rooms = append(rooms, model.PickerRoom{ID: row.ID, Name: row.Name})
	}
	return
}

func (c *JourneyDB) GetServicePoint(ctx context.Context, servicePoint model.MstServicePoint) (resp *model.MstServicePoint, err error) {
	if servicePoint.IDMstInstitution <= 0 {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgGetServicePoint)
		return
	}

	session := c.DB.SlaveDB.Table(database.MstServicePointTable)
	resp = &servicePoint
	found, err := session.
		Where("id = ? AND id_mst_institution = ?", resp.ID, resp.IDMstInstitution).
		Get(resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetServicePoint)
		return
	}

	if !found {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgGetServicePoint)
		return
	}

	return
}

func (c *JourneyDB) UpdateServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {
	session := c.DB.MasterDB.Table(database.MstServicePointTable)

	if mstServicePoint.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", mstServicePoint.IDMstInstitution)
	}

	count, err := session.
		Where("id = ?", mstServicePoint.ID).
		Update(mstServicePoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateServicePoint)
		return
	}

	if count == 0 {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgUpdateServicePoint)
		return
	}

	return
}

func (c *JourneyDB) DeleteServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {
	session := c.session(ctx).Table(database.MstServicePointTable)

	if mstServicePoint.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", mstServicePoint.IDMstInstitution)
	}

	count, err := session.
		Where("id = ?", mstServicePoint.ID).
		Delete(mstServicePoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgDeleteServicePoint)
		return
	}

	if count == 0 {
		err = errors.Wrap(constant.ErrorNoAffectedRow, WrapMsgDeleteServicePoint)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyPointMappedByStaff(ctx context.Context, mstStaff model.MstStaff) (journeyPoint []model.MstJourneyPoint, err error) {

	session := c.DB.SlaveDB.Table(database.MstJourneyPointTable).Alias("mjp")

	session.Join(database.SQLInner, database.MapStaffJourneyPoint+" msjp", "msjp.id_mst_journey_point = mjp.id")
	if mstStaff.ID > 0 {
		session.Where("msjp.id_mst_staff = ?", mstStaff.ID)
	} else if len(mstStaff.Email) > 0 {
		session.
			Join(database.SQLInner, database.MstStaff+" mms", "mms.id = msjp.id_mst_staff").
			Where("ms.email = ?", mstStaff.Email)
	}
	err = session.
		Select("mjp.short_id, mjp.name, mjp.id_mst_journey_board").
		Find(&journeyPoint)
	if err != nil {
		err = errors.Wrap(err, "conn.GetJourneyPointMappedByStaff")
		return
	}

	return
}

func (c *JourneyDB) InsertNewMapStaffJourneyPoint(ctx context.Context, mapStaffJourneyPoint *model.MapStaffJourneyPoint) (err error) {
	session := c.DB.MasterDB.Table(database.MapStaffJourneyPoint)
	_, err = session.
		NoAutoTime().
		InsertOne(mapStaffJourneyPoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewMapStaffJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) GetServicePointMappedByJourneyPoints(ctx context.Context, journeyPoints []model.MstJourneyPoint, mstStaff model.MstStaff) (servicePoints []model.MstServicePoint, err error) {
	session := c.DB.SlaveDB.Table(database.MstServicePointTable).Alias("msp")

	session.
		Join(database.SQLInner,
			database.MapStaffServicePoint+" mssp",
			"mssp.id_mst_service_point = msp.id")
	if mstStaff.ID > 0 {
		session.Where("mssp.id_mst_staff = ?", mstStaff.ID)
	} else if len(mstStaff.Email) > 0 {
		session.
			Join(database.SQLInner, database.MstStaff+" mms", "mms.id = msjp.id_mst_staff").
			Where("ms.email = ?", mstStaff.Email)
	}

	err = session.
		Select("msp.*").
		Find(&servicePoints)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetServicePointMappedByJourneyPoints)
		return
	}

	return
}
