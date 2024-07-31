package http

import (
	"net/http"
)

type AuthHandler interface {
	PseudoLogin(w http.ResponseWriter, r *http.Request)
	GetLoginByToken(w http.ResponseWriter, r *http.Request)
}
