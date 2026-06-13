package permission

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
	permissionrepo "github.com/faisalhardin/medilink/internal/entity/repo/permission"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/pkg/errors"
)

const (
	WrapErrMsgPrefix            = "PermissionDB."
	WrapMsgGetByRoleNames       = WrapErrMsgPrefix + "GetPermissionsByRoleNames"
)

type Conn struct {
	DB *xormlib.DBConnect
}

func NewPermissionDB(db *xormlib.DBConnect) permissionrepo.PermissionDB {
	return &Conn{DB: db}
}

func (c *Conn) GetPermissionsByRoleNames(ctx context.Context, roleNames []string) ([]model.MstPermission, error) {
	if len(roleNames) == 0 {
		return []model.MstPermission{}, nil
	}

	var permissions []model.MstPermission
	err := c.DB.SlaveDB.Context(ctx).
		Table(model.MST_PERMISSION_TABLE).
		Alias("p").
		Distinct("p.id, p.code, p.resource, p.action, p.description, p.create_time, p.update_time, p.delete_time").
		Join("INNER", model.MAP_ROLE_PERMISSION+" mrp", "mrp.id_mst_permission = p.id").
		Join("INNER", model.MST_ROLE_TABLE+" r", "r.id = mrp.id_mst_role AND r.delete_time IS NULL").
		In("r.name", roleNames).
		Where("p.delete_time IS NULL").
		Asc("p.code").
		Find(&permissions)
	if err != nil {
		return nil, errors.Wrap(err, WrapMsgGetByRoleNames)
	}

	return permissions, nil
}
