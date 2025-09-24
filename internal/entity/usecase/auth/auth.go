package auth

import (
	"context"
	"net/http"

	"github.com/faisalhardin/medilink/internal/entity/model"
	authmodel "github.com/faisalhardin/medilink/internal/usecase/auth"
)

type AuthUC interface {
	Login(w http.ResponseWriter, r *http.Request, params authmodel.AuthParams) (res model.LoginResponse, err error)
	RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (res model.RefreshTokenResponse, err error)
	Logout(ctx context.Context, r *http.Request) error
	LogoutAllSessions(ctx context.Context, userID int64) error
	GetUserSessions(ctx context.Context, userID int64) ([]model.SessionInfo, error)
	GetUserDetail(ctx context.Context, params authmodel.AuthParams) (userDetail model.UserSessionDetail, err error)
	HandleAuthMiddleware(ctx context.Context, token string) (ret model.UserJWTPayload, err error)
	GetToken(ctx context.Context, params authmodel.AuthParams) (tokenResponse authmodel.AuthParams, err error)
}
