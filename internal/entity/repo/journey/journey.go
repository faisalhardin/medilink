package journey

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type JourneyDB interface {
	InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
	ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error)
	GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (resp model.JourneyBoardJoinJourneyPoint, found bool, err error)
	UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
	DeleteJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)

	InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, count int64, err error)
	GetJourneyPoint(ctx context.Context, id int64) (resp model.MstJourneyPoint, err error)
	UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	DeleteJourneyPoint(ctx context.Context, id int64) (err error)

	InserNewServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error)
	ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error)
	GetServicePoint(ctx context.Context, id int64) (resp model.MstServicePoint, err error)
	UpdateServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error)
	DeleteServicePoint(ctx context.Context, id int64) (err error)
}