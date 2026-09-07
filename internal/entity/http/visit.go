package http

import "net/http"

type VisitContributorHandler interface {
	ListVisitContributors(w http.ResponseWriter, r *http.Request)
}
