package server

import (
	"context"
	"encoding/json"
	"fmt"
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
		// v1.Group(func(authed chi.Router) {
		// 	authed.Get("/auth/{provider}/callback", handler..GetAuthCallbackFunction)
		// 	authed.Get("/auth/{provider}", handler.BeginAuthProviderCallback)
		// 	authed.Get("/logout/{provider}", handler.Logout)
		// 	authed.Post("/register/user", handler.TestBinding)
		// 	authed.Get("/user", handler.TestSchemaBinding)
		// })

		v1.Group(func(authed chi.Router) {
			authed.Post("/institution", m.httpHandler.InstitutionHandler.InsertNewInstitution)
			authed.Get("/institution", m.httpHandler.InstitutionHandler.FindInstitutions)
			authed.Post("/patient", m.httpHandler.PatientHandler.RegisterNewPatient)
		})

	})

	r.Route("/auth", func(auth chi.Router) {
		auth.Group(func(authenticate chi.Router) {
			authenticate.Post("/pseudologin", m.httpHandler.AuthHandler.PseudoLogin)
			authenticate.Post("/get-login", m.httpHandler.AuthHandler.GetLoginByToken)

		})
	})

	// r.Get("/ping", handler.PingAPI)
	// r.Get("/redirect", handler.TestAPIRedirect)

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

func (s *Server) getAuthCallbackFunction(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Fprintln(w, r)
		return
	}

	fmt.Println(user)
	fmt.Println("here 3")
	http.Redirect(w, r, "http://localhost:5173", http.StatusFound)
	fmt.Println("here 4")
}

func (s *Server) beginAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	// try to get the user without re-authenticating
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
	fmt.Println("here")
	gothic.BeginAuthHandler(w, r)
	fmt.Println("here 2")
}

func (s *Server) logout(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}
