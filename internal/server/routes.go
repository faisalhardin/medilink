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
			})
			authed.Route("/patient", func(institution chi.Router) {
				institution.Post("/", m.httpHandler.PatientHandler.RegisterNewPatient)
				institution.Get("/", m.httpHandler.PatientHandler.GetPatient)
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
