package customerjourney

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyIface "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"

	"github.com/pkg/errors"
)

const (
	PrefixUserJourneyDB          = "UserJourneyDB"
	WrapMsgInsertNewJourneyBoard = PrefixUserJourneyDB + ".InsertNewJourneyBoard"
	WrapMsgListJourneyBoard      = PrefixUserJourneyDB + ".ListJourneyBoard"
	WrapMsgGetJourneyBoardDetail = PrefixUserJourneyDB + ".GetJourneyBoardDetail"
	WrapMsgUpdateJourneyBoard    = PrefixUserJourneyDB + ".UpdateJourneyBoard"
	WrapMsgDeleteJourneyBoard    = PrefixUserJourneyDB + ".DeleteJourneyBoard"

	WrapMsgInserNewJourneyPoint = PrefixUserJourneyDB + ".InsertNewJourneyPoint"
	WrapMsgListJourneyPoints    = PrefixUserJourneyDB + ".ListJourneyPoints"
	WrapMsgGetJourneyPoint      = PrefixUserJourneyDB + ".GetJourneyPoints"
	WrapMsgUpdateJourneyPoint   = PrefixUserJourneyDB + ".UpdateJourneyPoint"
	WrapMsgDeleteJourneyPoint   = PrefixUserJourneyDB + ".DeleteJourneyPoint"

	WrapMsgInserNewServicePoint = PrefixUserJourneyDB + ".InserNewServicePoint"
	WrapMsgListServicePoints    = PrefixUserJourneyDB + ".ListServicePoints"
	WrapMsgGetServicePoint      = PrefixUserJourneyDB + ".GetServicePoint"
	WrapMsgUpdateServicePoint   = PrefixUserJourneyDB + ".UpdateServicePoint"
	WrapMsgDeleteServicePoint   = PrefixUserJourneyDB + ".DeleteServicePoint"
)

type UserJourneyDB struct {
	JourneyDB journeyIface.JourneyDB
}

func NewUserJourneyDB(journeyDB *UserJourneyDB) *UserJourneyDB {
	return journeyDB
}

func (c *UserJourneyDB) InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	journeyBoard.IDMstInstitution = userDetail.InstitutionID

	err = c.JourneyDB.InsertNewJourneyBoard(ctx, journeyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyBoard)
	}

	return

}

func (c *UserJourneyDB) GetJourneyBoardByID(ctx context.Context, id int64) (resp model.MstJourneyBoard, err error) {
	return c.JourneyDB.GetJourneyBoardByID(ctx, id)
}

func (c *UserJourneyDB) GetJourneyBoardByJourneyPoint(ctx context.Context, journeyPointID int64) (resp model.MstJourneyBoard, err error) {
	return c.JourneyDB.GetJourneyBoardByJourneyPoint(ctx, journeyPointID)
}

func (c *UserJourneyDB) ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgListJourneyBoard)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.ListJourneyBoard(ctx, params)
}

func (c *UserJourneyDB) GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (resp model.JourneyBoardJoinJourneyPoint, found bool, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetJourneyBoardDetail)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.GetJourneyBoardDetail(ctx, params)
}

func (c *UserJourneyDB) UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgUpdateJourneyBoard)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	journeyBoard.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.UpdateJourneyBoard(ctx, journeyBoard)
}

func (c *UserJourneyDB) DeleteJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgDeleteJourneyBoard)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	journeyBoard.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.DeleteJourneyBoard(ctx, journeyBoard)
}

func (c *UserJourneyDB) InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.InsertMstJourneyPoint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgListJourneyBoard)
		}
	}()

	return c.JourneyDB.InsertNewJourneyPoint(ctx, journeyPoint)
}

func (c *UserJourneyDB) ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, count int64, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgListJourneyPoints)
		}
	}()

	return c.JourneyDB.ListJourneyPoints(ctx, params)
}

func (c *UserJourneyDB) GetJourneyPoint(ctx context.Context, id int64) (resp model.MstJourneyPoint, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetJourneyPoint)
		}
	}()

	return c.JourneyDB.GetJourneyPoint(ctx, id)
}

func (c *UserJourneyDB) UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		}
	}()

	return c.JourneyDB.UpdateJourneyPoint(ctx, journeyPoint)
}

func (c *UserJourneyDB) DeleteJourneyPoint(ctx context.Context, id int64) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgDeleteJourneyPoint)
		}
	}()

	return c.JourneyDB.DeleteJourneyPoint(ctx, id)
}

func (c *UserJourneyDB) InserNewServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgInserNewServicePoint)
		}
	}()

	return c.JourneyDB.InserNewServicePoint(ctx, mstServicePoint)
}

func (c *UserJourneyDB) ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgListServicePoints)
		}
	}()

	return c.JourneyDB.ListServicePoints(ctx, params)
}

func (c *UserJourneyDB) GetServicePoint(ctx context.Context, id int64) (resp model.MstServicePoint, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetServicePoint)
		}
	}()

	return c.JourneyDB.GetServicePoint(ctx, id)
}

func (c *UserJourneyDB) UpdateServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgUpdateServicePoint)
		}
	}()

	return c.JourneyDB.UpdateServicePoint(ctx, mstServicePoint)
}

func (c *UserJourneyDB) DeleteServicePoint(ctx context.Context, id int64) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgDeleteServicePoint)
		}
	}()

	return c.JourneyDB.DeleteServicePoint(ctx, id)
}
