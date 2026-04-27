package http

import "net/http"

type DiagnosisHandler interface {
	GetByVisitID(w http.ResponseWriter, r *http.Request)
	Save(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}
