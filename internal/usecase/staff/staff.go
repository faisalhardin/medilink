package staff

import (
	"context"
	"net/http"
	"strings"

	"github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	staffrepo "github.com/faisalhardin/medilink/internal/entity/repo/staff"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	xormlib "github.com/faisalhardin/medilink/internal/library/db/xorm"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	wrapStaffUCPrefix     = "StaffUC."
	wrapMsgListStaff      = wrapStaffUCPrefix + "ListStaff"
	wrapMsgGetStaff       = wrapStaffUCPrefix + "GetStaff"
	wrapMsgCreateStaff    = wrapStaffUCPrefix + "CreateStaff"
	wrapMsgAssignRole     = wrapStaffUCPrefix + "AssignRole"
	wrapMsgUnassignRole   = wrapStaffUCPrefix + "UnassignRole"
	wrapMsgDeactivateStaff = wrapStaffUCPrefix + "DeactivateStaff"
	wrapMsgActivateStaff  = wrapStaffUCPrefix + "ActivateStaff"
)

type StaffUC struct {
	StaffDB     staffrepo.StaffDB
	Transaction xormlib.DBTransactionInterface
}

func NewStaffUC(uc *StaffUC) *StaffUC {
	return uc
}

func (u *StaffUC) ListStaff(ctx context.Context, includeInactive bool) (model.ListStaffResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.ListStaffResponse{}, err
	}

	staff, err := u.StaffDB.ListStaffByInstitution(ctx, userDetail.InstitutionID, includeInactive)
	if err != nil {
		return model.ListStaffResponse{}, errors.Wrap(err, wrapMsgListStaff)
	}

	return model.ListStaffResponse{
		Staff: staff,
		Total: len(staff),
	}, nil
}

func (u *StaffUC) GetStaff(ctx context.Context, uuid string) (model.StaffWithRolesResponse, error) {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return model.StaffWithRolesResponse{}, err
	}

	staff, err := u.StaffDB.GetStaffByUUID(ctx, userDetail.InstitutionID, uuid, true)
	if err != nil {
		return model.StaffWithRolesResponse{}, errors.Wrap(err, wrapMsgGetStaff)
	}
	return staff, nil
}

func (u *StaffUC) CreateStaff(ctx context.Context, req model.CreateStaffRequest) error {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	exists, err := u.StaffDB.EmailExistsActiveGlobally(ctx, normalizedEmail)
	if err != nil {
		return errors.Wrap(err, wrapMsgCreateStaff)
	}
	if exists {
		return commonerr.SetNewBadRequest("email_exists", "a staff member with this email already exists")
	}

	rolePKs, err := u.resolveRolePKs(ctx, req.RoleIDs)
	if err != nil {
		return errors.Wrap(err, wrapMsgCreateStaff)
	}

	session, err := u.Transaction.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, wrapMsgCreateStaff)
	}
	defer u.Transaction.Finish(session, &err)
	ctx = xormlib.SetDBSession(ctx, session)

	staff := &model.MstStaff{
		Name:             req.Name,
		Email:            normalizedEmail,
		IdMstInstitution: userDetail.InstitutionID,
	}
	if err = u.StaffDB.InsertStaff(ctx, staff); err != nil {
		return errors.Wrap(err, wrapMsgCreateStaff)
	}

	for _, rolePK := range rolePKs {
		if err = u.StaffDB.AssignRole(ctx, staff.ID, rolePK); err != nil {
			return errors.Wrap(err, wrapMsgCreateStaff)
		}
	}

	return nil
}

func (u *StaffUC) AssignRole(ctx context.Context, req model.AssignRoleRequest) error {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return err
	}

	staffID, rolePK, _, err := u.resolveStaffAndRoleWithMeta(ctx, userDetail.InstitutionID, req.StaffUUID, req.RoleID)
	if err != nil {
		return errors.Wrap(err, wrapMsgAssignRole)
	}

	if err = u.StaffDB.AssignRole(ctx, staffID, rolePK); err != nil {
		return errors.Wrap(err, wrapMsgAssignRole)
	}
	return nil
}

