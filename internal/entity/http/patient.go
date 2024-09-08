package http

import "net/http"

type PatientHandler interface {
	RegisterNewPatient(w http.ResponseWriter, r *http.Request)
	GetPatient(w http.ResponseWriter, r *http.Request)
	UpdatePatient(w http.ResponseWriter, r *http.Request)
}
