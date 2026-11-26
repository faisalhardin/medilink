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

	ListJourneyPoints(ctx context.Context, params model.GetJourneyPointParams) (servicePoint []model.ListJourneyPointResponse, err error)
	InsertNewJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (response model.InsertJourneyPointResponse, err error)
	UpdateJourneyPoint(ctx context.Context, journeyPoint *model.MstJourneyPoint) (err error)
	ArchiveJourneyPoint(ctx context.Context, journeyPoint *model.ArchiveJourneyPointRequest) (err error)

	GetServicePoint(ctx context.Context, servicePointID int64) (servicePoint *model.Room, err error)
	GetRoomJourneyPoints(ctx context.Context, roomID int64) (points model.RoomJourneyPoints, err error)
	ReplaceLinks(ctx context.Context, side model.LinkKind, subjectID int64, targets []int64) (ids []int64, err error)
	GetColumnRooms(ctx context.Context, columnID int64) (rooms model.JourneyPointRooms, err error)
	ListServicePoints(ctx context.Context, params model.GetServicePointParams) (servicePoints []model.Room, err error)
	InsertNewServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error)
	UpdateServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error)
	ArchiveServicePoint(ctx context.Context, servicePoint *model.MstServicePoint) (err error)
}
