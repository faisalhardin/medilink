package model

import (
	"time"
)

const (
	MST_STAFF_TABLE = "mdl_mst_staff"
	MST_ROLE_TABLE  = "mdl_mst_role"
	MAP_ROLE_STAFF  = "mdl_map_role_staff"
)

type MstStaff struct {
	ID               int64      `xorm:"'id' pk autoincr" json:"-"`
	UUID             string     `xorm:"'uuid'" json:"uuid,omitempty"`
	Name             string     `xorm:"'name'" json:"name,omitempty"`
	Email            string     `xorm:"email" json:"email,omitempty"`
	IdMstInstitution int64      `xorm:"'id_mst_institution'" json:"-"`
	CreateTime       time.Time  `xorm:"'create_time' created" json:"-"`
	UpdateTime       time.Time  `xorm:"'update_time' updated" json:"-"`
	DeleteTime       *time.Time `xorm:"'delete_time' deleted" json:"-"`
}

type UserSessionDetail struct {
	UserID           int64  `json:"id"`
	Name             string `json:"name"`
	IdMstInstitution int64  `json:"id_mst_institution"`
	ExpiredAt        int64  `json:"expired_at"`
}

type UserJWTPayload struct {
	UserID          int64               `json:"id,omitempty"`
	UUID            string              `json:"uuid,omitempty"`
	Name            string              `json:"name,omitempty"`
	Email           string              `json:"email,omitempty"`
	InstitutionID   int64               `json:"institution_id,omitempty"`
	InstitutionUUID string              `json:"institution_uuid,omitempty"`
	InstitutionName string              `json:"institution_name,omitempty"`
	Roles           []UserRoleJWTDetail `json:"roles"`
	RolesIDSet      map[string]bool     `json:"-"`
	ImageURL        string              `json:"image_url"`
	ProviderUserID  string              `json:"provider_user_id"`
	JourneyPoints   []MstJourneyPoint   `json:"journey_points"`
	ServicePoints   []MstServicePoint   `json:"service_points"`
}

type UserRoleJWTDetail struct {
	RoleID int64  `json:"role_id"`
	Name   string `json:"name,omitempty" xorm:"'name'"`
}

type MstRole struct {
	ID         int64      `json:"-" xorm:"'id' pk autoincr"`
	RoleID     int64      `json:"role_id,omitempty" xorm:"'role_id'"`
	Name       string     `json:"name,omitempty" xorm:"'name'"`
	CreateTime time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

type RoleStaffMapping struct {
	ID      int64 `xorm:"'id' pk autoincr"`
	StaffID int64 `xorm:"'id_mst_staff'"`
	RoleID  int64 `xorm:"'id_mst_role'"`
}

type UserDetail struct {
	Staff MstStaff  `json:"staff" xorm:"extends"`
	Roles []MstRole `json:"roles" xorm:"roles"`
}

func GenerateUserDetailSessionInformation(u UserDetail, expiredTime time.Time) UserSessionDetail {

	return UserSessionDetail{
		UserID:           u.Staff.ID,
		Name:             u.Staff.Name,
		IdMstInstitution: u.Staff.IdMstInstitution,
		ExpiredAt:        expiredTime.Unix(),
	}
}

func GenerateUserDataJWTInformation(internalUserDetail UserDetail, externalUserDetail GoogleUser, journeyPoints []MstJourneyPoint, servicePoints []MstServicePoint) UserJWTPayload {
	userRoles := []UserRoleJWTDetail{}

	for _, role := range internalUserDetail.Roles {
		userRoles = append(userRoles, UserRoleJWTDetail{
			RoleID: role.RoleID,
			Name:   role.Name,
		})
	}
	return UserJWTPayload{
		UUID:           internalUserDetail.Staff.UUID,
		Name:           internalUserDetail.Staff.Name,
		Email:          internalUserDetail.Staff.Email,
		Roles:          userRoles,
		ImageURL:       externalUserDetail.Picture,
		ProviderUserID: externalUserDetail.ID,
		JourneyPoints:  journeyPoints,
		ServicePoints:  servicePoints,
	}
}
