package model

import (
	"time"

	"github.com/faisalhardin/medilink/internal/library/util/shortid"
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

// model for mdl_map_service_point_journey_point
type MapServicePointJourneyPoint struct {
	IDMstJourneyPoint int64 `xorm:"'id_mst_journey_point'"`
	IDMstServicePoint int64 `xorm:"'id_mst_service_point'"`
}

// model for mdl_mst_journey_point
type GetJourneyPointParams struct {
	IDs              []int64 `schema:"ids"`
	ID               int64
	Name             []string `schema:"name"`
	IDMstBoard       int64    `schema:"board_id"`
	IDMstInstitution int64
	CommonRequestPayload
}

type JourneyBoardJoinJourneyPoint struct {
	MstJourneyBoard `xorm:"extends"`
	JourneyPoints   []MstJourneyPoint `xorm:"mst_journey_point" json:"journey_points"`
}

type GetServicePointParams struct {
	ID               []int64  `schema:"id"`
	Name             []string `schema:"name"`
	IDMstBoard       int64    `schema:"board_id"`
	IDMstInstitution int64
	CommonRequestPayload
}

type RenameJourneyPointRequest struct {
	ID   int64  `xorm:"'id' pk autoincr" json:"id" validation:"required"`
	Name string `xorm:"'name'" json:"name" validation:"required"`
}

type ArchiveJourneyPointRequest struct {
	ShortID string `json:"id" validation:"required"`
}
