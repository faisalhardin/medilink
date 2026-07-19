package http

import "net/http"

type ProcedureHandler interface {
	SearchICD9CM(w http.ResponseWriter, r *http.Request)
	GetByVisitID(w http.ResponseWriter, r *http.Request)
	Save(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	GetPatientHistory(w http.ResponseWriter, r *http.Request)
}
