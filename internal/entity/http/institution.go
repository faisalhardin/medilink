package http

import "net/http"

type InstitutionHandler interface {
	InsertNewInstitution(w http.ResponseWriter, r *http.Request)
	FindInstitutions(w http.ResponseWriter, r *http.Request)
}
