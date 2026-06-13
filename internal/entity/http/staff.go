package http

import "net/http"

type StaffHandler interface {
	ListStaff(w http.ResponseWriter, r *http.Request)
	GetStaff(w http.ResponseWriter, r *http.Request)
	CreateStaff(w http.ResponseWriter, r *http.Request)
	AssignRole(w http.ResponseWriter, r *http.Request)
	UnassignRole(w http.ResponseWriter, r *http.Request)
	DeactivateStaff(w http.ResponseWriter, r *http.Request)
	ActivateStaff(w http.ResponseWriter, r *http.Request)
}
