package staff

import (
	"context"
	"net/http"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	staffrepo "github.com/faisalhardin/medilink/internal/entity/repo/staff"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/go-xorm/xorm"
	"github.com/pkg/errors"
)

const (
	wrapStaffDBPrefix              = "StaffDB."
	wrapMsgListStaffByInstitution  = wrapStaffDBPrefix + "ListStaffByInstitution"
	wrapMsgGetStaffByUUID          = wrapStaffDBPrefix + "GetStaffByUUID"
	wrapMsgGetStaffIDByUUID        = wrapStaffDBPrefix + "GetStaffIDByUUID"
	wrapMsgInsertStaff             = wrapStaffDBPrefix + "InsertStaff"
	wrapMsgDeactivateStaff         = wrapStaffDBPrefix + "DeactivateStaff"
	wrapMsgActivateStaff           = wrapStaffDBPrefix + "ActivateStaff"
	wrapMsgAssignRole              = wrapStaffDBPrefix + "AssignRole"
	wrapMsgUnassignRole            = wrapStaffDBPrefix + "UnassignRole"
	wrapMsgHasRoleAssignment       = wrapStaffDBPrefix + "HasRoleAssignment"
	wrapMsgGetRolePKByBusinessID    = wrapStaffDBPrefix + "GetRolePKByBusinessID"
	wrapMsgGetRolePKsByBusinessIDs = wrapStaffDBPrefix + "GetRolePKsByBusinessIDs"
	wrapMsgEmailExistsActiveGlobally  = wrapStaffDBPrefix + "EmailExistsActiveGlobally"
	wrapMsgEmailExistsInOtherInst     = wrapStaffDBPrefix + "EmailExistsActiveInOtherInstitution"
	wrapMsgCountActiveStaffWithRole = wrapStaffDBPrefix + "CountActiveStaffWithRole"
)

type staffListRow struct {
	UUID            string         `xorm:"uuid"`
	Name            string         `xorm:"name"`
	Email           string         `xorm:"email"`
	InstitutionID   int64          `xorm:"id_mst_institution"`
	InstitutionName string         `xorm:"institution_name"`
	CreateTime      time.Time      `xorm:"create_time"`
	UpdateTime      time.Time      `xorm:"update_time"`
	DeleteTime      *time.Time     `xorm:"delete_time"`
	Roles           []model.MstRole `xorm:"roles"`
}

func NewStaffDB(conn *Conn) staffrepo.StaffDB {
	return conn
}

func (c *Conn) ListStaffByInstitution(ctx context.Context, institutionID int64, includeInactive bool) ([]model.StaffWithRolesResponse, error) {
	const sql = `
		SELECT
			s.uuid,
			s.name,
			s.email,
			s.id_mst_institution,
			i.name AS institution_name,
			s.create_time,
			s.update_time,
			s.delete_time,
			COALESCE(
				jsonb_agg(
					jsonb_build_object('role_id', r.role_id, 'name', r.name)
				) FILTER (WHERE r.id IS NOT NULL),
				'[]'
			) AS roles
		FROM mdl_mst_staff s
		JOIN mdl_mst_institution i ON i.id = s.id_mst_institution AND i.delete_time IS NULL
		LEFT JOIN mdl_map_role_staff mrs ON mrs.id_mst_staff = s.id
		LEFT JOIN mdl_mst_role r ON r.id = mrs.id_mst_role AND r.delete_time IS NULL
		WHERE s.id_mst_institution = ?
		  AND (? OR s.delete_time IS NULL)
		GROUP BY s.id, i.name
		ORDER BY s.create_time DESC
	`

	var rows []staffListRow
	err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, includeInactive).Find(&rows)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgListStaffByInstitution)
	}

	return mapStaffListRows(rows), nil
}

