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
	MstJourneyBoardTable = "mdl_mst_journey_board"
	MstJourneyPointTable = "mdl_mst_journey_point"
	MstServicePointTable = "mdl_mst_service_point"

	WrapMsgInsertNewJourneyBoard = "InsertNewJourneyBoard"
	WrapMsgListJourneyBoard      = "ListJourneyBoard"
	WrapMsgGetJourneyBoardDetail = "GetJourneyBoardDetail"
	WrapMsgUpdateJourneyBoard    = "UpdateJourneyBoard"
	WrapMsgDeleteJourneyBoard    = "DeleteJourneyBoard"

	WrapMsgInserNewJourneyPoint = "InsertNewJourneyPoint"
	WrapMsgListJourneyPoints    = "ListJourneyPoints"
	WrapMsgGetJourneyPoints     = "GetJourneyPoints"
	WrapMsgUpdateJourneyPoint   = "UpdateJourneyPoint"
	WrapMsgDeleteJourneyPoint   = "DeleteJourneyPoint"

	WrapMsgInserNewServicePoint = "InserNewServicePoint"
	WrapMsgListServicePoints    = "ListServicePoints"
	WrapMsgGetServicePoint      = "GetServicePoint"
	WrapMsgUpdateServicePoint   = "UpdateServicePoint"
	WrapMsgDeleteServicePoint   = "DeleteServicePoint"
)

type JourneyDB struct {
	DB *xormlib.DBConnect
}

func NewJourneyDB(conn *JourneyDB) *JourneyDB {
	return conn
}

func (c *JourneyDB) InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	session := c.DB.MasterDB.Table(MstJourneyBoardTable)
	_, err = session.InsertOne(journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyBoard)
		return
	}

	return
}

// for customer use, wrap this with institution validation
func (c *JourneyDB) ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error) {
	session := c.DB.SlaveDB.Table(MstJourneyBoardTable)

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

	err = session.Find(&journeyBoards)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyBoard)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (resp model.JourneyBoardJoinJourneyPoint, found bool, err error) {
	session := c.DB.SlaveDB.Table(MstJourneyBoardTable)

	if len(params.ID) == 0 {
		err = errors.New("missing board id")
		return
	}
	boardID := params.ID[0]

	if params.IDMstInstitution > 0 {
		session.Where("id_mst_institution = ?", params.IDMstInstitution)
	}

	session.Alias("mmjb").
		Join(database.SQLLeft, "mdl_mst_journey_point mmjp", `mmjp.id_mst_journey_board = mmjb.id and mmjb.delete_time is null and mmjp.delete_time is null`)

	found, err = session.
		Select("mmjb.*, json_agg(json_build_object('id', mmjp.id, 'name', mmjp.name, 'position', mmjp.position, 'id_mst_journey_board', mmjp.id_mst_journey_board, 'create_time', mmjp.create_time, 'update_time', mmjp.update_time)) as mst_journey_point").
		GroupBy("mmjb.id").
		Where("mmjb.id = ?", boardID).
		Get(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardDetail)
		return
	}

	return
}

func (c *JourneyDB) UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	session := c.DB.MasterDB.Table(MstJourneyBoardTable)

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
	session := c.DB.MasterDB.Table(MstJourneyBoardTable)
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

// func (c *JourneyDB) InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {
// 	session := c.DB.MasterDB.Table(MstJourneyPointTable)
// 	_, err = session.InsertOne(journeyPoint)
// 	if err != nil {
// 		err = errors.Wrap(err, WrapMsgInserNewJourneyPoint)
// 		return
// 	}

// 	return
// }

func (c *JourneyDB) InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {
	session := c.DB.MasterDB.Table(MstJourneyPointTable)
	_, err = session.SQL(
		`INSERT INTO mdl_mst_journey_point (name, id_mst_board, position, create_time, update_time)
		SELECT ?, ?, COALESCE(
		(select j.position+5 FROM mdl_mst_journey_point j where j.id_mst_board = ? order by j.id desc limit 1), 0),
		now(), now()
		FROM mdl_mst_journey_board jb
		WHERE jb.id_mst_institution
		RETURNING id, name, id_mst_board, position
		`, journeyPoint.Name, journeyPoint.IDMstJourneyBoard, journeyPoint.IDMstJourneyBoard).
		Insert(journeyPoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInserNewJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, count int64, err error) {
	session := c.DB.SlaveDB.Table(MstJourneyPointTable)

	if len(params.IDs) > 0 {
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
		Where("id_mst_board = ?", params.IDMstBoard).
		FindAndCount(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListJourneyPoints)
		return
	}

	return
}

func (c *JourneyDB) GetJourneyPoint(ctx context.Context, id int64) (resp model.MstJourneyPoint, err error) {
	session := c.DB.SlaveDB.Table(MstJourneyPointTable)

	_, err = session.
		Where("id = ?", id).
		Get(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyPoints)
		return
	}

	return
}

func (c *JourneyDB) UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {
	session := c.DB.MasterDB.Table(MstJourneyPointTable)

	_, err = session.Where("id = ?", journeyPoint.ID).Update(journeyPoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		return
	}

	return
}

func (c *JourneyDB) DeleteJourneyPoint(ctx context.Context, id int64) (err error) {
	session := c.DB.MasterDB.Table(MstJourneyPointTable)

	count, err := session.Delete(model.MstJourneyPoint{
		ID: id,
	})
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
	session := c.DB.MasterDB.Table(MstServicePointTable)

	_, err = session.InsertOne(&mstServicePoint)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInserNewServicePoint)
		return
	}

	return

}

func (c *JourneyDB) ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error) {
	session := c.DB.SlaveDB.Table(MstServicePointTable)

	err = session.
		Where("id_mst_board = ?", params.IDMstBoard).
		Limit(params.Limit, params.Start).
		Find(&resp)
	if err != nil {
		err = errors.Wrap(err, WrapMsgListServicePoints)
		return
	}
	return
}

func (c *JourneyDB) GetServicePoint(ctx context.Context, id int64) (resp model.MstServicePoint, err error) {
	session := c.DB.SlaveDB.Table(MstServicePointTable)

	found, err := session.
		Where("id = ?", id).
		Get(&resp)
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
	session := c.DB.MasterDB.Table(MstServicePointTable)

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

func (c *JourneyDB) DeleteServicePoint(ctx context.Context, id int64) (err error) {
	session := c.DB.MasterDB.Table(MstServicePointTable)

	count, err := session.
		Delete(model.MstServicePoint{ID: id})
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
