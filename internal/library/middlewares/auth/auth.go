package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
	authuc "github.com/faisalhardin/medilink/internal/entity/usecase/auth"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	authusecase "github.com/faisalhardin/medilink/internal/usecase/auth"
)

const (
	ContentLength         = "Content-Length"
	ContentType           = "Content-Type"
	Authorization         = "Authorization"
	AccountsAuthorization = "accounts-authorization"
	Bearer                = "Bearer %s"
	Key                   = "key=%s"
	Basic                 = "Basic %s"
	XAppKey               = "X-App-Key"
)

var AllowedHeaders = []string{
	"Accept",
	ContentType,
	ContentLength,
	"Authorization",
	"Accept-Encoding",
	"accounts-authorization",
	"X-CSRF-Token",
	"API-KEY",
	"X-Device",
	"X-Element-ID",
	"x-requested-with",
	XAppKey,
}

var AllowedMethodRequest = []string{
	"OPTIONS",
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"PATCH",
}

type Module struct {
	Cfg    *config.Config
	AuthUC authuc.AuthUC
}

type userAuth struct{}

var (
	userContextKey = userAuth{}
)

func NewMiddlewareModule(module *Module) *Module {
	return module
}

func (m *Module) AuthHandler(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		bearerToken := r.Header.Get("Authorization")
		token, err := authusecase.GetBearerToken(bearerToken)
		if err != nil {
			handleError(ctx, w, r, err)
			return
		}

		userDetail, err := m.AuthUC.HandleAuthMiddleware(ctx, token)
		if err != nil && errors.Is(err, commonerr.SetNewTokenExpiredError()) {
			http.Redirect(w, r, "/token-expired", http.StatusFound)
			return
		} else if err != nil {
			handleError(ctx, w, r, err)
			return
		}

		ctx = SetUserDetailToCtx(ctx, userDetail)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}

func (m *Module) CorsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(AllowedMethodRequest, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(AllowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetUserDetailToCtx(ctx context.Context, data model.UserJWTPayload) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, userContextKey, data)
}

func GetUserDetailFromCtx(ctx context.Context) (model.UserJWTPayload, bool) {
	user, ok := ctx.Value(userContextKey).(model.UserJWTPayload)
	return user, ok
}

func handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
	commonwriter.SetError(ctx, w, err)
}