func (c *Conn) GetStaffByUUID(ctx context.Context, institutionID int64, uuid string, includeInactive bool) (model.StaffWithRolesResponse, error) {
	const sql = `
		SELECT
			s.uuid,
			s.name,
			s.email,
			s.id_mst_institution,
			i.name AS institution_name,
			s.create_time,
			s.update_time,
			s.delete_time,
			COALESCE(
				jsonb_agg(
					jsonb_build_object('role_id', r.role_id, 'name', r.name)
				) FILTER (WHERE r.id IS NOT NULL),
				'[]'
			) AS roles
		FROM mdl_mst_staff s
		JOIN mdl_mst_institution i ON i.id = s.id_mst_institution AND i.delete_time IS NULL
		LEFT JOIN mdl_map_role_staff mrs ON mrs.id_mst_staff = s.id
		LEFT JOIN mdl_mst_role r ON r.id = mrs.id_mst_role AND r.delete_time IS NULL
		WHERE s.id_mst_institution = ?
		  AND s.uuid = ?
		  AND (? OR s.delete_time IS NULL)
		GROUP BY s.id, i.name
	`

	var row staffListRow
	found, err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, uuid, includeInactive).Get(&row)
	if err != nil {
		return model.StaffWithRolesResponse{}, errors.Wrap(err, wrapMsgGetStaffByUUID)
	}
	if !found {
		return model.StaffWithRolesResponse{}, commonerr.SetNewError(http.StatusNotFound, "staff_not_found", "staff was not found in this institution")
	}

	mapped := mapStaffListRows([]staffListRow{row})
	return mapped[0], nil
}

func (c *Conn) GetStaffIDByUUID(ctx context.Context, institutionID int64, uuid string) (int64, error) {
	var staff model.MstStaff
	found, err := c.DB.SlaveDB.Context(ctx).
		Table(model.MST_STAFF_TABLE).
		Unscoped().
		Where("uuid = ? AND id_mst_institution = ?", uuid, institutionID).
		Cols("id").
		Get(&staff)
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgGetStaffIDByUUID)
	}
	if !found {
		return 0, commonerr.SetNewError(http.StatusNotFound, "staff_not_found", "staff was not found in this institution")
	}
	return staff.ID, nil
}

func (c *Conn) InsertStaff(ctx context.Context, staff *model.MstStaff) error {
	_, err := c.writeSession(ctx).
		Table(model.MST_STAFF_TABLE).
		InsertOne(staff)
	if err != nil {
		return errors.Wrap(err, wrapMsgInsertStaff)
	}
	return nil
}

func (c *Conn) DeactivateStaff(ctx context.Context, institutionID int64, uuid string) error {
	now := time.Now()
	affected, err := c.writeSession(ctx).
		Table(model.MST_STAFF_TABLE).
		Where("uuid = ? AND id_mst_institution = ? AND delete_time IS NULL", uuid, institutionID).
		Cols("delete_time", "update_time").
		Update(&model.MstStaff{DeleteTime: &now})
	if err != nil {
		return errors.Wrap(err, wrapMsgDeactivateStaff)
	}
	if affected == 0 {
		return commonerr.SetNewError(http.StatusNotFound, "staff_not_found", "staff was not found or is already inactive")
	}
	return nil
}

func (c *Conn) ActivateStaff(ctx context.Context, institutionID int64, uuid string) error {
	affected, err := c.writeSession(ctx).
		Table(model.MST_STAFF_TABLE).
		Unscoped().
		Where("uuid = ? AND id_mst_institution = ? AND delete_time IS NOT NULL", uuid, institutionID).
		Cols("delete_time", "update_time").
		Update(map[string]interface{}{"delete_time": nil})
	if err != nil {
		return errors.Wrap(err, wrapMsgActivateStaff)
	}
	if affected == 0 {
		return commonerr.SetNewError(http.StatusNotFound, "staff_not_found", "staff was not found or is already active")
	}
	return nil
}

func (c *Conn) AssignRole(ctx context.Context, staffID int64, rolePK int64) error {
	const sql = `
		INSERT INTO mdl_map_role_staff (id_mst_staff, id_mst_role)
		SELECT ?, ?
		WHERE NOT EXISTS (
			SELECT 1 FROM mdl_map_role_staff
			WHERE id_mst_staff = ? AND id_mst_role = ?
		)
	`
	_, err := c.writeSession(ctx).Exec(sql, staffID, rolePK, staffID, rolePK)
	if err != nil {
		return errors.Wrap(err, wrapMsgAssignRole)
	}
	return nil
}

func (c *Conn) UnassignRole(ctx context.Context, staffID int64, rolePK int64) error {
	_, err := c.writeSession(ctx).
		Table(model.MAP_ROLE_STAFF).
		Where("id_mst_staff = ? AND id_mst_role = ?", staffID, rolePK).
		Delete(&model.RoleStaffMapping{})
	if err != nil {
		return errors.Wrap(err, wrapMsgUnassignRole)
	}
	return nil
}

