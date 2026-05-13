package http

import "net/http"

type AnamnesaHandler interface {
	GetByVisitID(w http.ResponseWriter, r *http.Request)
	GetDetailedByVisitID(w http.ResponseWriter, r *http.Request)
	Upsert(w http.ResponseWriter, r *http.Request)
}
