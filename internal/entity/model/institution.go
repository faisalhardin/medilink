package model

import "time"

type Institution struct {
	ID          int64      `json:"id" xorm:"'id' pk autoincr"`
	Name        string     `json:"name" xorm:"'name'"`
	StaffNumber int32      `json:"staff_number" xorm:"'staff_number'"`
	MaxStaff    int32      `json:"max_staff" xorm:"'max_staff'"`
	CreateTime  time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime  time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime  *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type CreateInstitutionRequest struct {
	Name        string `json:"name" validation:"required"`
	StaffNumber int32  `json:"staff_number"`
	MaxStaff    int32  `json:"max_staff"`
}

type FindInstitutionParams struct {
	ID   int64  `schema:"id"`
	Name string `schema:"name"`
}
