package auth

import "net/http"

type AuthHandler interface {
	GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request)
	BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	PingAPI(w http.ResponseWriter, r *http.Request)
	TestAPIRedirect(w http.ResponseWriter, r *http.Request)
}
