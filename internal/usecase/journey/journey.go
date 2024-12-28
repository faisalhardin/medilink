package journey

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	WrapMsgPrefix                = "JourneyUC."
	WrapMsgInsertNewJourneyBoard = WrapMsgPrefix + "InsertNewJourneyBoard"
	WrapMsgListJourneyBoard      = WrapMsgPrefix + "ListJourneyBoard"
	WrapMsgGetJourneyBoardDetail = WrapMsgPrefix + "GetJourneyBoardDetail"
	WrapMsgUpdateJourneyBoard    = WrapMsgPrefix + "UpdateJourneyBoard"
	WrapMsgDeleteJourneyBoard    = WrapMsgPrefix + "DeleteJourneyBoard"

	WrapMsgInsertNewJourneyPoint = WrapMsgPrefix + "InsertNewJourneyPoint"
	WrapMsgUpdateJourneyPoint    = WrapMsgPrefix + "UpdateJourneyPoint"
	WrapMsgArchiveJourneyPoint   = WrapMsgPrefix + "ArchiveJourneyPoint"
)

type JourneyUC struct {
	JourneyDB journeyRepo.JourneyDB
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

func (u *JourneyUC) GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoard model.JourneyBoardJoinJourneyPoint, err error) {
	journeyBoard, _, err = u.JourneyDB.GetJourneyBoardDetail(ctx, params)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetJourneyBoardDetail)
		return
	}

	if len(journeyBoard.JourneyPoints) > 0 && journeyBoard.JourneyPoints[0].ID == 0 {
		journeyBoard.JourneyPoints = []model.MstJourneyPoint{}
	}

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

func (u *JourneyUC) validateJourneyBoardOwnership(ctx context.Context, boardID int64) (err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	mstJourneyBoard, err := u.JourneyDB.GetJourneyBoardByID(ctx, boardID)
	if err != nil {
		return
	}

	if mstJourneyBoard.IDMstInstitution != userDetail.InstitutionID {
		return commonerr.SetNewUnauthorizedAPICall()
	}

	return nil
}

func (u *JourneyUC) validateJourneyPointOwnership(ctx context.Context, journeyPointID int64) (err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = errors.Wrap(commonerr.SetNewUnauthorizedAPICall(), WrapMsgUpdateJourneyPoint)
		return
	}

	journeyBoard, err := u.JourneyDB.GetJourneyBoardByJourneyPoint(ctx, journeyPointID)
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

func (u *JourneyUC) InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {

	err = u.validateJourneyBoardOwnership(ctx, journeyPoint.IDMstJourneyBoard)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyPoint)
		return
	}

	insertJourneyPointRequest := &model.InsertMstJourneyPoint{
		MstJourneyPoint: journeyPoint,
	}

	err = u.JourneyDB.InsertNewJourneyPoint(ctx, insertJourneyPointRequest)
	if err != nil {
		err = errors.Wrap(err, WrapMsgInsertNewJourneyPoint)
		return
	}

	return
}

func (u *JourneyUC) UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {

	err = u.validateJourneyPointOwnership(ctx, journeyPoint.ID)
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

func (u *JourneyUC) ArchiveJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error) {
	err = u.validateJourneyPointOwnership(ctx, journeyPoint.ID)
	if err != nil {
		err = errors.Wrap(err, WrapMsgArchiveJourneyPoint)
		return
	}

	err = u.JourneyDB.DeleteJourneyPoint(ctx, journeyPoint.ID)
	if err != nil {
		err = errors.Wrap(err, WrapMsgArchiveJourneyPoint)
		return
	}

	return
}
