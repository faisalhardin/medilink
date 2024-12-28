package http

import "net/http"

type JourneyHandler interface {
	InsertNewJourneyBoard(w http.ResponseWriter, r *http.Request)
	ListJourneyBoard(w http.ResponseWriter, r *http.Request)
	GetJourneyBoard(w http.ResponseWriter, r *http.Request)
	UpdateJourneyBoard(w http.ResponseWriter, r *http.Request)
	DeleteJourneyBoard(w http.ResponseWriter, r *http.Request)

	InsertNewJourneyPoint(w http.ResponseWriter, r *http.Request)
	UpdateJourneyPoint(w http.ResponseWriter, r *http.Request)
	ArchiveJourneyPoint(w http.ResponseWriter, r *http.Request)
}
