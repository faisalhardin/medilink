package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
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
	SessionRepo *authRepo.SessionRepository
	StaffRepo   staff.Conn
	JourneyRepo journeyRepo.JourneyDB

	refreshTokenDurationInHours int
	tokenDurationInMinutes      int
}

type userAuth struct{}

type AuthParams struct {
	Token    string `json:"token,omitempty"`
	Email    string `json:"email,omitempty"`
	TokenKey string `json:"auth_token_key,omitempty"`
}

func New(u *AuthUC) *AuthUC {
	u.refreshTokenDurationInHours = u.Cfg.JWTConfig.RefreshTokenDurationInHours
	if u.refreshTokenDurationInHours == 0 {
		u.refreshTokenDurationInHours = 24
	}
	u.tokenDurationInMinutes = u.Cfg.JWTConfig.DurationInMinutes
	if u.tokenDurationInMinutes == 0 {
		u.tokenDurationInMinutes = 15
	}
	return u
}

// Login handles the OAuth login flow and creates a new session
func (u *AuthUC) Login(w http.ResponseWriter, r *http.Request, params AuthParams) (res model.LoginResponse, err error) {
	ctx := r.Context()
	code := r.URL.Query().Get("code")

	// 1. Authenticate with Google OAuth
	tokens, err := u.AuthRepo.GetGoogleAuthCallback(ctx, code)
	if err != nil {
		return
	}

	authedUser, err := u.AuthRepo.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return
	}

	// 2. Get user details from database
	userDetail, err := u.StaffRepo.GetUserDetailByEmail(ctx, authedUser.Email)
	if err != nil {
		return
	}

	// 3. Get journey points and service points
	journeyPoints, err := u.JourneyRepo.GetJourneyPointMappedByStaff(ctx, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return
	}

	servicePoints, err := u.JourneyRepo.GetServicePointMappedByJourneyPoints(ctx, journeyPoints, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return
	}

	// 4. Generate tokens
	currTime := time.Now()
	accessTokenExpiry := currTime.Add(time.Duration(u.tokenDurationInMinutes) * time.Minute)
	refreshTokenExpiry := currTime.Add(time.Duration(u.refreshTokenDurationInHours) * time.Hour)

	// Create JWT payload
	jwtPayload := model.GenerateUserDataJWTInformation(userDetail, authedUser, journeyPoints, servicePoints)

	sessionID, err := authRepo.GenerateSessionID()
	if err != nil {
		return
	}

	accessToken, err := u.AuthRepo.CreateJWTToken(ctx, sessionID, jwtPayload, currTime, accessTokenExpiry)
	if err != nil {
		return
	}

	refreshToken, err := authRepo.GenerateRefreshToken()
	if err != nil {
		return
	}

	// 5. Create session in database
	session := &model.UserSession{
		SessionKey:       sessionID,
		UserID:           userDetail.Staff.ID,
		AccessTokenHash:  authRepo.HashToken(accessToken),
		RefreshTokenHash: authRepo.HashToken(refreshToken),
		Status:           string(model.SessionStatusActive),
		ExpiresAt:        accessTokenExpiry,
		RefreshExpiresAt: refreshTokenExpiry,
		LastAccessedAt:   currTime,
		IPAddress:        u.getClientIP(r),
		UserAgent:        r.UserAgent(),
	}

	err = u.SessionRepo.CreateSession(ctx, session)
	if err != nil {
		return
	}

	return model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(u.tokenDurationInMinutes * 60),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken handles token refresh using refresh token
func (u *AuthUC) RefreshToken(ctx context.Context, req model.RefreshTokenRequest) (res model.RefreshTokenResponse, err error) {
	// 1. Validate refresh token
	session, err := u.SessionRepo.GetSessionByRefreshToken(ctx, authRepo.HashToken(req.RefreshToken))
	if err != nil {
		return
	}

	// 2. Check if session is active and refresh token not expired
	if session.Status != string(model.SessionStatusActive) {
		err = commonerr.SetNewRevokedSessionError()
		return
	}

	if time.Now().After(session.RefreshExpiresAt) {
		// Mark session as expired
		u.SessionRepo.UpdateSessionStatus(ctx, session.SessionKey, string(model.SessionStatusExpired))
		err = commonerr.SetNewRevokedSessionError()
		return
	}

	// 3. Generate new access token
	currTime := time.Now()
	newAccessTokenExpiry := currTime.Add(time.Duration(u.tokenDurationInMinutes) * time.Minute)

	userDetail, err := u.StaffRepo.GetUserDetailByID(ctx, session.UserID)
	if err != nil {
		return
	}

	// Get journey points and service points for the user
	journeyPoints, err := u.JourneyRepo.GetJourneyPointMappedByStaff(ctx, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return
	}

	servicePoints, err := u.JourneyRepo.GetServicePointMappedByJourneyPoints(ctx, journeyPoints, model.MstStaff{ID: userDetail.Staff.ID})
	if err != nil {
		return
	}

	// Create new JWT payload
	jwtPayload := model.GenerateUserDataJWTInformation(userDetail, model.GoogleUser{}, journeyPoints, servicePoints)
	newAccessToken, err := u.AuthRepo.CreateJWTToken(ctx, session.SessionKey, jwtPayload, currTime, newAccessTokenExpiry)
	if err != nil {
		return
	}

	// 4. Update session with new access token
	// err = u.SessionRepo.UpdateSessionAccessToken(ctx, session.SessionKey, authRepo.HashToken(newAccessToken), newAccessTokenExpiry)
	err = u.SessionRepo.UpdateSessionAccessToken(ctx, &model.UserSession{
		SessionKey:      session.SessionKey,
		AccessTokenHash: authRepo.HashToken(newAccessToken),
		ExpiresAt:       newAccessTokenExpiry,
	})
	if err != nil {
		return
	}

	return model.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: req.RefreshToken,
		ExpiresIn:    int64(u.tokenDurationInMinutes * 60),
		TokenType:    "Bearer",
	}, nil
}

