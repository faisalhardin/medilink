package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	authRepo "github.com/faisalhardin/medilink/internal/repo/auth"
	"github.com/faisalhardin/medilink/internal/repo/staff"
)

type AuthUC struct {
	Cfg       config.Config
	AuthRepo  authRepo.Options
	StaffRepo staff.Conn
}

type AuthParams struct {
	Token string `json:"token,omitempty"`
	Email string `json:"email,omitempty"`
}

func New(u *AuthUC) *AuthUC {
	return u
}

func (u *AuthUC) Login(w http.ResponseWriter, r *http.Request, params AuthParams) (res AuthParams, err error) {
	ctx := r.Context()

	resp, err := u.AuthRepo.Str.Get(ctx, params.Token)
	if err != nil && !errors.Is(err, constant.ErrorNotFound) {
		return
	}

	if len(resp) > 0 {
		commonwriter.Redirect(ctx, w, r, u.Cfg.GoogleAuthConfig.HomepageRedirect, http.StatusFound)
	}

	userDetail, err := u.StaffRepo.GetUserDetailByEmail(ctx, params.Email)
	if err != nil {
		return
	}

	token, err := u.AuthRepo.CreateJWTToken(ctx, userDetail, u.Cfg.JWTConfig.DurationInMinutes)
	if err != nil {
		return
	}

	return AuthParams{Token: token}, nil
}

func (u *AuthUC) GetUserDetail(ctx context.Context, params AuthParams) (userDetail model.UserDetail, err error) {
	userDtlString, err := u.AuthRepo.ValidateToken(ctx, params.Token)
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(userDtlString), &userDetail)
	if err != nil {
		return
	}

	return
}
