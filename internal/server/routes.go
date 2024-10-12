package server

import (
	"encoding/json"
	"log"
	"net/http"

	utilhandler "github.com/faisalhardin/medilink/internal/library/util/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/markbates/goth/gothic"
)

func RegisterRoutes(m *module) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(utilhandler.Handler)
	r.Route("/v1", func(v1 chi.Router) {
		v1.Group(func(authed chi.Router) {

			authed.Use(m.authModule.Handler)
			authed.Get("/logout/{provider}", m.httpHandler.AuthHandler.Logout)
			authed.Route("/institution", func(institution chi.Router) {
				institution.Post("/", m.httpHandler.InstitutionHandler.InsertNewInstitution)
				institution.Get("/", m.httpHandler.InstitutionHandler.FindInstitutions)
				institution.Route("/product", func(product chi.Router) {
					product.Get("/", m.httpHandler.InstitutionHandler.FindInstitutionProducts)
					product.Post("/", m.httpHandler.InstitutionHandler.InsertInstitutionProduct)
					product.Patch("/{id}/stock", m.httpHandler.InstitutionHandler.UpdateInstitutionProductStock)
					product.Patch("/{id}", m.httpHandler.InstitutionHandler.UpdateInstitutionProduct)
				})
			})
			authed.Route("/patient", func(patient chi.Router) {
				patient.Post("/", m.httpHandler.PatientHandler.RegisterNewPatient)
				patient.Get("/", m.httpHandler.PatientHandler.GetPatient)
				patient.Patch("/", m.httpHandler.PatientHandler.UpdatePatient)
				patient.Route("/{id}/visit", func(visit chi.Router) {
					visit.Post("/", m.httpHandler.PatientHandler.InsertNewVisit)
					visit.Get("/", m.httpHandler.PatientHandler.GetPatientVisits)
					visit.Patch("/", m.httpHandler.PatientHandler.UpdatePatientVisit)
				})
			})

			authed.Route("/visit/{id}", func(visit chi.Router) {
				visit.Get("/", m.httpHandler.PatientHandler.GetVisitTouchpoint)
				visit.Post("/", m.httpHandler.PatientHandler.InsertVisitTouchpoint)
				visit.Patch("/", m.httpHandler.PatientHandler.UpdateVisitTouchpoint)
			})
		})

		v1.Route("/auth", func(auth chi.Router) {

			auth.Get("/{provider}/callback", m.httpHandler.AuthHandler.GetAuthCallbackFunction)
			auth.Get("/{provider}", m.httpHandler.AuthHandler.BeginAuthProviderCallback)
			auth.Post("/pseudologin", m.httpHandler.AuthHandler.PseudoLogin)
			auth.Group(func(authenticate chi.Router) {
				authenticate.Post("/get-login", m.httpHandler.AuthHandler.GetLoginByToken)
				authenticate.Get("/verify", m.httpHandler.AuthHandler.GetUserFromToken)

			})
		})
	})
	r.Get("/ping", m.httpHandler.AuthHandler.PingAPI)

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// jsonResp, _ := json.Marshal(s.db.Health())
	// _, _ = w.Write(jsonResp)
}

func (s *Server) logout(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
