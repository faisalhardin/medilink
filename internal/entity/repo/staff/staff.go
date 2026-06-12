package staff

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type StaffDB interface {
	ListStaffByInstitution(ctx context.Context, institutionID int64, includeInactive bool) ([]model.StaffWithRolesResponse, error)
	GetStaffByUUID(ctx context.Context, institutionID int64, uuid string, includeInactive bool) (model.StaffWithRolesResponse, error)
	GetStaffIDByUUID(ctx context.Context, institutionID int64, uuid string) (int64, error)
	InsertStaff(ctx context.Context, staff *model.MstStaff) error
	DeactivateStaff(ctx context.Context, institutionID int64, uuid string) error
	ActivateStaff(ctx context.Context, institutionID int64, uuid string) error
	AssignRole(ctx context.Context, staffID int64, rolePK int64) error
	UnassignRole(ctx context.Context, staffID int64, rolePK int64) error
	HasRoleAssignment(ctx context.Context, staffID int64, rolePK int64) (bool, error)
	GetRolePKByBusinessID(ctx context.Context, roleID int64) (model.MstRole, bool, error)
	GetRolePKsByBusinessIDs(ctx context.Context, roleIDs []int64) ([]model.MstRole, error)
	EmailExistsActiveGlobally(ctx context.Context, email string) (bool, error)
	EmailExistsActiveInOtherInstitution(ctx context.Context, institutionID int64, email string) (bool, error)
	CountActiveStaffWithRole(ctx context.Context, institutionID int64, roleName string) (int64, error)
}
