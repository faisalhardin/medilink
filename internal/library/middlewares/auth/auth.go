package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
	authuc "github.com/faisalhardin/medilink/internal/entity/usecase/auth"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
)

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
		token, err := GetBearerToken(bearerToken)
		if err != nil {
			handleError(ctx, w, r, err)
			return
		}

		userDetail, err := m.AuthUC.HandleAuthMiddleware(ctx, token)
		if err != nil {
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
		w.Header().Set("Access-Control-Allow-Origin", m.Cfg.WebConfig.Host)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	})
}

func GetBearerToken(token string) (string, error) {
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		return "", errors.New("invalid token")
	}

	return splitToken[1], nil
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
