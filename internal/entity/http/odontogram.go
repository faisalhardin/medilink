package http

import "net/http"

// OdontogramHandler defines the HTTP handler interface for odontogram operations
type OdontogramHandler interface {
	CreateEvents(w http.ResponseWriter, r *http.Request)
	GetEvents(w http.ResponseWriter, r *http.Request)
	GetSnapshot(w http.ResponseWriter, r *http.Request)
}
