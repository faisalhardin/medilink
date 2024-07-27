package http

import "net/http"

type PatientHandler interface {
	RegisterNewPatient(w http.ResponseWriter, r *http.Request)
}
