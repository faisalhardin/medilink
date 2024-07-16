package model

import "time"

type Patient struct {
	ID                    int64     `xorm:"'id'"`
	Name                  string    `xorm:"'name'"`
	PlaceOfBirth          string    `xorm:"'place_of_birth'"`
	DateOfBirth           time.Time `xorm:"'data_of_birth'"`
	RegisterInstitutionID int64     `xorm:"'registered_institution_id'"`
	CreateTime            time.Time `xorm:"'create_time' created"`
	UpdateTime            time.Time `xorm:"'update_time' updated"`
	DeleteTime            time.Time `xorm:"'delete_time' deleted"`
}
