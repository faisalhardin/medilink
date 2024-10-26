package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	authrepo "github.com/faisalhardin/medilink/internal/entity/repo/auth"
	authuc "github.com/faisalhardin/medilink/internal/entity/usecase/auth"
	"github.com/faisalhardin/medilink/internal/entity/user"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	"github.com/faisalhardin/medilink/internal/library/util/common/binding"
	userrepo "github.com/faisalhardin/medilink/internal/repo/staff"
	authmodel "github.com/faisalhardin/medilink/internal/usecase/auth"
	"github.com/go-chi/chi/v5"
)

var (
	bindingBind = binding.Bind
	providerKey = "provider"
)

type AuthHandler struct {
	Cfg      *config.Config
	AuthRepo authrepo.AuthRepo
	UserRepo userrepo.Conn
	AuthUC   authuc.AuthUC
}

func New(handler *AuthHandler) *AuthHandler {
	return handler
}

func (h *AuthHandler) GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(ctx, providerKey, provider))

	authParams, err := h.AuthUC.Login(w, r, authmodel.AuthParams{})
	if err != nil {
		commonwriter.SetError(ctx, w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     h.Cfg.AuthSessionConfig.SessionKey,
		Value:    authParams.Token,
		Expires:  time.Now().Add(time.Duration(h.Cfg.JWTConfig.DurationInMinutes) * time.Minute),
		HttpOnly: true, // Securely store the cookie
		// Secure:   true, // Only send over HTTPS
		Path:   h.Cfg.AuthSessionConfig.Path,
		Domain: h.Cfg.AuthSessionConfig.Domain,
	})

	http.Redirect(w, r, h.Cfg.WebConfig.Host, http.StatusFound)
}

func (h *AuthHandler) BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	// try to get the user without re-authenticating
	provider := chi.URLParam(r, providerKey)
	r = r.WithContext(context.WithValue(ctx, providerKey, provider))
	h.AuthRepo.BeginAuthProviderCallback(w, r)

	// commonwriter.Redirect(ctx, w, r, h.Cfg.GoogleAuthConfig.HomepageRedirect, http.StatusFound)
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
