package auth

import "net/http"

type Auth interface {
	GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request)
	BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request)
	Logout(res http.ResponseWriter, req *http.Request)
}
