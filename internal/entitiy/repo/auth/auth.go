package auth

import (
	"net/http"

	"github.com/markbates/goth"
)

type AuthRepo interface {
	GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request) (goth.User, error)
	BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request)
	Logout(res http.ResponseWriter, req *http.Request)
}
