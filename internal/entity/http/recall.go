package http

import "net/http"

// RecallHandler defines the HTTP handler interface for recall (scheduled control/appointment) operations
type RecallHandler interface {
	CreateRecall(w http.ResponseWriter, r *http.Request)
	UpdateRecall(w http.ResponseWriter, r *http.Request)
	GetNextRecallByPatient(w http.ResponseWriter, r *http.Request)
	ListRecalls(w http.ResponseWriter, r *http.Request)
}
