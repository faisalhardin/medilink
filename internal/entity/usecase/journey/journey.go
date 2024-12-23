package journey

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type JourneyUC interface {
	InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
	ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error)
	GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (resp model.JourneyBoardJoinJourneyPoint, err error)
	UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
	DeleteJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
}