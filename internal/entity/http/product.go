package http

import "net/http"

type ProductHandler interface {
	InsertMstProduct(w http.ResponseWriter, r *http.Request)
	ListMstProduct(w http.ResponseWriter, r *http.Request)
	UpdateMstProduct(w http.ResponseWriter, r *http.Request)
	DeleteMstProduct(w http.ResponseWriter, r *http.Request)
}