func (c *Conn) HasRoleAssignment(ctx context.Context, staffID int64, rolePK int64) (bool, error) {
	found, err := c.DB.SlaveDB.Context(ctx).
		Table(model.MAP_ROLE_STAFF).
		Where("id_mst_staff = ? AND id_mst_role = ?", staffID, rolePK).
		Exist(&model.RoleStaffMapping{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgHasRoleAssignment)
	}
	return found, nil
}

func (c *Conn) GetRolePKByBusinessID(ctx context.Context, roleID int64) (model.MstRole, bool, error) {
	var role model.MstRole
	found, err := c.DB.SlaveDB.Context(ctx).
		Table(model.MST_ROLE_TABLE).
		Where("role_id = ? AND delete_time IS NULL", roleID).
		Get(&role)
	if err != nil {
		return model.MstRole{}, false, errors.Wrap(err, wrapMsgGetRolePKByBusinessID)
	}
	return role, found, nil
}

func (c *Conn) GetRolePKsByBusinessIDs(ctx context.Context, roleIDs []int64) ([]model.MstRole, error) {
	if len(roleIDs) == 0 {
		return []model.MstRole{}, nil
	}

	roles := make([]model.MstRole, 0, len(roleIDs))
	err := c.DB.SlaveDB.Context(ctx).
		Table(model.MST_ROLE_TABLE).
		In("role_id", roleIDs).
		Where("delete_time IS NULL").
		Find(&roles)
	if err != nil {
		return nil, errors.Wrap(err, wrapMsgGetRolePKsByBusinessIDs)
	}
	return roles, nil
}

func (c *Conn) EmailExistsActiveGlobally(ctx context.Context, email string) (bool, error) {
	found, err := c.DB.SlaveDB.Context(ctx).
		Table(model.MST_STAFF_TABLE).
		Where("LOWER(email) = LOWER(?) AND delete_time IS NULL", email).
		Exist(&model.MstStaff{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgEmailExistsActiveGlobally)
	}
	return found, nil
}

func (c *Conn) EmailExistsActiveInOtherInstitution(ctx context.Context, institutionID int64, email string) (bool, error) {
	found, err := c.DB.SlaveDB.Context(ctx).
		Table(model.MST_STAFF_TABLE).
		Where("LOWER(email) = LOWER(?) AND delete_time IS NULL AND id_mst_institution <> ?", email, institutionID).
		Exist(&model.MstStaff{})
	if err != nil {
		return false, errors.Wrap(err, wrapMsgEmailExistsInOtherInst)
	}
	return found, nil
}

func (c *Conn) CountActiveStaffWithRole(ctx context.Context, institutionID int64, roleName string) (int64, error) {
	const sql = `
		SELECT COUNT(DISTINCT s.id)
		FROM mdl_mst_staff s
		JOIN mdl_map_role_staff mrs ON mrs.id_mst_staff = s.id
		JOIN mdl_mst_role r ON r.id = mrs.id_mst_role AND r.delete_time IS NULL
		WHERE s.id_mst_institution = ?
		  AND s.delete_time IS NULL
		  AND r.name = ?
	`
	count, err := c.DB.SlaveDB.Context(ctx).SQL(sql, institutionID, roleName).Count()
	if err != nil {
		return 0, errors.Wrap(err, wrapMsgCountActiveStaffWithRole)
	}
	return count, nil
}

func (c *Conn) writeSession(ctx context.Context) *xorm.Session {
	if session := xormlib.GetDBSession(ctx); session != nil {
		return session
	}
	return c.DB.MasterDB.Context(ctx)
}

func mapStaffListRows(rows []staffListRow) []model.StaffWithRolesResponse {
	result := make([]model.StaffWithRolesResponse, 0, len(rows))
	for _, row := range rows {
		roles := make([]model.StaffRoleResponse, 0, len(row.Roles))
		for _, role := range row.Roles {
			if role.RoleID == 0 && role.Name == "" {
				continue
			}
			roles = append(roles, model.StaffRoleResponse{
				RoleID: role.RoleID,
				Name:   role.Name,
			})
		}
		result = append(result, model.StaffWithRolesResponse{
			UUID:            row.UUID,
			Name:            row.Name,
			Email:           row.Email,
			InstitutionID:   row.InstitutionID,
			InstitutionName: row.InstitutionName,
			Roles:           roles,
			IsActive:        row.DeleteTime == nil,
			CreatedAt:       row.CreateTime,
			UpdatedAt:       row.UpdateTime,
		})
	}
	return result
}
