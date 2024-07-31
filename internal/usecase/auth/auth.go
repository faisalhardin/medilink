package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/constant"
	"github.com/faisalhardin/medilink/internal/entity/model"
	commonwriter "github.com/faisalhardin/medilink/internal/library/common/writer"
	authRepo "github.com/faisalhardin/medilink/internal/repo/auth"
)

var (
	MockUser = map[string]model.UserDetail{
		"email1@gmail.com": model.UserDetail{
			Staff: model.MstStaff{
				ID:               1,
				UUID:             "1245",
				Name:             "Name 1",
				IdMstInstitution: 1,
			},
			Roles: []model.MstRole{
				model.MstRole{
					RoleID: 1,
					Name:   "Admin",
				},
			},
		},
		"email2@gmail.com": model.UserDetail{
			Staff: model.MstStaff{
				ID:               1,
				UUID:             "13213",
				Name:             "Name 2",
				IdMstInstitution: 1,
			},
			Roles: []model.MstRole{
				model.MstRole{
					RoleID: 2,
					Name:   "Doctor",
				},
			},
		},
	}
)

type AuthUC struct {
	Cfg      config.Config
	AuthRepo authRepo.Options
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

	userDetail, err := MockGetUserDB(ctx, params.Email)
	if err != nil {
		return
	}

	userDetailByte, _ := json.Marshal(userDetail)

	eightHourExpirationDuration := 8 * time.Hour
	token, err := u.AuthRepo.CreateToken(ctx, string(userDetailByte), eightHourExpirationDuration)
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

func MockGetUserDB(ctx context.Context, email string) (model.UserDetail, error) {
	if userDetail, found := MockUser[email]; found {
		return userDetail, nil
	}
	return model.UserDetail{}, constant.ErrorNotFound
}
