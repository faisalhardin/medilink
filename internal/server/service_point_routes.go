package server

import (
	"net/http"

	permconst "github.com/faisalhardin/medilink/internal/entity/constant/permission"
	authmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-chi/chi/v5"
)

type servicePointHandlers struct {
	List             http.HandlerFunc
	Get              http.HandlerFunc
	Create           http.HandlerFunc
	Patch            http.HandlerFunc
	Archive          http.HandlerFunc
	GetJourneyPoints http.HandlerFunc
	PutJourneyPoints http.HandlerFunc
}

type journeyPointRoomHandlers struct {
	Get http.HandlerFunc
	Put http.HandlerFunc
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
	r.With(mw.RequirePermissionStrict(permconst.ServicePointManage)).
		Get("/{id}/journey-points", h.GetJourneyPoints)
	r.With(mw.RequirePermissionStrict(permconst.ServicePointManage)).
		Put("/{id}/journey-points", h.PutJourneyPoints)
}

func registerJourneyPointRoomRoutes(r chi.Router, h journeyPointRoomHandlers, mw *authmodule.Module) {
	r.With(mw.RequirePermissionStrict(permconst.ServicePointManage)).
		Put("/{id}/service-points", h.Put)
	r.Get("/{id}/service-points", h.Get)
}
