package model

// type UserAuth struct {
// 	ID          int64          `json:"id"`
// 	UserUUID    string         `json:"uuid"`
// 	Name        string         `json:"name"`
// 	Email       string         `json:"email"`
// 	CompanyUUID string         `json:"company_uuid"`
// 	CompanyName string         `json:"company_name"`
// 	CompanyID   int64          `json:"company_id"`
// 	Roles       []UserAuthRole `json:"roles"`
// 	RolesIDSet  map[string]bool
// }

type GoogleUser struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	FirstName string `json:"given_name"`
	LastName  string `json:"family_name"`
	Link      string `json:"link"`
	Picture   string `json:"picture"`
}

// type UserAuthRole struct {
// 	RoleID int64  `xorm:"'role_id'"`
// 	Name   string `json:"name,omitempty" xorm:"'name'"`
// }
