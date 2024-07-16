package user

import "time"

type UserAuth struct {
	Provider          string    `json:"provider"`
	Email             string    `json:"email"`
	Name              string    `json:"name"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	NickName          string    `json:"nick_name"`
	UserID            string    `json:"user_id"`
	Location          string    `json:"location"`
	AccessToken       string    `json:"access_token"`
	AccessTokenSecret string    `json:"access_token_secret"`
	RefreshToken      string    `json:"refresh_token"`
	ExpiresAt         time.Time `json:"expires_at"`
	IDToken           string    `json:"id_token"`
}
