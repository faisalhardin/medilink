package model

import (
	"time"
)

const (
	MST_STAFF_TABLE      = "mdl_mst_staff"
	MST_ROLE_TABLE       = "mdl_mst_role"
	MAP_ROLE_STAFF       = "mdl_map_role_staff"
	MST_PERMISSION_TABLE = "mdl_mst_permission"
	MAP_ROLE_PERMISSION  = "mdl_map_role_permission"
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
	Permissions     []string            `json:"permissions"`
	PermissionsSet  map[string]bool     `json:"-"`
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

type MstPermission struct {
	ID          int64      `json:"-" xorm:"'id' pk autoincr"`
	Code        string     `json:"code" xorm:"'code' unique notnull"`
	Resource    string     `json:"resource" xorm:"'resource'"`
	Action      string     `json:"action" xorm:"'action'"`
	Description string     `json:"description" xorm:"'description'"`
	CreateTime  time.Time  `json:"-" xorm:"'create_time' created"`
	UpdateTime  time.Time  `json:"-" xorm:"'update_time' updated"`
	DeleteTime  *time.Time `json:"-" xorm:"'delete_time' deleted"`
}

func (MstPermission) TableName() string {
	return MST_PERMISSION_TABLE
}

type RolePermissionMapping struct {
	ID           int64 `xorm:"'id' pk autoincr"`
	RoleID       int64 `xorm:"'id_mst_role'"`
	PermissionID int64 `xorm:"'id_mst_permission'"`
}

func (RolePermissionMapping) TableName() string {
	return MAP_ROLE_PERMISSION
}

type UserDetail struct {
	Staff       MstStaff    `json:"staff" xorm:"extends"`
	Institution Institution `json:"institution" xorm:"extends"`
	Roles       []MstRole   `json:"roles" xorm:"roles"`
}

func GenerateUserDetailSessionInformation(u UserDetail, expiredTime time.Time) UserSessionDetail {

	return UserSessionDetail{
		UserID:           u.Staff.ID,
		Name:             u.Staff.Name,
		IdMstInstitution: u.Staff.IdMstInstitution,
		ExpiredAt:        expiredTime.Unix(),
	}
}

func GenerateUserDataJWTInformation(internalUserDetail UserDetail, externalUserDetail GoogleUser, journeyPoints []MstJourneyPoint, servicePoints []MstServicePoint, permissions []MstPermission) UserJWTPayload {
	userRoles := make([]UserRoleJWTDetail, 0, len(internalUserDetail.Roles))
	rolesIDSet := make(map[string]bool, len(internalUserDetail.Roles))

	for _, role := range internalUserDetail.Roles {
		userRoles = append(userRoles, UserRoleJWTDetail{
			RoleID: role.RoleID,
			Name:   role.Name,
		})
		if role.Name != "" {
			rolesIDSet[role.Name] = true
		}
	}

	permissionCodes := make([]string, 0, len(permissions))
	permissionsSet := make(map[string]bool, len(permissions))
	for _, p := range permissions {
		permissionCodes = append(permissionCodes, p.Code)
		permissionsSet[p.Code] = true
	}

	return UserJWTPayload{
		UUID:            internalUserDetail.Staff.UUID,
		Name:            internalUserDetail.Staff.Name,
		Email:           internalUserDetail.Staff.Email,
		Roles:           userRoles,
		RolesIDSet:      rolesIDSet,
		Permissions:     permissionCodes,
		PermissionsSet:  permissionsSet,
		InstitutionName: internalUserDetail.Institution.Name,
		ImageURL:        externalUserDetail.Picture,
		ProviderUserID:  externalUserDetail.ID,
		JourneyPoints:   journeyPoints,
		ServicePoints:   servicePoints,
	}
}

func (p *UserJWTPayload) EnsureAuthSets() {
	if p.RolesIDSet == nil {
		p.RolesIDSet = make(map[string]bool, len(p.Roles))
		for _, role := range p.Roles {
			if role.Name != "" {
				p.RolesIDSet[role.Name] = true
			}
		}
	}

	if p.PermissionsSet == nil {
		p.PermissionsSet = make(map[string]bool, len(p.Permissions))
		for _, code := range p.Permissions {
			p.PermissionsSet[code] = true
		}
	}
}

type ListStaffParams struct {
	IncludeInactive bool `schema:"include_inactive"`
}

type CreateStaffRequest struct {
	Name    string  `json:"name" validate:"required"`
	Email   string  `json:"email" validate:"required,email"`
	RoleIDs []int64 `json:"role_ids" validate:"required,min=1,dive,gt=0"`
}

type UpdateStaffRequest struct {
	UUID string `json:"uuid" validate:"required"`
	Name string `json:"name" validate:"required"`
}

type AssignRoleRequest struct {
	StaffUUID string `json:"staff_uuid" validate:"required"`
	RoleID    int64  `json:"role_id" validate:"required"`
}

type UnassignRoleRequest struct {
	StaffUUID string `json:"staff_uuid" validate:"required"`
	RoleID    int64  `json:"role_id" validate:"required"`
}

type StaffStatusRequest struct {
	UUID string `json:"uuid" validate:"required"`
}

type StaffRoleResponse struct {
	RoleID int64  `json:"role_id"`
	Name   string `json:"name"`
}

type StaffWithRolesResponse struct {
	UUID            string              `json:"uuid"`
	Name            string              `json:"name"`
	Email           string              `json:"email"`
	InstitutionID   int64               `json:"institution_id"`
	InstitutionName string              `json:"institution_name"`
	Roles           []StaffRoleResponse `json:"roles"`
	IsActive        bool                `json:"is_active"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type ListStaffResponse struct {
	Staff []StaffWithRolesResponse `json:"staff"`
	Total int                      `json:"total"`
}

func RoleNamesFromUserDetail(userDetail UserDetail) []string {
	roleNames := make([]string, 0, len(userDetail.Roles))
	for _, role := range userDetail.Roles {
		if role.Name != "" {
			roleNames = append(roleNames, role.Name)
		}
	}
	return roleNames
}
