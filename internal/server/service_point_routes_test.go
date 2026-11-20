package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	permconst "github.com/faisalhardin/medilink/internal/entity/constant/permission"
	roleconst "github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
	authmodule "github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/go-chi/chi/v5"
)

type grantPerms struct {
	grant bool
}

func (p *grantPerms) HasStaffPermission(context.Context, int64, int64, string) (bool, error) {
	return p.grant, nil
}

func TestServicePointRoutesRequireLiveGrant(t *testing.T) {
	payloads := []model.UserJWTPayload{
		{
			UserID:      4,
			Roles:       []model.UserRoleJWTDetail{{Name: roleconst.Administrator}},
			Permissions: []string{permconst.ServicePointManage},
		},
		{
			UserID:      5,
			Roles:       []model.UserRoleJWTDetail{{Name: "clerk"}},
			Permissions: []string{permconst.ServicePointManage},
		},
	}

	for _, user := range payloads {
		hit := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hit++
			w.WriteHeader(http.StatusOK)
		})
		mw := authmodule.NewMiddlewareModule(&authmodule.Module{Permissions: &grantPerms{grant: false}})
		router := chi.NewRouter()
		router.Route("/v1/journey/service-point", func(r chi.Router) {
			registerServicePointRoutes(r, servicePointHandlers{
				List: handler, Get: handler, Create: handler, Patch: handler, Archive: handler,
			}, mw)
		})

		denied := []struct {
			method string
			path   string
		}{
			{http.MethodPost, "/v1/journey/service-point"},
			{http.MethodPatch, "/v1/journey/service-point/9"},
			{http.MethodDelete, "/v1/journey/service-point/9"},
		}
		for _, call := range denied {
			hit = 0
			rec := requestAs(router, call.method, call.path, user)
			if rec.Code != http.StatusForbidden || hit != 0 || !strings.Contains(rec.Body.String(), "insufficient_permissions") {
				t.Fatalf("%s %s user %d: status %d hit %d body %s", call.method, call.path, user.UserID, rec.Code, hit, rec.Body.Bytes())
			}
		}

		for _, path := range []string{"/v1/journey/service-point", "/v1/journey/service-point/9"} {
			hit = 0
			rec := requestAs(router, http.MethodGet, path, user)
			if rec.Code != http.StatusOK || hit != 1 {
				t.Fatalf("GET %s user %d: status %d hit %d body %s", path, user.UserID, rec.Code, hit, rec.Body.Bytes())
			}
		}
	}
}

func requestAs(router http.Handler, method, path string, user model.UserJWTPayload) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req = req.WithContext(authmodule.SetUserDetailToCtx(req.Context(), user))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}
