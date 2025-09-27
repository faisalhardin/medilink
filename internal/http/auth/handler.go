package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
	authrepo "github.com/faisalhardin/medilink/internal/entity/repo/auth"
	authuc "github.com/faisalhardin/medilink/internal/entity/usecase/auth"
	"github.com/faisalhardin/medilink/internal/entity/user"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	"github.com/faisalhardin/medilink/internal/repo/auth"
	userrepo "github.com/faisalhardin/medilink/internal/repo/staff"
	authmodel "github.com/faisalhardin/medilink/internal/usecase/auth"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
	providerKey = "provider"
)

type AuthHandler struct {
	Cfg         *config.Config
	AuthRepo    authrepo.AuthRepo
	SessionRepo *auth.SessionRepository
	UserRepo    userrepo.Conn
	AuthUC      authuc.AuthUC
}

func New(handler *AuthHandler) *AuthHandler {
	return handler
}

func (h *AuthHandler) GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authParams, err := h.AuthUC.Login(w, r, authmodel.AuthParams{})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, authParams)
}

func (h *AuthHandler) GetTokenFromTokenKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tokenCookie, err := r.Cookie(h.Cfg.AuthSessionConfig.SessionKey)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.AuthUC.GetToken(ctx, authmodel.AuthParams{TokenKey: tokenCookie.Value})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

func (h *AuthHandler) BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	provider := chi.URLParam(r, providerKey)
	r = r.WithContext(context.WithValue(ctx, providerKey, provider))
	h.AuthRepo.BeginAuthProviderCallback(w, r)

}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.AuthRepo.Logout(w, r)
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

func (h *AuthHandler) PseudoLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(ctx, providerKey, provider))

	// user, err := h.AuthRepo.GetAuthCallbackFunction(w, r)

	requestData := authmodel.AuthParams{}

	res, err := h.AuthUC.Login(w, r, requestData)
	if err != nil {
		fmt.Println(err)
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, res)
}

func (h *AuthHandler) GetLoginByToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData := authmodel.AuthParams{}
	err := bindingBind(r, &requestData)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	userDetail, err := h.AuthUC.GetUserDetail(ctx, requestData)
	if err != nil {
		fmt.Println(err)
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, userDetail)
}

func (h *AuthHandler) GetUserFromToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonwriter.SetOKWithData(ctx, w, "userClaim")
}

func (h *AuthHandler) PingAPI(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()
	commonwriter.SetOKWithData(ctx, w, "OK")
}

// RefreshToken handles refresh token requests
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData := model.RefreshTokenRequest{}
	err := bindingBind(r, &requestData)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	resp, err := h.AuthUC.RefreshToken(ctx, requestData)
	if err != nil && errors.Is(err, commonerr.SetNewRevokedSessionError()) {
		w.Header().Set("X-Redirect-To", "/token-expired")
		commonwriter.SetError(ctx, w, err)
		return
	} else if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, resp)
}

// LogoutSession handles single session logout
func (h *AuthHandler) LogoutSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestData := model.LogoutRequest{}
	err := bindingBind(r, &requestData)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	err = h.AuthUC.Logout(ctx, r)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, map[string]string{"message": "Logged out successfully"})
}

// LogoutAllSessions handles logout from all sessions
func (h *AuthHandler) LogoutAllSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT token or session
	// This is a simplified approach - in production, you'd extract this from the JWT
	userID := int64(1) // This should be extracted from the authenticated user's context

	err := h.AuthUC.LogoutAllSessions(ctx, userID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, map[string]string{"message": "Logged out from all sessions successfully"})
}

// GetUserSessions retrieves all sessions for the current user
func (h *AuthHandler) GetUserSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user ID from JWT token or session
	// This is a simplified approach - in production, you'd extract this from the JWT
	userID := int64(1) // This should be extracted from the authenticated user's context

	sessions, err := h.AuthUC.GetUserSessions(ctx, userID)
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	commonwriter.SetOKWithData(ctx, w, sessions)
}
