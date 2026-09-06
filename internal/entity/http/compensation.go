package http

import "net/http"

type CompensationPeriodHandler interface {
	CreatePeriod(w http.ResponseWriter, r *http.Request)
	ListPeriods(w http.ResponseWriter, r *http.Request)
	GetPeriod(w http.ResponseWriter, r *http.Request)
	DraftPeriod(w http.ResponseWriter, r *http.Request)
	FinalizePeriod(w http.ResponseWriter, r *http.Request)
	ReopenPeriod(w http.ResponseWriter, r *http.Request)
	DeletePeriod(w http.ResponseWriter, r *http.Request)
}
