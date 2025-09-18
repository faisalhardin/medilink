package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/faisalhardin/medilink/internal/entity/model"
	journeyRepo "github.com/faisalhardin/medilink/internal/entity/repo/journey"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	authRepo "github.com/faisalhardin/medilink/internal/repo/auth"
	"github.com/faisalhardin/medilink/internal/repo/cache"
	"github.com/faisalhardin/medilink/internal/repo/staff"
)

type AuthUC struct {
	Cfg         config.Config
	AuthRepo    authRepo.Options
	StaffRepo   staff.Conn
	JourneyRepo journeyRepo.JourneyDB
}

type userAuth struct{}

type AuthParams struct {
	Token     string          `json:"token,omitempty"`
	Email     string          `json:"email,omitempty"`
	TokenKey  string          `json:"auth_token_key,omitempty"`
	TokenPair model.TokenPair `json:"token_pair,omitempty"`
}

func New(u *AuthUC) *AuthUC {
	return u
}

func (u *AuthUC) Login(w http.ResponseWriter, r *http.Request, params AuthParams) (res AuthParams, err error) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")

	tokens, err := u.AuthRepo.GetGoogleAuthCallback(ctx, code)
	if err != nil {
		return
	}

	authedUser, err := u.AuthRepo.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return
	}

	userDetail, err := u.StaffRepo.GetUserDetailByEmail(ctx, authedUser.Email)
	if err != nil {
		return
	}

	journeyPoints, err := u.JourneyRepo.GetJourneyPointMappedByStaff(ctx, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return
	}

	servicePoints, err := u.JourneyRepo.GetServicePointMappedByJourneyPoints(ctx, journeyPoints, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return
	}

	// Extract device information
	deviceInfo := model.DeviceInfo{
		DeviceID:  r.Header.Get("X-Device-ID"),
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: getClientIP(r),
	}

	// Create token pair instead of single JWT
	tokenPair, err := u.AuthRepo.CreateTokenPair(ctx,
		model.GenerateUserDataJWTInformation(userDetail, authedUser, journeyPoints, servicePoints),
		deviceInfo)
	if err != nil {
		return
	}

	// Store session information (optional, for additional security)
	sessionPayloadInBytes, err := json.Marshal(model.GenerateUserDetailSessionInformation(userDetail, time.Now().Add(time.Duration(u.Cfg.JWTConfig.DurationInMinutes)*time.Minute)))
	if err != nil {
		return
	}

	_, err = u.AuthRepo.StoreLoginInformation(ctx, getSessionKey(authedUser.Email, tokenPair.AccessToken), string(sessionPayloadInBytes), time.Duration(u.Cfg.JWTConfig.DurationInMinutes)*time.Minute)
	if err != nil {
		return
	}

	return AuthParams{TokenPair: tokenPair}, nil
}

func (u *AuthUC) GetToken(ctx context.Context, params AuthParams) (tokenResponse AuthParams, err error) {
	tokenKey, err := u.AuthRepo.GetTokenFromKeyToken(ctx, params.TokenKey)
	if err != nil {
		return
	}

	return AuthParams{Token: tokenKey}, nil
}

func (u *AuthUC) GetUserDetail(ctx context.Context, params AuthParams) (userDetail model.UserSessionDetail, err error) {
	userDtlString, err := u.AuthRepo.GetLoginInformation(ctx, getSessionKey(params.Email, "test"))
	if err != nil {
		return
	}

	err = json.Unmarshal([]byte(userDtlString), &userDetail)
	if err != nil {
		return
	}

	return
}

func (u *AuthUC) HandleAuthMiddleware(ctx context.Context, token string) (ret model.UserJWTPayload, err error) {

	claims, err := u.GetTokenClaims(token)
	if err != nil {
		return
	}

	err = claims.Verify()
	if err != nil {
		return
	}

	userDetail := claims.Payload

	userSessionDetail, err := u.ValidateUserFromSession(ctx, &userDetail, token)
	if err != nil {
		return
	}

	consolidateUserAuthWithSession(&userDetail, userSessionDetail)

	return userDetail, nil
}

func (u *AuthUC) GetTokenClaims(token string) (claims *authRepo.Claims, err error) {

	claims = &authRepo.Claims{}

	err = u.AuthRepo.VerifyJWT(token, claims)
	if err != nil {
		return
	}

	return claims, nil
}

func (u *AuthUC) ValidateUserFromSession(ctx context.Context, jwtPayload *model.UserJWTPayload, token string) (sessionDetail model.UserSessionDetail, err error) {

	sessionInfo, err := u.AuthRepo.GetLoginInformation(ctx, getSessionKey(jwtPayload.Email, token))
	if err != nil && errors.Is(err, cache.ErrKeyNotFound) {
		err = commonerr.SetNewTokenExpiredError()
		return
	} else if err != nil {
		return
	}

	userSessionDetail := model.UserSessionDetail{}
	err = json.Unmarshal([]byte(sessionInfo), &userSessionDetail)
	if err != nil {
		return
	}

	return userSessionDetail, nil
}

func consolidateUserAuthWithSession(payload *model.UserJWTPayload, sessionDetail model.UserSessionDetail) {
	payload.InstitutionID = sessionDetail.IdMstInstitution
	payload.UserID = sessionDetail.UserID
	return
}

func (u *AuthUC) RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (tokenPair model.TokenPair, err error) {
	// Validate refresh token
	refreshToken, err := u.AuthRepo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return tokenPair, commonerr.SetNewUnauthorizedAPICall()
	}

	// Get user details
	userDetail, err := u.StaffRepo.GetUserDetailByID(ctx, refreshToken.UserID)
	if err != nil {
		return tokenPair, err
	}

	// Get journey points and service points
	journeyPoints, err := u.JourneyRepo.GetJourneyPointMappedByStaff(ctx, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return tokenPair, err
	}

	servicePoints, err := u.JourneyRepo.GetServicePointMappedByJourneyPoints(ctx, journeyPoints, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return tokenPair, err
	}

	// Generate new token pair
	newTokenPair, err := u.AuthRepo.CreateTokenPair(ctx,
		model.GenerateUserDataJWTInformation(userDetail, model.GoogleUser{}, journeyPoints, servicePoints),
		model.DeviceInfo{
			DeviceID:  req.DeviceID,
			UserAgent: req.UserAgent,
			IPAddress: req.IPAddress,
		})
	if err != nil {
		return tokenPair, err
	}

	// Revoke old refresh token
	err = u.AuthRepo.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return tokenPair, err
	}

	return newTokenPair, nil
}

func (u *AuthUC) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken != "" {
		return u.AuthRepo.RevokeRefreshToken(ctx, refreshToken)
	}
	return nil
}

func (u *AuthUC) LogoutAllDevices(ctx context.Context, userID int64) error {
	return u.AuthRepo.RevokeAllUserRefreshTokens(ctx, userID)
}

func getClientIP(r *http.Request) string {
	// Check for forwarded headers first
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fall back to remote address
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}

func getSessionKey(userIdentifier, token string) string {
	var subToken string
	splitToken := strings.Split(token, ".")
	if len(splitToken) == 3 && len(splitToken[2]) > 8 {
		lenSignature := len(splitToken[2])
		subToken = splitToken[2][lenSignature-8 : lenSignature]
	}

	return fmt.Sprintf("%s:%s", userIdentifier, subToken)
}
