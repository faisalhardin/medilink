package user

import "time"

type User struct {
	UserID     int64      `json:"-" xorm:"'id' pk autoincr"`
	UUID       string     `json:"uuid" xorm:"'uuid'"`
	Email      string     `json:"email" validate:"email" xorm:"email"`
	Domain     float64    `json:"domain" xorm:"-"`
	CreateTime time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime *time.Time `json:"-" xorm:"'delete_time' deleted"`
}
