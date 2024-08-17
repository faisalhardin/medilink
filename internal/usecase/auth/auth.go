package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
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

	userDetail, err := u.StaffRepo.GetUserDetailByEmail(ctx, params.Email)
	if err != nil {
		return
	}

	currTime := time.Now()
	expireDuration := time.Duration(u.Cfg.JWTConfig.DurationInMinutes) * time.Minute
	expiredTime := currTime.Add(expireDuration)
	token, err := u.AuthRepo.CreateJWTToken(ctx, model.GenerateUserDataJWTInformation(userDetail), currTime, expiredTime)
	if err != nil {
		return
	}

	sessionPayloadInBytes, err := json.Marshal(model.GenerateUserDetailSessionInformation(userDetail, expiredTime))
	if err != nil {
		return
	}

	_, err = u.AuthRepo.StoreLoginInformation(ctx, getExistingSessionByEmailKey(params.Email), string(sessionPayloadInBytes), expireDuration)
	if err != nil {
		return
	}

	return AuthParams{Token: token}, nil
}

func (u *AuthUC) GetUserDetail(ctx context.Context, params AuthParams) (userDetail model.UserSessionDetail, err error) {
	userDtlString, err := u.AuthRepo.GetLoginInformation(ctx, getExistingSessionByEmailKey(params.Email))
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(userDtlString), &userDetail)
	if err != nil {
		return
	}

	return
}

func getExistingSessionByEmailKey(email string) string {
	return fmt.Sprintf("session-email:%s", email)
}
