package model

import "time"

const (
	MST_STAFF_TABLE = "mdl_mst_staff"
	MST_ROLE_TABLE  = "mdl_mst_role"
	MAP_ROLE_STAFF  = "mdl_map_role_staff"
)

type MstStaff struct {
	ID               int64      `xorm:"'id' pk autoincr" json:"-"`
	UUID             string     `xorm:"'uuid'" json:"uuid"`
	Name             string     `xorm:"'name'" json:"name"`
	IdMstInstitution int64      `xorm:"'id_mst_institution'" json:"id_mst_institution"`
	CreateTime       time.Time  `xorm:"'create_time' created" json:"-"`
	UpdateTime       time.Time  `xorm:"'update_time' updated" json:"-"`
	DeleteTime       *time.Time `xorm:"'delete_time' deleted" json:"-"`
}

type MstRole struct {
	ID         int64      `json:"-" xorm:"'id' pk autoincr"`
	RoleID     int64      `json:"role_id" xorm:"'role_id'"`
	Name       string     `json:"name" xorm:"'name'"`
	CreateTime time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type RoleStaffMapping struct {
	ID      int64 `xorm:"'id' pk autoincr"`
	StaffID int64 `xorm:"'id_mst_staff'"`
	RoleID  int64 `xorm:"'id_mst_role'"`
}
