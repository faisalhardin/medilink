package http

import "net/http"

type InstitutionHandler interface {
	InsertNewInstitution(w http.ResponseWriter, r *http.Request)
	FindInstitutions(w http.ResponseWriter, r *http.Request)
	GetUserInstitution(w http.ResponseWriter, r *http.Request)

	FindInstitutionProducts(w http.ResponseWriter, r *http.Request)
	InsertInstitutionProduct(w http.ResponseWriter, r *http.Request)
	UpdateInstitutionProduct(w http.ResponseWriter, r *http.Request)
	UpdateInstitutionProductStock(w http.ResponseWriter, r *http.Request)
}
