package journey

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/constant/database"
	"github.com/faisalhardin/medilink/internal/entity/model"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/lib/pq"
	"github.com/pkg/errors"
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
	WrapMsgGetServicePoint      = "GetServicePoint"
	WrapMsgUpdateServicePoint   = "UpdateServicePoint"
	WrapMsgDeleteServicePoint   = "DeleteServicePoint"

	WrapMsgGetServicePointMappedByJourneyPoints = WrapMsgPrefix + "GetServicePointMappedByJourneyPoints"
)

type JourneyDB struct {
	DB *xormlib.DBConnect
}

func NewJourneyDB(conn *JourneyDB) *JourneyDB {
	return conn
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

func (c *JourneyDB) GetJourneyBoardByID(ctx context.Context, id int64) (resp model.MstJourneyBoard, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyBoardTable)

	found, err := session.
		ID(id).
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
	// journeyPoint.MstJourneyPoint.BeforeInsert()

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
		journeyPoint.MstJourneyPoint.IDMstInstitution,
		journeyPoint.MstJourneyPoint.IDMstJourneyBoard,
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

func (c *JourneyDB) ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, count int64, err error) {
	session := c.DB.SlaveDB.Table(database.MstJourneyPointTable)

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

	if journeyPoint.ShortID != "" {
		session.Where("short_id = ?", journeyPoint.ShortID).Omit("short_id")
	} else if journeyPoint.ID > 0 {
		session.Where("id = ?", journeyPoint.ID)
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

func (c *JourneyDB) ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error) {
	session := c.DB.SlaveDB.Table(database.MstServicePointTable)

	if params.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", params.IDMstInstitution)
	}

	if len(params.ID) > 0 {
		session.Where("id = ANY(?)", pq.Array(params))
	}

	if len(params.Name) > 0 {
		substringNames := []string{}
		for _, name := range params.Name {
			substringNames = append(substringNames, fmt.Sprintf("%%%s%%", name))
		}
		session.Where("name ilike ANY(?)", pq.Array(substringNames))
	}

	err = session.
		Where("id_mst_journey_board = ?", params.IDMstBoard).
		Limit(params.Limit, params.Offset).
		OrderBy("id ASC").
		Find(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListServicePoints)
		return
	}
	return
}

func (c *JourneyDB) GetServicePoint(ctx context.Context, servicePoint model.MstServicePoint) (resp *model.MstServicePoint, err error) {
	session := c.DB.SlaveDB.Table(database.MstServicePointTable)
	resp = &servicePoint
	found, err := session.
		Where("id = ?", resp.ID).
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
	session := c.DB.MasterDB.Table(database.MstServicePointTable)

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
