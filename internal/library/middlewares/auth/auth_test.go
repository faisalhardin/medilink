package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	permconst "github.com/faisalhardin/medilink/internal/entity/constant/permission"
	roleconst "github.com/faisalhardin/medilink/internal/entity/constant/role"
	"github.com/faisalhardin/medilink/internal/entity/model"
)

type fakePerms struct {
	grant   bool
	err     error
	calls   int
	staffID int64
	code    string
}

func (f *fakePerms) HasStaffPermission(_ context.Context, staffID int64, code string) (bool, error) {
	f.calls++
	f.staffID = staffID
	f.code = code
	return f.grant, f.err
}

type flipPerms struct {
	grants []bool
	i      int
}

func (f *flipPerms) HasStaffPermission(context.Context, int64, string) (bool, error) {
	grant := f.grants[f.i]
	f.i++
	return grant, nil
}

func serve(mw func(http.Handler) http.Handler, user *model.UserJWTPayload, target string) *httptest.ResponseRecorder {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if user != nil {
		req = req.WithContext(SetUserDetailToCtx(req.Context(), *user))
	}
	rec := httptest.NewRecorder()
	mw(next).ServeHTTP(rec, req)
	return rec
}

func TestRequirePermissionStrictIgnoresJWTSnapshot(t *testing.T) {
	code := permconst.ServicePointManage
	admin := model.UserJWTPayload{
		UserID:      4,
		Roles:       []model.UserRoleJWTDetail{{Name: roleconst.Administrator}},
		Permissions: []string{code},
	}
	clerk := model.UserJWTPayload{
		UserID:      5,
		Roles:       []model.UserRoleJWTDetail{{Name: "clerk"}},
		Permissions: []string{code},
	}

	tests := []struct {
		name   string
		user   *model.UserJWTPayload
		perms  PermissionChecker
		loose  int
		strict int
		secret string
	}{
		{
			name:   "admin jwt grant revoked in db",
			user:   &admin,
			perms:  &fakePerms{grant: false},
			loose:  http.StatusOK,
			strict: http.StatusForbidden,
		},
		{
			name:   "non-admin jwt grant revoked in db",
			user:   &clerk,
			perms:  &fakePerms{grant: false},
			loose:  http.StatusOK,
			strict: http.StatusForbidden,
		},
		{
			name:   "grant still in db",
			user:   &clerk,
			perms:  &fakePerms{grant: true},
			loose:  http.StatusOK,
			strict: http.StatusOK,
		},
		{
			name:   "no user",
			perms:  &fakePerms{grant: true},
			loose:  http.StatusUnauthorized,
			strict: http.StatusUnauthorized,
		},
		{
			name:   "nil checker",
			user:   &admin,
			loose:  http.StatusOK,
			strict: http.StatusInternalServerError,
		},
		{
			name:   "checker error",
			user:   &admin,
			perms:  &fakePerms{err: errors.New("postgres secret")},
			loose:  http.StatusOK,
			strict: http.StatusInternalServerError,
			secret: "postgres secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &Module{Permissions: tt.perms}
			loose := serve(mod.RequirePermission(code), tt.user, "/v1/journey/service-point")
			strict := serve(mod.RequirePermissionStrict(code), tt.user, "/v1/journey/service-point")
			if loose.Code != tt.loose || strict.Code != tt.strict {
				t.Fatalf("loose %d want %d, strict %d want %d\nloose body %s\nstrict body %s",
					loose.Code, tt.loose, strict.Code, tt.strict, loose.Body.Bytes(), strict.Body.Bytes())
			}
			if tt.strict == http.StatusForbidden && !strings.Contains(strict.Body.String(), "insufficient_permissions") {
				t.Fatalf("strict body %s", strict.Body.Bytes())
			}
			if tt.secret != "" && strings.Contains(strict.Body.String(), tt.secret) {
				t.Fatalf("strict body leaked checker error: %s", strict.Body.Bytes())
			}
			if fake, ok := tt.perms.(*fakePerms); ok && tt.user != nil && tt.strict != http.StatusUnauthorized {
				if fake.staffID != tt.user.UserID || fake.code != code {
					t.Fatalf("checker saw staff %d code %q, want %d %q", fake.staffID, fake.code, tt.user.UserID, code)
				}
			}
			if tt.name == "no user" {
				if fake, ok := tt.perms.(*fakePerms); ok && fake.calls != 0 {
					t.Fatalf("checker calls %d, want 0", fake.calls)
				}
			}
		})
	}
}

func TestRequirePermissionStrictSeesRevokeInTheSameSession(t *testing.T) {
	code := permconst.ServicePointManage
	user := model.UserJWTPayload{
		UserID:      4,
		Roles:       []model.UserRoleJWTDetail{{Name: roleconst.Administrator}},
		Permissions: []string{code},
	}
	checker := &flipPerms{grants: []bool{true, false}}
	mod := &Module{Permissions: checker}
	mw := mod.RequirePermissionStrict(code)

	first := serve(mw, &user, "/v1/journey/service-point")
	second := serve(mw, &user, "/v1/journey/service-point")
	if first.Code != http.StatusOK {
		t.Fatalf("first %d %s", first.Code, first.Body.Bytes())
	}
	if second.Code != http.StatusForbidden || !strings.Contains(second.Body.String(), "insufficient_permissions") {
		t.Fatalf("second %d %s", second.Code, second.Body.Bytes())
	}
}

func TestRequirePermissionStrictWhenIncludeArchived(t *testing.T) {
	code := permconst.ServicePointManage
	user := model.UserJWTPayload{UserID: 8, Permissions: []string{code}}

	tests := []struct {
		name   string
		query  string
		grant  bool
		status int
		calls  int
	}{
		{name: "absent", status: http.StatusOK},
		{name: "false", query: "include_archived=false", status: http.StatusOK},
		{name: "true", query: "include_archived=true", status: http.StatusForbidden, calls: 1},
		{name: "TRUE", query: "include_archived=TRUE", status: http.StatusForbidden, calls: 1},
		{name: "on", query: "include_archived=on", status: http.StatusForbidden, calls: 1},
		{name: "one", query: "include_archived=1", status: http.StatusForbidden, calls: 1},
		{name: "repeated", query: "include_archived=false&include_archived=true", status: http.StatusForbidden, calls: 1},
		{name: "upper key", query: "INCLUDE_ARCHIVED=true", status: http.StatusForbidden, calls: 1},
		{name: "unparseable", query: "include_archived=maybe", status: http.StatusForbidden, calls: 1},
		{name: "granted", query: "include_archived=true", grant: true, status: http.StatusOK, calls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakePerms{grant: tt.grant}
			mod := &Module{Permissions: checker}
			target := "/v1/journey/service-point"
			if tt.query != "" {
				target += "?" + tt.query
			}
			rec := serve(mod.RequirePermissionStrictWhen(code, WantsArchivedServicePointList), &user, target)
			if rec.Code != tt.status {
				t.Fatalf("status %d body %s, want %d", rec.Code, rec.Body.Bytes(), tt.status)
			}
			if checker.calls != tt.calls {
				t.Fatalf("checker calls %d, want %d", checker.calls, tt.calls)
			}
			if tt.status == http.StatusForbidden && !strings.Contains(rec.Body.String(), "insufficient_permissions") {
				t.Fatalf("body %s", rec.Body.Bytes())
			}
		})
	}
}
