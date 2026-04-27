package http

import "net/http"

// ICD10Handler is the HTTP handler interface for ICD-10 reference lookups.
type ICD10Handler interface {
	Search(w http.ResponseWriter, r *http.Request)
}
