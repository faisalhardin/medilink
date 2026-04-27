package http

import "net/http"

// PractitionerHandler is the HTTP handler interface for doctor / nurse autocomplete lookups.
type PractitionerHandler interface {
	SearchDoctors(w http.ResponseWriter, r *http.Request)
	SearchNurses(w http.ResponseWriter, r *http.Request)
}
