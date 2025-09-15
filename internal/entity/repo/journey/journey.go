package journey

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type JourneyDB interface {
	InsertNewJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
	ListJourneyBoard(ctx context.Context, params model.GetJourneyBoardParams) (journeyBoards []model.MstJourneyBoard, err error)
	GetJourneyBoardByID(ctx context.Context, id int64) (resp model.MstJourneyBoard, err error)
	GetJourneyBoardByJourneyPoint(ctx context.Context, journeyPointID int64) (resp model.MstJourneyBoard, err error)
	GetJourneyBoardDetail(ctx context.Context, params model.GetJourneyBoardParams) (resp model.JourneyBoardJoinJourneyPoint, found bool, err error)
	UpdateJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)
	DeleteJourneyBoard(ctx context.Context, journeyBoard *model.MstJourneyBoard) (err error)

	InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.InsertMstJourneyPoint) (err error)
	ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, count int64, err error)
	ListJourneyPointsWithoutShortID(ctx context.Context, params model.GetJourneyPointParams) (resp []model.MstJourneyPoint, err error)
	GetJourneyPointByShortID(ctx context.Context, shortID string) (resp *model.MstJourneyPoint, err error)
	GetJourneyPoint(ctx context.Context, param model.MstJourneyPoint) (resp *model.MstJourneyPoint, err error)
	UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	DeleteJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	GetJourneyPointMappedByStaff(ctx context.Context, mstStaff model.MstStaff) (journeyPoint []model.MstJourneyPoint, err error)
	GetServicePointMappedByJourneyPoints(ctx context.Context, journeyPoints []model.MstJourneyPoint, mstStaff model.MstStaff) (servicePoints []model.MstServicePoint, err error)

	InserNewServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error)
	ListServicePoints(ctx context.Context, params model.GetServicePointParams) (resp []model.MstServicePoint, err error)
	GetServicePoint(ctx context.Context, servicePoint model.MstServicePoint) (resp *model.MstServicePoint, err error)
	UpdateServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error)
	DeleteServicePoint(ctx context.Context, mstServicePoint *model.MstServicePoint) (err error)
}
