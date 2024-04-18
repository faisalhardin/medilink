package auth

import (
	"context"
	"net/http"

	"github.com/faisalhardin/auth-vessel/internal/config"
	authrepo "github.com/faisalhardin/auth-vessel/internal/entitiy/repo/auth"
	commonwriter "github.com/faisalhardin/auth-vessel/internal/library/common/writer"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	Cfg      *config.Config
	AuthRepo authrepo.Auth
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
