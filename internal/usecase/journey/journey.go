package journey

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/pkg/errors"
)

const (
	WrapMsgPrefix                = "JourneyUC."
	WrapMsgInsertNewJourneyBoard = WrapMsgPrefix + "InsertNewJourneyBoard"
	WrapMsgListJourneyBoard      = WrapMsgPrefix + "ListJourneyBoard"
	WrapMsgGetJourneyBoardDetail = WrapMsgPrefix + "GetJourneyBoardDetail"
	WrapMsgUpdateJourneyBoard    = WrapMsgPrefix + "UpdateJourneyBoard"
	WrapMsgDeleteJourneyBoard    = WrapMsgPrefix + "DeleteJourneyBoard"
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
