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

	InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	ArchiveJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)

	GetServicePoint(ctx context.Context, servicePointID int64) (servicePoint *model.MstServicePoint, err error)
	ListServicePoints(ctx context.Context, params model.GetServicePointParams) (servicePoints []model.MstServicePoint, err error)
	InsertNewServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error)
	UpdateServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error)
	ArchiveServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error)
}
