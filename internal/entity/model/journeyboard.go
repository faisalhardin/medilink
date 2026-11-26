package model

import (
	"time"

	"github.com/faisalhardin/medilink/internal/library/util/shortid"
	"github.com/volatiletech/null/v8"
)

// model for mdl_mst_journey_board
type MstJourneyBoard struct {
	ID               int64      `xorm:"'id' pk autoincr" json:"id"`
	Name             string     `xorm:"'name'" json:"name"`
	IDMstInstitution int64      `xorm:"'id_mst_institution'" json:"-"`
	CreateTime       time.Time  `xorm:"'create_time' created" json:"create_time"`
	UpdateTime       time.Time  `xorm:"'update_time' updated" json:"update_time"`
	DeleteTime       *time.Time `xorm:"'delete_time' deleted" json:"-" `
}

type GetJourneyBoardParams struct {
	ID               []int64  `json:"id" schema:"id"`
	Name             []string `json:"name" schema:"name"`
	IDMstInstitution int64    `json:"id_mst_institution" schema:"id_mst_institution"`
}

// model for mdl_mst_journey_point
type MstJourneyPoint struct {
	ID                int64  `xorm:"'id' pk autoincr" json:"-"`
	ShortID           string `xorm:"'short_id' unique" json:"id,omitempty"`
	Name              string `xorm:"'name'" json:"name,omitempty"`
	Position          int    `xorm:"'position'" json:"position,omitempty"`
	IDMstJourneyBoard int64  `xorm:"'id_mst_journey_board'" json:"board_id,omitempty"`
	IDMstInstitution  int64  `xorm:"id_mst_institution" json:"-"`

	CreateTime time.Time  `xorm:"'create_time' created" json:"create_time"`
	UpdateTime time.Time  `xorm:"'update_time' updated" json:"update_time"`
	DeleteTime *time.Time `xorm:"'delete_time' deleted" json:"-" `
}

type ListJourneyPointResponse struct {
	MstJourneyPoint `xorm:"extends"`
	IsOwned         null.Bool `json:"is_owned"`
}

type InsertJourneyPointResponse struct {
	MstJourneyPoint `xorm:"extends"`
	IsOwned         null.Bool `json:"is_owned"`
}

// BeforeInsert generates a short ID before inserting the journey point
func (m *MstJourneyPoint) BeforeInsert() {
	if m.ShortID == "" {
		shortID, err := shortid.GenerateShortID(8)
		if err == nil {
			m.ShortID = shortID
		}
	}
}

type InsertMstJourneyPoint struct {
	MstJourneyPoint *MstJourneyPoint
}

// model for mdl_mst_service_point
type MstServicePoint struct {
	ID               int64  `xorm:"'id' pk autoincr" json:"id"`
	Name             string `xorm:"'name'" json:"name"`
	IDMstBoard       int64  `xorm:"'id_mst_journey_board'" json:"board_id"`
	IDMstInstitution int64  `xorm:"id_mst_institution" json:"-"`

	CreateTime time.Time  `xorm:"'create_time' created" json:"create_time"`
	UpdateTime time.Time  `xorm:"'update_time' updated" json:"update_time"`
	DeleteTime *time.Time `xorm:"'delete_time' deleted" json:"-" `
}

// CreateRoomRequest is a new room. The clinic comes from the session.
type CreateRoomRequest struct {
	Name       string `json:"name"`
	BoardID    int64  `json:"board_id"`
	IsClinical bool   `json:"is_clinical"`
}

// RenameRoomRequest changes a room. A nil Name keeps the stored name.
type RenameRoomRequest struct {
	Name       *string `json:"name"`
	IsClinical bool    `json:"is_clinical"`
}

type SyncStatus string

const (
	SyncPending    SyncStatus = "pending"
	SyncProcessing SyncStatus = "processing"
	SyncSynced     SyncStatus = "synced"
	SyncFailed     SyncStatus = "failed"
	SyncInactive   SyncStatus = "inactive"
)

func ParseSyncStatus(s string) (SyncStatus, bool) {
	switch status := SyncStatus(s); status {
	case SyncPending, SyncProcessing, SyncSynced, SyncFailed, SyncInactive:
		return status, true
	}
	return "", false
}

// Room is the wire shape of a service point. No field uses omitempty so every
// client sees the same keys: Location is null when unknown and JourneyPointIDs
// is [] when the room has no mapped column.
type Room struct {
	ID              int64         `json:"id"`
	Name            string        `json:"name"`
	BoardID         int64         `json:"board_id"`
	IsClinical      bool          `json:"is_clinical"`
	Archived        bool          `json:"archived"`
	JourneyPointIDs []int64       `json:"journey_point_ids"`
	Location        *RoomLocation `json:"location"`
}

type RoomLocation struct {
	SyncStatus  SyncStatus `json:"sync_status"`
	LastJobType string     `json:"last_job_type"`
	FailureKind *string    `json:"failure_kind"`
	FromClear   bool       `json:"from_clear"`
}

// model for mdl_map_service_point_journey_point
type MapServicePointJourneyPoint struct {
	IDMstJourneyPoint int64 `xorm:"'id_mst_journey_point'"`
	IDMstServicePoint int64 `xorm:"'id_mst_service_point'"`
}

// LinkRoom is a service point. LinkColumn is a journey point.
type LinkKind int

const (
	LinkRoom LinkKind = iota
	LinkColumn
)

func (k LinkKind) Other() LinkKind {
	if k == LinkColumn {
		return LinkRoom
	}
	return LinkColumn
}

type LinkQuery struct {
	Kind   LinkKind
	Clinic int64
	IDs    []int64
}

type LinkNode struct {
	ID      int64
	BoardID int64
}

type LinkSet struct {
	Side    LinkKind
	Clinic  int64
	Board   int64
	Subject int64
	Targets []int64
}

type PutRoomJourneyPointsRequest struct {
	JourneyPointIDs *[]int64 `json:"journey_point_ids"`
}

type PutJourneyPointRoomsRequest struct {
	ServicePointIDs *[]int64 `json:"service_point_ids"`
}

type RoomJourneyPoints struct {
	JourneyPointIDs []int64 `json:"journey_point_ids"`
}

type JourneyPointRooms struct {
	PickerRequired bool         `json:"picker_required"`
	Items          []PickerRoom `json:"items"`
}

type PickerRoom struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	IsClinical bool   `json:"is_clinical"`
}

// model for mdl_mst_journey_point
type GetJourneyPointParams struct {
	IDs              []int64 `schema:"ids"`
	ID               int64
	Name             []string `schema:"name"`
	IDMstBoard       int64    `schema:"board_id"`
	IDMstInstitution int64
	StaffID          int64
	CommonRequestPayload
}

type JourneyBoardJoinJourneyPoint struct {
	MstJourneyBoard `xorm:"extends"`
	JourneyPoints   []ListJourneyPointResponse `xorm:"mst_journey_point" json:"journey_points"`
}

type GetServicePointParams struct {
	ID               []int64  `schema:"id"`
	Name             []string `schema:"name"`
	IDMstBoard       int64    `schema:"board_id"`
	IDMstInstitution int64
	IncludeArchived  bool `schema:"include_archived"`
	CommonRequestPayload
}

type RenameJourneyPointRequest struct {
	ID   string `xorm:"'id' pk autoincr" json:"id" validation:"required"`
	Name string `xorm:"'name'" json:"name" validation:"required"`
}

type ArchiveJourneyPointRequest struct {
	ShortID string `json:"id" validation:"required"`
}
