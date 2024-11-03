package auth

import (
	"context"
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	authmodel "github.com/faisalhardin/medilink/internal/usecase/auth"
)

type AuthUC interface {
	Login(w http.ResponseWriter, r *http.Request, params authmodel.AuthParams) (res authmodel.AuthParams, err error)
	GetUserDetail(ctx context.Context, params authmodel.AuthParams) (userDetail model.UserSessionDetail, err error)
	HandleAuthMiddleware(ctx context.Context, token string) (ret model.UserJWTPayload, err error)
	GetToken(ctx context.Context, params authmodel.AuthParams) (tokenResponse authmodel.AuthParams, err error)
}
