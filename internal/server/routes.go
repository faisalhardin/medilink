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
	r.Use(m.middlewareModule.CorsHandler)
	r.Route("/v1", func(v1 chi.Router) {
		v1.Group(func(authed chi.Router) {

			authed.Use(m.middlewareModule.AuthHandler)
			authed.Get("/logout/{provider}", m.httpHandler.AuthHandler.Logout)
			authed.Route("/institution", func(institution chi.Router) {
				institution.Post("/", m.httpHandler.InstitutionHandler.InsertNewInstitution)
				institution.Get("/", m.httpHandler.InstitutionHandler.GetUserInstitution)
				institution.Route("/product", func(product chi.Router) {
					product.Get("/", m.httpHandler.InstitutionHandler.FindInstitutionProducts)
					product.Post("/", m.httpHandler.InstitutionHandler.InsertInstitutionProduct)
					product.Patch("/", m.httpHandler.InstitutionHandler.UpdateInstitutionProduct)
					product.Post("/resupply", m.httpHandler.InstitutionHandler.UpdateInstitutionProductStock)
				})
			})
			authed.Route("/patient", func(patient chi.Router) {
				patient.Post("/", m.httpHandler.PatientHandler.RegisterNewPatient)
				patient.Patch("/", m.httpHandler.PatientHandler.UpdatePatient)
				patient.Get("/", m.httpHandler.PatientHandler.ListPatient)
				patient.Route("/{uuid}", func(patient chi.Router) {
					patient.Get("/", m.httpHandler.PatientHandler.GetPatient)
					patient.Get("/visit", m.httpHandler.PatientHandler.ListPatientVisitsByPatientUUID)
				})
			})

			// START: /v1/visit
			authed.Route("/visit", func(visit chi.Router) {
				visit.Post("/", m.httpHandler.PatientHandler.InsertNewVisit)
				visit.Get("/", m.httpHandler.PatientHandler.ListPatientVisits)
				visit.Get("/detailed", m.httpHandler.PatientHandler.ListPatientVisitsDetailed)
				visit.Patch("/archive", m.httpHandler.PatientHandler.ArchivePatientVisit)
				visit.Route("/{id}", func(visit chi.Router) {
					visit.Patch("/", m.httpHandler.PatientHandler.UpdatePatientVisit)
					visit.Get("/", m.httpHandler.PatientHandler.GetPatientVisits)
					visit.Get("/detail", m.httpHandler.PatientHandler.ListVisitTouchpoints)
				})
				visit.Get("/product", m.httpHandler.PatientHandler.ListVisitProducts)
				visit.Post("/product", m.httpHandler.PatientHandler.InsertVisitProduct)
			})
			// END: /v1/visit

			authed.Route("/visit-detail", func(visit chi.Router) {
				visit.Post("/", m.httpHandler.PatientHandler.UpsertVisitTouchpoint)
				visit.Route("/{id}", func(visit chi.Router) {
					visit.Get("/", m.httpHandler.PatientHandler.GetVisitTouchpoint)
				})
			})

			authed.Route("/journey", func(journey chi.Router) {
				journey.Route("/board", func(board chi.Router) {
					board.Get("/{id}", m.httpHandler.JourneyHandler.GetJourneyBoard)
					board.Get("/", m.httpHandler.JourneyHandler.ListJourneyBoard)
					board.Post("/", m.httpHandler.JourneyHandler.InsertNewJourneyBoard)
					board.Patch("/", m.httpHandler.JourneyHandler.UpdateJourneyBoard)
					board.Delete("/", m.httpHandler.JourneyHandler.DeleteJourneyBoard)
				})

				journey.Route("/point", func(board chi.Router) {
					board.Post("/", m.httpHandler.JourneyHandler.InsertNewJourneyPoint)
					board.Patch("/{id}", m.httpHandler.JourneyHandler.UpdateJourneyPoint)
					board.Patch("/rename", m.httpHandler.JourneyHandler.RenameJourneyPoint)
					board.Patch("/archive", m.httpHandler.JourneyHandler.ArchiveJourneyPoint)
				})

				journey.Route("/service-point", func(servicePoint chi.Router) {
					servicePoint.Get("/", m.httpHandler.JourneyHandler.ListServicePoint)
					servicePoint.Get("/{id}", m.httpHandler.JourneyHandler.GetServicePoint)
					servicePoint.Post("/", m.httpHandler.JourneyHandler.InsertNewServicePoint)
					servicePoint.Patch("/{id}", m.httpHandler.JourneyHandler.UpdateServicePoint)
					servicePoint.Delete("/{id}", m.httpHandler.JourneyHandler.ArchiveServicePoint)
				})
			})
		})

		v1.Route("/auth", func(auth chi.Router) {

			auth.Get("/{provider}/callback", m.httpHandler.AuthHandler.GetAuthCallbackFunction)
			auth.Get("/{provider}", m.httpHandler.AuthHandler.BeginAuthProviderCallback)
			auth.Get("/jwt", m.httpHandler.AuthHandler.GetTokenFromTokenKey)
			auth.Post("/pseudologin", m.httpHandler.AuthHandler.PseudoLogin)

			// New authentication endpoints
			auth.Post("/refresh", m.httpHandler.AuthHandler.RefreshToken)
			auth.Post("/logout", m.httpHandler.AuthHandler.LogoutSession)
			auth.Post("/logout-all", m.httpHandler.AuthHandler.LogoutAllSessions)
			auth.Get("/sessions", m.httpHandler.AuthHandler.GetUserSessions)

			auth.Group(func(authenticate chi.Router) {
				authenticate.Post("/get-login", m.httpHandler.AuthHandler.GetLoginByToken)
				authenticate.Get("/verify", m.httpHandler.AuthHandler.GetUserFromToken)

			})
		})

		v1.Route("/admin", func(auth chi.Router) {
			auth.Use(m.middlewareModule.AuthHandler)
			auth.Route("/product", func(product chi.Router) {
				product.Get("/", m.httpHandler.ProductHandler.ListMstProduct)
				product.Post("/", m.httpHandler.ProductHandler.InsertMstProduct)
				product.Patch("/", m.httpHandler.ProductHandler.UpdateMstProduct)
				product.Delete("/", m.httpHandler.ProductHandler.DeleteMstProduct)
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