func (u *StaffUC) UnassignRole(ctx context.Context, req model.UnassignRoleRequest) error {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return err
	}

	staffID, rolePK, roleMeta, err := u.resolveStaffAndRoleWithMeta(ctx, userDetail.InstitutionID, req.StaffUUID, req.RoleID)
	if err != nil {
		return errors.Wrap(err, wrapMsgUnassignRole)
	}

	hasRole, err := u.StaffDB.HasRoleAssignment(ctx, staffID, rolePK)
	if err != nil {
		return errors.Wrap(err, wrapMsgUnassignRole)
	}
	if !hasRole {
		return nil
	}

	if roleMeta.Name == role.Administrator {
		count, countErr := u.StaffDB.CountActiveStaffWithRole(ctx, userDetail.InstitutionID, role.Administrator)
		if countErr != nil {
			return errors.Wrap(countErr, wrapMsgUnassignRole)
		}
		if count <= 1 {
			return commonerr.SetNewBadRequest("last_administrator", "cannot remove the last administrator role in this institution")
		}
	}

	if err = u.StaffDB.UnassignRole(ctx, staffID, rolePK); err != nil {
		return errors.Wrap(err, wrapMsgUnassignRole)
	}
	return nil
}

func (u *StaffUC) DeactivateStaff(ctx context.Context, req model.StaffStatusRequest) error {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return err
	}

	if req.UUID == userDetail.UUID {
		return commonerr.SetNewBadRequest("cannot_deactivate_self", "you cannot deactivate your own account")
	}

	if err = u.StaffDB.DeactivateStaff(ctx, userDetail.InstitutionID, req.UUID); err != nil {
		return errors.Wrap(err, wrapMsgDeactivateStaff)
	}
	return nil
}

func (u *StaffUC) ActivateStaff(ctx context.Context, req model.StaffStatusRequest) error {
	userDetail, err := u.requireUser(ctx)
	if err != nil {
		return err
	}

	staff, err := u.StaffDB.GetStaffByUUID(ctx, userDetail.InstitutionID, req.UUID, true)
	if err != nil {
		return errors.Wrap(err, wrapMsgActivateStaff)
	}

	existsInOtherInst, err := u.StaffDB.EmailExistsActiveInOtherInstitution(ctx, userDetail.InstitutionID, staff.Email)
	if err != nil {
		return errors.Wrap(err, wrapMsgActivateStaff)
	}
	if existsInOtherInst {
		return commonerr.SetNewBadRequest("email_exists_other_institution", "an active staff with this email already exists in another institution")
	}

	if err = u.StaffDB.ActivateStaff(ctx, userDetail.InstitutionID, req.UUID); err != nil {
		return errors.Wrap(err, wrapMsgActivateStaff)
	}
	return nil
}

func (u *StaffUC) requireUser(ctx context.Context) (model.UserJWTPayload, error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		return model.UserJWTPayload{}, commonerr.SetNewUnauthorizedAPICall()
	}
	return userDetail, nil
}

func (u *StaffUC) resolveRolePKs(ctx context.Context, roleIDs []int64) ([]int64, error) {
	roles, err := u.StaffDB.GetRolePKsByBusinessIDs(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	roleByBusinessID := make(map[int64]int64, len(roles))
	for _, roleMeta := range roles {
		roleByBusinessID[roleMeta.RoleID] = roleMeta.ID
	}

	rolePKs := make([]int64, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		rolePK, found := roleByBusinessID[roleID]
		if !found {
			return nil, commonerr.SetNewBadRequest("role_not_found", "one or more roles were not found")
		}
		rolePKs = append(rolePKs, rolePK)
	}

	return rolePKs, nil
}

func (u *StaffUC) resolveStaffAndRoleWithMeta(ctx context.Context, institutionID int64, staffUUID string, roleID int64) (int64, int64, model.MstRole, error) {
	staffID, err := u.StaffDB.GetStaffIDByUUID(ctx, institutionID, staffUUID)
	if err != nil {
		return 0, 0, model.MstRole{}, err
	}

	roleMeta, found, err := u.StaffDB.GetRolePKByBusinessID(ctx, roleID)
	if err != nil {
		return 0, 0, model.MstRole{}, err
	}
	if !found {
		return 0, 0, model.MstRole{}, commonerr.SetNewError(http.StatusNotFound, "role_not_found", "role was not found")
	}

	return staffID, roleMeta.ID, roleMeta, nil
}
