package staff

import (
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	staffuc "github.com/faisalhardin/medilink/internal/entity/usecase/staff"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
)

type StaffHandler struct {
	StaffUC staffuc.StaffUC
}

func New(handler *StaffHandler) *StaffHandler {
	return handler
}

func (h *StaffHandler) ListStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params := model.ListStaffParams{}
	err := bindingBind(r, &params)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	response, err := h.StaffUC.ListStaff(ctx, params.IncludeInactive)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *StaffHandler) GetStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "uuid")

	response, err := h.StaffUC.GetStaff(ctx, uuid)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, response)
}

func (h *StaffHandler) CreateStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.CreateStaffRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.StaffUC.CreateStaff(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, "Staff created successfully")
}

func (h *StaffHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.AssignRoleRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.StaffUC.AssignRole(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, "Role assigned successfully")
}

func (h *StaffHandler) UnassignRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request := model.UnassignRoleRequest{}
	err := bindingBind(r, &request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.StaffUC.UnassignRole(ctx, request)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, "Role unassigned successfully")
}

func (h *StaffHandler) DeactivateStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "uuid")

	err := h.StaffUC.DeactivateStaff(ctx, model.StaffStatusRequest{UUID: uuid})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, "Staff deactivated successfully")
}

func (h *StaffHandler) ActivateStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "uuid")

	err := h.StaffUC.ActivateStaff(ctx, model.StaffStatusRequest{UUID: uuid})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}
	commonwriter.SetOKWithData(ctx, w, "Staff activated successfully")
}