// Logout handles user logout
func (u *AuthUC) Logout(ctx context.Context, r *http.Request) error {
	token := r.Header.Get("Authorization")
	token, err := GetBearerToken(token)
	if err != nil {
		return err
	}
	claims, err := u.GetTokenClaims(token)
	if err != nil {
		return err
	}

	sessionKey := claims.ID
	// Mark session as revoked
	return u.SessionRepo.UpdateSessionStatus(ctx, sessionKey, string(model.SessionStatusRevoked))
}

// LogoutAllSessions revokes all sessions for a user
func (u *AuthUC) LogoutAllSessions(ctx context.Context, userID int64) error {
	return u.SessionRepo.RevokeAllUserSessions(ctx, userID)
}

// GetUserSessions retrieves all sessions for a user
func (u *AuthUC) GetUserSessions(ctx context.Context, userID int64) ([]model.SessionInfo, error) {
	return u.SessionRepo.GetUserSessions(ctx, userID)
}

// GetToken handles legacy token retrieval (keeping for backward compatibility)
func (u *AuthUC) GetToken(ctx context.Context, params AuthParams) (tokenResponse AuthParams, err error) {
	tokenKey, err := u.AuthRepo.GetTokenFromKeyToken(ctx, params.TokenKey)
	if err != nil {
		return
	}

	return AuthParams{Token: tokenKey}, nil
}

// GetUserDetail handles legacy user detail retrieval (keeping for backward compatibility)
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

// HandleAuthMiddleware validates the session and returns user information
func (u *AuthUC) HandleAuthMiddleware(ctx context.Context, token string) (ret model.UserJWTPayload, err error) {
	// 1. Verify JWT token
	claims, err := u.GetTokenClaims(token)
	if err != nil {
		return
	}

	err = claims.Verify()
	if err != nil {
		return
	}

	userDetail := claims.Payload

	// 2. Validate session from database
	session, err := u.ValidateSession(ctx, token, claims.ID)
	if err != nil {
		return
	}

	// 3. Get additional session details and consolidate
	userSessionDetail, err := u.getUserSessionDetail(ctx, session.UserID)
	if err != nil {
		return
	}

	consolidateUserAuthWithSession(&userDetail, userSessionDetail)

	return userDetail, nil
}

// ValidateSession validates a session from the database
func (u *AuthUC) ValidateSession(ctx context.Context, accessToken, sessionKey string) (*model.UserSession, error) {
	// 1. Get session from database
	session, err := u.SessionRepo.GetSessionByKey(ctx, sessionKey)
	if err != nil {
		return nil, commonerr.SetNewUnauthorizedError("session not found", "Your session could not be found")
	}

	// 2. Check session status
	if session.Status != string(model.SessionStatusActive) {
		return nil, commonerr.SetNewUnauthorizedError("session revoked", "Your session has been revoked")
	}

	// 3. Check if access token is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, commonerr.SetNewTokenExpiredError()
	}

	// 4. Update last accessed time
	u.SessionRepo.UpdateLastAccessed(ctx, sessionKey)

	return session, nil
}

// getUserSessionDetail gets user session details from the database
func (u *AuthUC) getUserSessionDetail(ctx context.Context, userID int64) (model.UserSessionDetail, error) {
	userDetail, err := u.StaffRepo.GetUserDetailByID(ctx, userID)
	if err != nil {
		return model.UserSessionDetail{}, err
	}

	return model.UserSessionDetail{
		UserID:           userDetail.Staff.ID,
		Name:             userDetail.Staff.Name,
		IdMstInstitution: userDetail.Staff.IdMstInstitution,
		ExpiredAt:        time.Now().Add(time.Duration(u.tokenDurationInMinutes) * time.Minute).Unix(),
	}, nil
}

func (u *AuthUC) GetTokenClaims(token string) (claims *authRepo.Claims, err error) {
	claims = &authRepo.Claims{}

	err = u.AuthRepo.VerifyJWT(token, claims)
	if err != nil {
		return
	}

	return claims, nil
}

// Legacy method - keeping for backward compatibility
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

// getClientIP extracts the client IP address from the request
func (u *AuthUC) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// Legacy function - keeping for backward compatibility
func getSessionKey(userIdentifier, token string) string {
	var subToken string
	splitToken := strings.Split(token, ".")
	if len(splitToken) == 3 && len(splitToken[2]) > 8 {
		lenSignature := len(splitToken[2])
		subToken = splitToken[2][lenSignature-8 : lenSignature]
	}

	return fmt.Sprintf("%s:%s", userIdentifier, subToken)
}

func GetBearerToken(token string) (string, error) {
	splitToken := strings.Split(token, "Bearer ")
	if len(splitToken) != 2 {
		return "", errors.New("invalid token")
	}

	return splitToken[1], nil
}
