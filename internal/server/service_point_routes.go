package server

import (
	"net/http"

	permconst "github.com/faisalhardin/medilink/internal/entity/constant/permission"
	authmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-chi/chi/v5"
)

type servicePointHandlers struct {
	List    http.HandlerFunc
	Get     http.HandlerFunc
	Create  http.HandlerFunc
	Patch   http.HandlerFunc
	Archive http.HandlerFunc
}

func registerServicePointRoutes(r chi.Router, h servicePointHandlers, mw *authmodule.Module) {
	r.With(mw.RequirePermissionStrictWhen(permconst.ServicePointManage, authmodule.WantsArchivedServicePointList)).
		Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.With(mw.RequirePermissionStrict(permconst.ServicePointManage)).
		Post("/", h.Create)
	r.With(mw.RequirePermissionStrict(permconst.ServicePointManage)).
		Patch("/{id}", h.Patch)
	r.With(mw.RequirePermissionStrict(permconst.ServicePointManage)).
		Delete("/{id}", h.Archive)
}
