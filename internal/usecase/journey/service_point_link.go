package journey

import (
	"context"
	"sort"

	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
)

func (u *JourneyUC) GetRoomJourneyPoints(ctx context.Context, roomID int64) (model.RoomJourneyPoints, error) {
	room, err := u.GetServicePoint(ctx, roomID)
	if err != nil {
		return model.RoomJourneyPoints{}, err
	}
	ids := room.JourneyPointIDs
	if ids == nil {
		ids = []int64{}
	}
	return model.RoomJourneyPoints{JourneyPointIDs: ids}, nil
}

func (u *JourneyUC) ReplaceLinks(ctx context.Context, side model.LinkKind, subjectID int64, targets []int64) (ids []int64, err error) {
	ids = uniqueIDs(targets)
	var board int64
	defer func() {
		if err == nil {
			return
		}
		logged := commonerr.RoomIDs{Board: board}
		if user, ok := auth.GetUserDetailFromCtx(ctx); ok {
			logged.Institution = user.InstitutionID
		}
		if side == model.LinkRoom {
			logged.Room = subjectID
		} else {
			logged.Column = subjectID
		}
		commonerr.LogRoom(commonerr.RoomOpLinkColumn, logged, err)
	}()

	user, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return nil, err
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xorm.SetDBSession(ctx, session)

	subject, err := u.JourneyDB.ListLinkNodes(ctx, model.LinkQuery{Kind: side, IDs: []int64{subjectID}})
	if err != nil {
		return nil, err
	}
	if node, ok := nodeByID(subject, subjectID); ok {
		board = node.BoardID
		if err = u.JourneyDB.LockJourneyBoard(ctx, board); err != nil {
			err = roomLookupErr(err)
			return nil, err
		}
	}

	foundNodes, err := u.JourneyDB.ListLinkNodes(ctx, model.LinkQuery{Kind: side.Other(), IDs: ids})
	if err != nil {
		return nil, err
	}
	if err = checkLinkTargets(subjectID, subject, ids, foundNodes); err != nil {
		return nil, err
	}

	err = u.JourneyDB.ReplaceLinks(ctx, model.LinkSet{
		Side:    side,
		Clinic:  user.InstitutionID,
		Board:   board,
		Subject: subjectID,
		Targets: ids,
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (u *JourneyUC) GetColumnRooms(ctx context.Context, columnID int64) (model.JourneyPointRooms, error) {
	nodes, err := u.JourneyDB.ListLinkNodes(ctx, model.LinkQuery{
		Kind: model.LinkColumn,
		IDs:  []int64{columnID},
	})
	if err != nil {
		return model.JourneyPointRooms{}, err
	}
	if err = checkLinkTargets(columnID, nodes, nil, nil); err != nil {
		return model.JourneyPointRooms{}, err
	}
	rooms, err := u.JourneyDB.ListColumnRooms(ctx, 0, columnID)
	if err != nil {
		return model.JourneyPointRooms{}, err
	}
	return newJourneyPointRooms(rooms), nil
}

func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func nodeByID(nodes []model.LinkNode, id int64) (model.LinkNode, bool) {
	for _, node := range nodes {
		if node.ID == id {
			return node, true
		}
	}
	return model.LinkNode{}, false
}

func checkLinkTargets(subjectID int64, subject []model.LinkNode, want []int64, targets []model.LinkNode) error {
	sub, ok := nodeByID(subject, subjectID)
	if !ok {
		return roomLookupErr(constant.ErrorNoAffectedRow)
	}
	byID := make(map[int64]model.LinkNode, len(targets))
	for _, node := range targets {
		byID[node.ID] = node
	}
	for _, id := range want {
		if _, found := byID[id]; !found {
			return roomLookupErr(constant.ErrorNoAffectedRow)
		}
	}
	for _, id := range want {
		if byID[id].BoardID != sub.BoardID {
			return commonerr.RoomCrossBoard()
		}
	}
	return nil
}

func newJourneyPointRooms(rooms []model.PickerRoom) model.JourneyPointRooms {
	items := make([]model.PickerRoom, 0, len(rooms))
	for _, room := range rooms {
		room.IsClinical = false
		items = append(items, room)
	}
	return model.JourneyPointRooms{
		PickerRequired: len(items) > 1,
		Items:          items,
	}
}
