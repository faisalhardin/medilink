package http

import (
	"net/http"
)

type AuthHandler interface {
	GetAuthCallbackFunction(w http.ResponseWriter, r *http.Request)
	BeginAuthProviderCallback(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	PseudoLogin(w http.ResponseWriter, r *http.Request)
	GetLoginByToken(w http.ResponseWriter, r *http.Request)
	GetUserFromToken(w http.ResponseWriter, r *http.Request)
	PingAPI(w http.ResponseWriter, r *http.Request)
}
