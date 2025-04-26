package http

import "net/http"

type PatientHandler interface {
	RegisterNewPatient(w http.ResponseWriter, r *http.Request)
	ListPatient(w http.ResponseWriter, r *http.Request)
	GetPatient(w http.ResponseWriter, r *http.Request)
	UpdatePatient(w http.ResponseWriter, r *http.Request)
	InsertNewVisit(w http.ResponseWriter, r *http.Request)
	ListVisitTouchpoints(w http.ResponseWriter, r *http.Request)
	GetPatientVisits(w http.ResponseWriter, r *http.Request)
	ListPatientVisitsByPatientUUID(w http.ResponseWriter, r *http.Request)
	ListPatientVisits(w http.ResponseWriter, r *http.Request)
	UpdatePatientVisit(w http.ResponseWriter, r *http.Request)

	GetVisitTouchpoint(w http.ResponseWriter, r *http.Request)
	UpsertVisitTouchpoint(w http.ResponseWriter, r *http.Request)
	InsertVisitProduct(w http.ResponseWriter, r *http.Request)
}
