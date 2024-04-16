package auth

import "net/http"

type AuthHandler interface {
	GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request)
	BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}
