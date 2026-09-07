package http

import "net/http"

type VisitContributorHandler interface {
	ListVisitContributors(w http.ResponseWriter, r *http.Request)
	AddVisitContributor(w http.ResponseWriter, r *http.Request)
	DeleteVisitContributor(w http.ResponseWriter, r *http.Request)
}
