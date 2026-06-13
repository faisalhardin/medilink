package staff

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type StaffUC interface {
	ListStaff(ctx context.Context, includeInactive bool) (model.ListStaffResponse, error)
	GetStaff(ctx context.Context, uuid string) (model.StaffWithRolesResponse, error)
	CreateStaff(ctx context.Context, req model.CreateStaffRequest) error
	AssignRole(ctx context.Context, req model.AssignRoleRequest) error
	UnassignRole(ctx context.Context, req model.UnassignRoleRequest) error
	DeactivateStaff(ctx context.Context, req model.StaffStatusRequest) error
	ActivateStaff(ctx context.Context, req model.StaffStatusRequest) error
}
