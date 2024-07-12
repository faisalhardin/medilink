package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/faisalhardin/medilink/internal/config"
	authrepo "github.com/faisalhardin/medilink/internal/entitiy/repo/auth"
	"github.com/faisalhardin/medilink/internal/entitiy/user"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	userrepo "github.com/faisalhardin/medilink/internal/repo/user"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
)

type AuthHandler struct {
	Cfg      *config.Config
	AuthRepo authrepo.AuthRepo
	UserRepo userrepo.Conn
}

func New(handler *AuthHandler) *AuthHandler {
	return handler
}

func (h *AuthHandler) GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	h.AuthRepo.GetAuthCallbackFunction(w, r)
	http.Redirect(w, r, h.Cfg.GoogleAuthConfig.HomepageRedirect, http.StatusFound)
}

func (h *AuthHandler) BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {

	// try to get the user without re-authenticating
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	h.AuthRepo.BeginAuthProviderCallback(w, r)

	ctx := context.Background()
	commonwriter.Redirect(ctx, w, r, h.Cfg.GoogleAuthConfig.HomepageRedirect, http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.AuthRepo.Logout(w, r)
}

func (h *AuthHandler) PingAPI(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	commonwriter.SetOKWithData(ctx, w, "OK")
}

func (h *AuthHandler) TestAPIRedirect(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	commonwriter.Redirect(ctx, w, r, "https://inspirybox.id", http.StatusTemporaryRedirect)
}

func (h *AuthHandler) TestReturnError(w http.ResponseWriter, r *http.Request) {

}

func (h *AuthHandler) TestBinding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData := user.User{}
	err := bindingBind(r, &requestData)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, requestData)

}

func (h *AuthHandler) TestSchemaBinding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData := user.UserRequest{}
	err := bindingBind(r, &requestData)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, requestData)

}

func (h *AuthHandler) TestGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData := user.User{}
	err := bindingBind(r, &requestData)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	user, _, err := h.UserRepo.GetUserByParams(ctx, requestData)
	if err != nil {
		fmt.Println(err)
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, user)
}
