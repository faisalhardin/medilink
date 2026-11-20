package permission

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

type PermissionDB interface {
	GetPermissionsByRoleNames(ctx context.Context, roleNames []string) ([]model.MstPermission, error)
	HasStaffPermission(ctx context.Context, staffID, clinicID int64, code string) (bool, error)
}
