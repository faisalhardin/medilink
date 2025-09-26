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

	WrapMsgGetServicePointMappedByJourneyPoints = PrefixUserJourneyDB + ".GetServicePointMappedByJourneyPoints"
	WrapMsgInsertNewMapStaffJourneyPoint        = PrefixUserJourneyDB + ".InsertNewMapStaffJourneyPoint"
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

func (c *UserJourneyDB) GetJourneyBoardByParams(ctx context.Context, params model.MstJourneyBoard) (resp model.MstJourneyBoard, err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.GetJourneyBoardByParams(ctx, params)
}

func (c *UserJourneyDB) GetJourneyBoardByJourneyPoint(ctx context.Context, journeyPoint model.MstJourneyPoint) (resp model.MstJourneyBoard, err error) {
	return c.JourneyDB.GetJourneyBoardByJourneyPoint(ctx, journeyPoint)
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

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	journeyPoint.MstJourneyPoint.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.InsertNewJourneyPoint(ctx, journeyPoint)
}

func (c *UserJourneyDB) ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.ListJourneyPointResponse, count int64, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgListJourneyPoints)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.IDMstInstitution = userDetail.InstitutionID
	params.StaffID = userDetail.UserID

	return c.JourneyDB.ListJourneyPoints(ctx, params)
}

func (c *UserJourneyDB) GetJourneyPoint(ctx context.Context, param model.MstJourneyPoint) (resp *model.MstJourneyPoint, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetJourneyPoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	param.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.GetJourneyPoint(ctx, param)
}

func (c *UserJourneyDB) ListJourneyPointsWithoutShortID(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, err error) {
	return c.JourneyDB.ListJourneyPointsWithoutShortID(ctx, params)
}

// GetJourneyPointByShortID retrieves a journey point by its short ID
func (c *UserJourneyDB) GetJourneyPointByShortID(ctx context.Context, shortID string) (resp *model.MstJourneyPoint, err error) {
	return c.JourneyDB.GetJourneyPointByShortID(ctx, shortID)
}

func (c *UserJourneyDB) UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgUpdateJourneyPoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	journeyPoint.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.UpdateJourneyPoint(ctx, journeyPoint)
}

func (c *UserJourneyDB) DeleteJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgDeleteJourneyPoint)
		}
	}()

	return c.JourneyDB.DeleteJourneyPoint(ctx, journeyPoint)
}

func (c *UserJourneyDB) InserNewServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgInserNewServicePoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstServicePoint.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.InserNewServicePoint(ctx, mstServicePoint)
}

func (c *UserJourneyDB) ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgListServicePoints)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	params.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.ListServicePoints(ctx, params)
}

func (c *UserJourneyDB) GetServicePoint(ctx context.Context, servicePoint model.MstServicePoint) (resp *model.MstServicePoint, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetServicePoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	servicePoint.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.GetServicePoint(ctx, servicePoint)
}

func (c *UserJourneyDB) UpdateServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgUpdateServicePoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstServicePoint.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.UpdateServicePoint(ctx, mstServicePoint)
}

func (c *UserJourneyDB) DeleteServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgDeleteServicePoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstServicePoint.IDMstInstitution = userDetail.InstitutionID

	return c.JourneyDB.DeleteServicePoint(ctx, mstServicePoint)
}

func (c *UserJourneyDB) GetJourneyPointMappedByStaff(ctx context.Context, mstStaff model.MstStaff) (journeyPoint []model.MstJourneyPoint, err error) {

	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgDeleteServicePoint)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstStaff.ID = userDetail.UserID
	return c.JourneyDB.GetJourneyPointMappedByStaff(ctx, mstStaff)
}

func (c *UserJourneyDB) GetServicePointMappedByJourneyPoints(ctx context.Context, journeyPoints []model.MstJourneyPoint, mstStaff model.MstStaff) (servicePoints []model.MstServicePoint, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgGetServicePointMappedByJourneyPoints)
		}
	}()

	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
	}

	mstStaff.ID = userDetail.UserID
	return c.JourneyDB.GetServicePointMappedByJourneyPoints(ctx, journeyPoints, mstStaff)

}

func (c *UserJourneyDB) InsertNewMapStaffJourneyPoint(ctx context.Context, mapStaffJourneyPoint *model.MapStaffJourneyPoint) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, WrapMsgInsertNewMapStaffJourneyPoint)
		}
	}()

	return c.JourneyDB.InsertNewMapStaffJourneyPoint(ctx, mapStaffJourneyPoint)
}
